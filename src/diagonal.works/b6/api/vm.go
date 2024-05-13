package api

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"time"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
)

type Options struct {
	Cores         int
	FileIOAllowed bool
}

type Context struct {
	World           b6.World
	Worlds          ingest.Worlds
	Cores           int
	FileIOAllowed   bool
	Clock           func() time.Time
	Values          map[interface{}]interface{}
	FunctionSymbols FunctionSymbols
	Adaptors        Adaptors
	Context         context.Context

	VM *VM
}

func (c *Context) FillFromOptions(options *Options) {
	c.Cores = options.Cores
	c.FileIOAllowed = options.FileIOAllowed
}

func (c *Context) Fork(n int) []*Context {
	var vms []VM
	if c.VM != nil {
		vms = c.VM.Fork(n)
	}
	ctxs := make([]*Context, n)
	for i := range ctxs {
		copy := *c
		ctxs[i] = &copy
		if vms != nil {
			ctxs[i].VM = &vms[i]
		}
	}
	return ctxs
}

type Op uint8

const (
	OpPushValue Op = iota
	OpStore
	OpDiscard
	OpLoad
	OpJump
	OpCallValue
	OpCallStack
	OpReturn
)

func (o Op) String() string {
	switch o {
	case OpPushValue:
		return "PushValue"
	case OpStore:
		return "Store"
	case OpDiscard:
		return "Discard"
	case OpLoad:
		return "Load"
	case OpJump:
		return "Jump"
	case OpCallValue:
		return "CallValue"
	case OpCallStack:
		return "CallStack"
	case OpReturn:
		return "Return"
	}
	return "Bad"
}

type Callable interface {
	// The number of arguments expected by the function, excluding Context. If
	// the last argument is varadic, it's counted as a single argument with a
	// list type (matching go's semantics)
	NumArgs() int
	// Call a function using arguments on the VM's stack. Assumes that n
	// arguments have been pushed onto the stack in left-to-right order. These
	// arguments are removed from the stack, leaving the function's
	// result at the top of the stack.
	// Scratch is a temporary buffer used to avoid allocations.
	CallFromStack(context *Context, n int, scratch []reflect.Value) ([]reflect.Value, error)
	// Call a function using arguments passed to the method. Scratch is a temporary
	// buffer used to avoid allocations. Execution happens in the VM specified by
	// the context, and the stack is left unmodified.
	//CallWithArgs(context *Context, args []interface{}, scratch []reflect.Value) (interface{}, []reflect.Value, error)
	ToFunctionValue(t reflect.Type, context *Context) reflect.Value
	Expression() b6.Expression
}

type Adaptors struct {
	Functions   map[reflect.Type]func(Callable) reflect.Value
	Collections map[reflect.Type]func(b6.UntypedCollection) reflect.Value
}

func Call0(context *Context, c Callable) (interface{}, error) {
	return context.VM.CallWithArgs(context, c, []interface{}{})
}

func Call1(context *Context, arg0 interface{}, c Callable) (interface{}, error) {
	return context.VM.CallWithArgs(context, c, []interface{}{arg0})
}

func Call2(context *Context, arg0 interface{}, arg1 interface{}, c Callable) (interface{}, error) {
	return context.VM.CallWithArgs(context, c, []interface{}{arg0, arg1})
}

type Instruction struct {
	Op         Op
	Args       [2]int16
	Value      reflect.Value
	Callable   Callable
	Expression b6.Expression
}

func (i Instruction) String() string {
	if i.Value.Kind() != reflect.Invalid && i.Value.CanInterface() {
		return fmt.Sprintf("%-12s %d,%d %v", i.Op, i.Args[0], i.Args[1], i.Value.Interface())
	} else if i.Callable != nil {
		return fmt.Sprintf("%-12s %d,%d %s", i.Op, i.Args[0], i.Args[1], i.Callable)
	} else {
		return fmt.Sprintf("%-12s %d,%d", i.Op, i.Args[0], i.Args[1])
	}
}

type target struct {
	Expression b6.Expression
	Args       *frame
	Done       func(entrypoint int)
}

type frame struct {
	Symbols  []string
	Args     []int
	Previous *frame
}

func (f *frame) Bind(s string, i int) {
	f.Symbols = append(f.Symbols, s)
	f.Args = append(f.Args, i)
}

func (f *frame) Lookup(s b6.SymbolExpression) (int, bool) {
	if f == nil {
		return 0, false
	}
	for i, ss := range f.Symbols {
		if string(s) == ss {
			return f.Args[i], true
		}
	}
	return f.Previous.Lookup(s)
}

type compilation struct {
	Globals      Symbols
	Instructions []Instruction
	Targets      []target
	Args         *frame
	NumArgs      int
}

func (c *compilation) Append(i Instruction) {
	c.Instructions = append(c.Instructions, i)
}

func Evaluate(expression b6.Expression, context *Context) (interface{}, error) {
	vm, err := newVM(expression, context.FunctionSymbols)
	if err != nil {
		return nil, err
	}
	return vm.Execute(context)
}

func EvaluateAndFill(expression b6.Expression, context *Context, toFill interface{}) error {
	vm, err := newVM(expression, context.FunctionSymbols)
	if err != nil {
		return err
	}
	v := reflect.ValueOf(toFill)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("expected a pointer, found %T", toFill)
	}
	r, err := vm.Execute(context)
	if err != nil {
		return err
	}
	c, err := ConvertWithContext(reflect.ValueOf(r), v.Type().Elem(), context)
	if err != nil {
		return err
	}
	v.Elem().Set(c)
	return nil
}

func EvaluateString(e string, context *Context) (interface{}, error) {
	expression, err := ParseExpression(e)
	if err != nil {
		return nil, err
	}
	expression = Simplify(expression, context.FunctionSymbols)
	vm, err := newVM(expression, context.FunctionSymbols)
	if err != nil {
		return nil, err
	}
	return vm.Execute(context)
}

func newVM(expression b6.Expression, fs FunctionSymbols) (*VM, error) {
	c := compilation{
		Globals: fs,
		Targets: []target{{Expression: expression, Done: func(int) {}}},
		Instructions: []Instruction{{
			Op:         OpPushValue,
			Value:      reflect.ValueOf(0),
			Expression: expression,
		}},
	}
	for i := 0; i < len(c.Targets); i++ {
		entrypoint := len(c.Instructions)
		c.Args = c.Targets[i].Args
		if c.Args != nil {
			c.Append(Instruction{Op: OpStore})
		}
		if err := compile(c.Targets[i].Expression, &c); err != nil {
			return nil, err
		}
		c.Append(Instruction{Op: OpDiscard})
		c.Append(Instruction{Op: OpReturn})
		c.Targets[i].Done(entrypoint)
	}
	return &VM{Instructions: c.Instructions}, nil
}

func compile(e b6.Expression, c *compilation) error {
	var err error
	switch e.AnyExpression.(type) {
	case *b6.CallExpression:
		err = compileCall(e, c)
	case *b6.SymbolExpression:
		err = compileSymbol(e, c)
	case *b6.LambdaExpression:
		var l *lambdaCall
		l, err = compileLambda(e, c)
		if err == nil {
			c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(l)})
		}
	case b6.AnyLiteral:
		err = compileLiteral(e, c)
	default:
		err = fmt.Errorf("Don't know how to compile %T", e)
	}
	return err
}

func compileCall(e b6.Expression, c *compilation) error {
	call := e.AnyExpression.(*b6.CallExpression)
	for _, a := range call.Args {
		if err := compile(a, c); err != nil {
			return err
		}
	}
	var args [2]int16
	args[ArgsNumArgs] = int16(len(call.Args))
	switch f := call.Function.AnyExpression.(type) {
	case *b6.SymbolExpression:
		if ff, ok := c.Globals.Function(*f); ok {
			callable := goCall{f: ff, expression: call.Function}
			c.Append(Instruction{Op: OpCallValue, Callable: callable, Args: args, Expression: e})
		} else {
			return fmt.Errorf("undefined symbol %q", *f)
		}
	case *b6.LambdaExpression:
		if l, err := compileLambda(e, c); err == nil {
			c.Append(Instruction{Op: OpCallValue, Callable: l, Args: args, Expression: e})
		} else {
			return err
		}
	case *b6.CallExpression:
		if err := compile(call.Function, c); err != nil {
			return err
		}
		c.Append(Instruction{Op: OpCallStack, Args: [2]int16{int16(len(call.Args)), 0}, Expression: e})
	default:
		return fmt.Errorf("can't call %T", call.Function.AnyExpression)
	}
	return nil
}

func compileSymbol(e b6.Expression, c *compilation) error {
	symbol := e.AnyExpression.(*b6.SymbolExpression)
	if a, ok := c.Args.Lookup(*symbol); ok {
		c.Append(Instruction{Op: OpLoad, Args: [2]int16{int16(a)}, Expression: e})
	} else if f, ok := c.Globals.Function(*symbol); ok {
		c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(&goCall{f: f, expression: e}), Expression: e})
	} else {
		return fmt.Errorf("undefined symbol %q", symbol)
	}
	return nil
}

func compileLambda(e b6.Expression, c *compilation) (*lambdaCall, error) {
	lambda := e.AnyExpression.(*b6.LambdaExpression)
	l := &lambdaCall{args: len(lambda.Args), expression: e}
	f := &frame{Previous: c.Args}
	for _, s := range lambda.Args {
		if c.NumArgs > MaxArgs {
			return nil, fmt.Errorf("Can't use more than %d args", MaxArgs)
		}
		f.Bind(s, c.NumArgs)
		c.NumArgs++
	}
	c.Targets = append(c.Targets, target{
		Expression: lambda.Expression,
		Args:       f,
		Done:       func(entrypoint int) { l.pc = entrypoint },
	})
	return l, nil
}

func compileLiteral(e b6.Expression, c *compilation) error {
	l := e.AnyExpression.(b6.AnyLiteral)
	c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(l.Literal()), Expression: e})
	return nil
}

const MaxArgs = 32

type StackFrame struct {
	Value      reflect.Value
	Expression b6.Expression
}

type VM struct {
	Instructions []Instruction
	PC           int
	Args         [MaxArgs]StackFrame
	Stack        []StackFrame
}

const (
	ArgsJumpDestination = 0

	ArgsNumArgs    = 0
	ArgsEntrypoint = 1

	ArgsArgToPush = 0
)

func (v *VM) Fork(n int) []VM {
	vms := make([]VM, n)
	for i := range vms {
		vms[i] = *v
		vms[i].Stack = make([]StackFrame, len(v.Stack))
		copy(vms[i].Stack, v.Stack)
	}
	return vms
}

func (v *VM) Execute(context *Context) (interface{}, error) {
	if err := v.execute(context); err != nil {
		return nil, err
	}
	result := v.Stack[len(v.Stack)-1].Value.Interface()
	v.Stack = v.Stack[0 : len(v.Stack)-1]
	return result, nil
}

func (v *VM) execute(context *Context) error {
	context.VM = v
	var err error
	args := make([]reflect.Value, 0)
	done := false
	for !done {
		switch v.Instructions[v.PC].Op {
		case OpJump:
			v.PC = int(v.Instructions[v.PC].Args[ArgsJumpDestination]) - 1 // Incremented below
		case OpPushValue:
			v.Stack = append(v.Stack, StackFrame{
				Value:      v.Instructions[v.PC].Value,
				Expression: v.Instructions[v.PC].Expression,
			})
		case OpStore:
			n := int(v.Stack[len(v.Stack)-1].Value.Int())
			argsStart := len(v.Stack) - n - 1
			for i := 0; i < n; i++ {
				v.Args[i] = v.Stack[argsStart+i]
			}
		case OpDiscard:
			n := int(v.Stack[len(v.Stack)-2].Value.Int())
			v.Stack[len(v.Stack)-n-2] = v.Stack[len(v.Stack)-1]
			v.Stack = v.Stack[0 : len(v.Stack)-n-1]
		case OpLoad:
			v.Stack = append(v.Stack, v.Args[v.Instructions[v.PC].Args[0]])
		case OpCallValue:
			// Before a call, the stack contains the arguments for the
			// function, followed by the expression representing the
			// function call. The function is responsible for removing
			// the arguments and expression from the stack, and leaving
			// the result.
			n := int(v.Instructions[v.PC].Args[ArgsNumArgs])
			v.Stack = append(v.Stack, StackFrame{
				Value:      reflect.ValueOf(n),
				Expression: v.Instructions[v.PC].Expression,
			})
			if args, err = v.Instructions[v.PC].Callable.CallFromStack(context, n, args); err != nil {
				return err
			}
		case OpCallStack:
			n := int(v.Instructions[v.PC].Args[ArgsNumArgs])
			f := v.Stack[len(v.Stack)-1].Value.Interface().(Callable)
			v.Stack[len(v.Stack)-1].Value = reflect.ValueOf(n)
			if args, err = f.CallFromStack(context, n, args); err != nil {
				return err
			}
		case OpReturn:
			done = true
		default:
			return fmt.Errorf("can't execute instruction %+v", v.Instructions[v.PC])
		}
		v.PC++
	}
	return nil
}

func (v *VM) CallWithArgsAndExpressions(context *Context, c Callable, args []StackFrame) (interface{}, error) {
	l := len(v.Stack)
	v.Stack = append(v.Stack, args...)

	var aes [MaxArgs]b6.Expression
	for i, arg := range args {
		aes[i] = arg.Expression
	}
	e := b6.NewCallExpression(c.Expression(), aes[0:len(args)])

	v.Stack = append(v.Stack, StackFrame{Value: reflect.ValueOf(len(args)), Expression: e})
	var scratch [MaxArgs]reflect.Value
	_, err := c.CallFromStack(context, len(args), scratch[0:])
	var result interface{}
	if err == nil {
		result = v.Stack[len(v.Stack)-1].Value.Interface()
	}
	v.Stack = v.Stack[0:l]
	return result, err
}

func (v *VM) CallWithArgs(context *Context, c Callable, args []interface{}) (interface{}, error) {
	var frames [MaxArgs]StackFrame
	for i, arg := range args {
		literal, err := b6.FromLiteral(arg)
		if err != nil {
			return nil, err
		}
		frames[i].Value = reflect.ValueOf(arg)
		frames[i].Expression = b6.Expression{AnyExpression: literal.AnyLiteral}
	}
	return v.CallWithArgsAndExpressions(context, c, frames[0:len(args)])
}

func (v *VM) Expression() b6.Expression {
	if len(v.Stack) == 0 {
		panic("not in a call, stack empty")
	}
	_, ok := v.Stack[len(v.Stack)-1].Expression.AnyExpression.(*b6.CallExpression)
	if !ok {
		panic("not in a call")
	}
	return v.Stack[len(v.Stack)-1].Expression
}

func (v *VM) ArgExpressions() []b6.Expression {
	if len(v.Stack) == 0 {
		panic("not in a call, stack empty")
	}
	_, ok := v.Stack[len(v.Stack)-1].Expression.AnyExpression.(*b6.CallExpression)
	if !ok {
		panic("not in a call")
	}
	n := int(v.Stack[len(v.Stack)-1].Value.Int())
	argsStart := len(v.Stack) - n - 1
	args := make([]b6.Expression, n)
	for i := range args {
		args[i] = v.Stack[argsStart+i].Expression
	}
	return args
}

func (v *VM) WriteStackForDebug(w io.Writer) {
	w.Write([]byte("stack:\n"))
	for i := range v.Stack {
		if s, ok := UnparseExpression(v.Stack[i].Expression); ok {
			w.Write([]byte(fmt.Sprintf("  %d: %v %s\n", i, v.Stack[i].Value, s)))
		} else {
			w.Write([]byte(fmt.Sprintf("  %d: %v %T\n", i, v.Stack[i].Value, v.Stack[i].Expression.AnyExpression)))
		}
	}
}

func NewNativeFunction0[Returns any](f func(*Context) (Returns, error), e b6.Expression) Callable {
	return goCall{
		f:          reflect.ValueOf(f),
		expression: e,
	}
}

func NewNativeFunction1[Arg0 any, Returns any](f func(*Context, Arg0) (Returns, error), e b6.Expression) Callable {
	return goCall{
		f:          reflect.ValueOf(f),
		expression: e,
	}
}

// TODO(andrew): rename to nativeFunction
type goCall struct {
	f          reflect.Value
	expression b6.Expression
}

func (g goCall) NumArgs() int {
	return g.f.Type().NumIn() - 1
}

func (g goCall) Expression() b6.Expression {
	return g.expression
}

func (g goCall) String() string {
	return g.expression.String()
}

// TODO: use context to store scrach arg []reflect.Value?
func (g goCall) CallFromStack(context *Context, n int, scratch []reflect.Value) ([]reflect.Value, error) {
	vm := context.VM
	t := g.f.Type()
	scratch = scratch[0:0]
	expected := g.NumArgs()
	if t.IsVariadic() {
		// handle the last argument separately, as it's turned into a slice
		expected--
	} else if n > expected {
		return nil, fmt.Errorf("%s: expected %d arguments, found %d", g.String(), expected, n)
	}
	argsStart := len(vm.Stack) - n - 1
	expression := vm.Stack[len(vm.Stack)-1].Expression
	if n >= expected {
		scratch = append(scratch, reflect.ValueOf(context))
		for i := 0; i < expected; i++ {
			arg := argsStart + i
			if v, err := ConvertWithContext(vm.Stack[arg].Value, t.In(i+1), context); err == nil {
				scratch = append(scratch, v)
			} else {
				return nil, fmt.Errorf("%s: %s", g.String(), err.Error())
			}
		}
		if t.IsVariadic() {
			for i := expected; i < n; i++ {
				arg := argsStart + i
				if v, err := ConvertWithContext(vm.Stack[arg].Value, t.In(expected+1).Elem(), context); err == nil {
					scratch = append(scratch, v)
				} else {
					return scratch, fmt.Errorf("%s: %s", g.String(), err.Error())
				}
			}
		}
		result := g.f.Call(scratch)
		vm.Stack = vm.Stack[0:argsStart]
		if len(result) < 1 {
			// Checked during init() in functions, but not guaranteed if
			// unvalidated local functions have been registered.
			panic(fmt.Sprintf("expected 2 results from %s", g.f.Type()))
		}
		if len(result) > 1 {
			if err, ok := result[1].Interface().(error); ok && err != nil {
				vm.Stack = append(vm.Stack, StackFrame{Value: reflect.Value{}, Expression: expression})
				return nil, err
			}
		}
		if result[0].Kind() == reflect.Func {
			vm.Stack = append(vm.Stack, StackFrame{
				Value:      reflect.ValueOf(&goCall{f: result[0], expression: expression}),
				Expression: expression,
			})
		} else {
			vm.Stack = append(vm.Stack, StackFrame{
				Value:      result[0],
				Expression: expression,
			})
		}
	} else {
		p := &partialCall{c: g, e: expression, vmArgs: vm.Args}
		p.n = n
		for i := 0; i < n; i++ {
			p.args[i] = vm.Stack[argsStart+i]
		}
		vm.Stack = vm.Stack[0:argsStart]
		vm.Stack = append(vm.Stack, StackFrame{
			Value:      reflect.ValueOf(p),
			Expression: expression,
		})
	}
	return scratch, nil
}

func (g goCall) ToFunctionValue(t reflect.Type, context *Context) reflect.Value {
	// If the underlying function matches the Go type we need, we can
	// return the function itself, otherwise, we need to call it via
	// CallWithArgs, which handles the necessary conversions.
	// The latter path is triggered, for example, when we pass a specific
	// function, eg Area -> Geometry, to map which expects
	// interface{} -> interface{}.
	if g.f.Type().AssignableTo(t) {
		return g.f
	}
	if w, ok := context.Adaptors.Functions[t]; ok {
		return w(g)
	}
	panic(fmt.Sprintf("Can't convert values of type %s", t)) // Checked in init()
}

type lambdaCall struct {
	pc         int
	args       int
	expression b6.Expression
}

func (l *lambdaCall) NumArgs() int {
	return l.args
}

func (l *lambdaCall) Expression() b6.Expression {
	return l.expression
}

func (l *lambdaCall) String() string {
	return fmt.Sprintf("lambda: pc: %d args: %d", l.pc, l.args)
}

func (l *lambdaCall) CallFromStack(context *Context, n int, scratch []reflect.Value) ([]reflect.Value, error) {
	var err error
	vm := context.VM
	argsStart := len(vm.Stack) - n - 1
	expression := vm.Stack[len(vm.Stack)-1].Expression
	if n >= l.args {
		opc := vm.PC
		vm.PC = l.pc
		err = vm.execute(context)
		vm.PC = opc
	} else if n < l.args {
		p := &partialCall{c: l, e: expression, vmArgs: vm.Args}
		p.n = n
		for i := 0; i < n; i++ {
			p.args[i] = vm.Stack[argsStart+i]
		}
		vm.Stack = vm.Stack[0:argsStart]
		vm.Stack = append(vm.Stack, StackFrame{
			Value:      reflect.ValueOf(p),
			Expression: expression,
		})
	} else {
		err = fmt.Errorf("lambda: expected at most %d args, found %d", l.args, n)
	}
	return scratch, err
}

func (l *lambdaCall) ToFunctionValue(t reflect.Type, context *Context) reflect.Value {
	if w, ok := context.Adaptors.Functions[t]; ok {
		return w(l)
	}
	panic(fmt.Sprintf("Can't convert values of type %s", t)) // Checked in init()
}

type partialCall struct {
	c      Callable
	e      b6.Expression
	n      int
	args   [MaxArgs]StackFrame
	vmArgs [MaxArgs]StackFrame
}

func (p *partialCall) NumArgs() int {
	return p.c.NumArgs() - p.n
}

func (p *partialCall) CallFromStack(context *Context, n int, scratch []reflect.Value) ([]reflect.Value, error) {
	vm := context.VM
	expression := vm.Stack[len(vm.Stack)-1].Expression
	if n+p.n == p.c.NumArgs() {
		vm.Stack = vm.Stack[0 : len(vm.Stack)-1]
		vm.Stack = append(vm.Stack, p.args[0:p.n]...)
		call := &b6.CallExpression{
			Function: expression.AnyExpression.(*b6.CallExpression).Function,
			Args:     make([]b6.Expression, p.c.NumArgs()),
		}
		for i := 0; i < p.c.NumArgs(); i++ {
			call.Args[i] = vm.Stack[len(vm.Stack)-p.c.NumArgs()+i].Expression
		}
		vm.Stack = append(vm.Stack, StackFrame{Value: reflect.ValueOf(n + p.n), Expression: b6.Expression{AnyExpression: call}})
		oargs := vm.Args
		vm.Args = p.vmArgs
		scratch, err := p.c.CallFromStack(context, n+p.n, scratch)
		vm.Args = oargs
		return scratch, err
	} else if n+p.n < p.c.NumArgs() {
		argsStart := len(vm.Args) - n - 1
		es := make([]b6.Expression, 0, n)
		pp := &partialCall{c: p, e: expression, vmArgs: vm.Args}
		pp.n = n
		for i := 0; i < n; i++ {
			pp.args[i] = vm.Stack[argsStart+i]
			es[i] = vm.Stack[argsStart+i].Expression
		}
		vm.Stack = vm.Stack[0:argsStart]
		vm.Stack = append(vm.Stack, StackFrame{
			Value:      reflect.ValueOf(pp),
			Expression: expression,
		})
	} else {
		return scratch, fmt.Errorf("(partial): expected at most %d args, found %d", p.c.NumArgs()-p.n, n)
	}
	return scratch, nil
}

func (p *partialCall) ToFunctionValue(t reflect.Type, context *Context) reflect.Value {
	if w, ok := context.Adaptors.Functions[t]; ok {
		return w(p)
	}
	panic(fmt.Sprintf("Can't convert values of type %s", t)) // Checked in init()
}

func (p *partialCall) Expression() b6.Expression {
	return p.e
}
