package api

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"diagonal.works/b6"
	pb "diagonal.works/b6/proto"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

type Context struct {
	World            b6.World
	Cores            int
	Clock            func() time.Time
	Values           map[interface{}]interface{}
	FunctionSymbols  FunctionSymbols
	FunctionWrappers FunctionWrappers
	Context          context.Context

	VM *VM
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
	// The number of arguments expected by the function, excluding Context
	NumArgs() int
	// Call a function using arguments on the VM's stack. Assumes that n
	// arguments have been pushed onto the stack in left-to-right order. These
	// arguments are removed from the stack, leaving the function's
	// result at the top of the stack.
	// Scratch is a temporary buffer used to avoid allocations.
	CallFromStack(n int, scratch []reflect.Value, context *Context) ([]reflect.Value, error)
	// Call a function using arguments passed to yhe method. Scratch is a temporary
	// buffer used to avoid allocations. Execution happens in the VM specified by
	// the context, and the stack is left unmodified.
	CallWithArgs(args []interface{}, scratch []reflect.Value, context *Context) (interface{}, []reflect.Value, error)
	ToFunctionValue(t reflect.Type, context *Context) reflect.Value
	String() string
}

type FunctionWrappers map[reflect.Type]func(Callable) reflect.Value

func Call0(c Callable, context *Context) (interface{}, error) {
	var scratch [MaxArgs]reflect.Value
	result, _, err := c.CallWithArgs([]interface{}{}, scratch[0:], context)
	return result, err
}

func Call1(arg0 interface{}, c Callable, context *Context) (interface{}, error) {
	var scratch [MaxArgs]reflect.Value
	result, _, err := c.CallWithArgs([]interface{}{arg0}, scratch[0:], context)
	return result, err
}

func Call2(arg0 interface{}, arg1 interface{}, c Callable, context *Context) (interface{}, error) {
	var scratch [MaxArgs]reflect.Value
	result, _, err := c.CallWithArgs([]interface{}{arg0, arg1}, scratch[0:], context)
	return result, err
}

type Instruction struct {
	Op       Op
	Args     [2]int16
	Value    reflect.Value
	Callable Callable
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
	Node *pb.NodeProto
	Args *frame
	Done func(entrypoint int)
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

func (f *frame) Lookup(s string) (int, bool) {
	if f == nil {
		return 0, false
	}
	for i, ss := range f.Symbols {
		if s == ss {
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

func Evaluate(node *pb.NodeProto, context *Context) (interface{}, error) {
	vm, err := newVM(node, context.FunctionSymbols)
	if err != nil {
		return nil, err
	}
	return vm.Execute(context)
}

func EvaluateAndFill(node *pb.NodeProto, context *Context, toFill interface{}) error {
	vm, err := newVM(node, context.FunctionSymbols)
	if err != nil {
		return err
	}
	v := reflect.ValueOf(toFill)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("Expected a pointer, found %T", toFill)
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
	node, err := ParseExpression(e)
	if err != nil {
		return nil, err
	}
	node = Simplify(node, context.FunctionSymbols)
	vm, err := newVM(node, context.FunctionSymbols)
	if err != nil {
		return nil, err
	}
	return vm.Execute(context)
}

func newVM(node *pb.NodeProto, fs FunctionSymbols) (*VM, error) {
	c := compilation{
		Globals: fs,
		Targets: []target{{Node: node, Done: func(int) {}}},
	}
	for i := 0; i < len(c.Targets); i++ {
		entrypoint := len(c.Instructions)
		c.Args = c.Targets[i].Args
		if c.Args != nil {
			for i := len(c.Args.Symbols) - 1; i >= 0; i-- {
				c.Append(Instruction{Op: OpStore, Args: [2]int16{int16(c.Args.Args[i])}})
			}
		}
		if err := compile(c.Targets[i].Node, &c); err != nil {
			return nil, err
		}
		c.Append(Instruction{Op: OpReturn})
		c.Targets[i].Done(entrypoint)
	}
	return &VM{Instructions: c.Instructions}, nil
}

func compile(p *pb.NodeProto, c *compilation) error {
	var err error
	switch n := p.Node.(type) {
	case *pb.NodeProto_Call:
		err = compileCall(n.Call, c)
	case *pb.NodeProto_Symbol:
		err = compileSymbol(n.Symbol, c)
	case *pb.NodeProto_Lambda_:
		var l *lambdaCall
		l, err = compileLambda(n.Lambda_, c)
		if err == nil {
			c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(l)})
		}
	case *pb.NodeProto_Literal:
		err = compileLiteral(n.Literal, c)
	default:
		err = fmt.Errorf("Don't know how to compile node %T", p.Node)
	}
	return err
}

func compileCall(call *pb.CallNodeProto, c *compilation) error {
	for _, a := range call.Args {
		if err := compile(a, c); err != nil {
			return err
		}
	}
	var args [2]int16
	args[ArgsNumArgs] = int16(len(call.Args))
	if symbol, ok := call.Function.Node.(*pb.NodeProto_Symbol); ok {
		if f, ok := c.Globals.Function(symbol.Symbol); ok {
			c.Append(Instruction{Op: OpCallValue, Callable: goCall{f: f, name: symbol.Symbol}, Args: args})
		} else {
			return fmt.Errorf("Undefined symbol %q", symbol.Symbol)
		}
	} else if f, ok := call.Function.Node.(*pb.NodeProto_Lambda_); ok {
		if l, err := compileLambda(f.Lambda_, c); err == nil {
			c.Append(Instruction{Op: OpCallValue, Callable: l, Args: args})
		} else {
			return err
		}
	} else if _, ok := call.Function.Node.(*pb.NodeProto_Call); ok {
		if err := compile(call.Function, c); err != nil {
			return err
		}
		c.Append(Instruction{Op: OpCallStack, Args: [2]int16{int16(len(call.Args)), 0}})
	} else {
		return fmt.Errorf("Can't call %T", call.Function.Node)
	}
	return nil
}

func compileSymbol(symbol string, c *compilation) error {
	if a, ok := c.Args.Lookup(symbol); ok {
		c.Append(Instruction{Op: OpLoad, Args: [2]int16{int16(a)}})
	} else if f, ok := c.Globals.Function(symbol); ok {
		c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(&goCall{f: f, name: symbol})})
	} else {
		return fmt.Errorf("Undefined symbol %q", symbol)
	}
	return nil
}

func compileLambda(lambda *pb.LambdaNodeProto, c *compilation) (*lambdaCall, error) {
	l := &lambdaCall{args: len(lambda.Args)}
	f := &frame{Previous: c.Args}
	for _, s := range lambda.Args {
		if c.NumArgs > MaxArgs {
			return nil, fmt.Errorf("Can't use more than %d args", MaxArgs)
		}
		f.Bind(s, c.NumArgs)
		c.NumArgs++
	}
	c.Targets = append(c.Targets, target{
		Node: lambda.Node,
		Args: f,
		Done: func(entrypoint int) { l.pc = entrypoint },
	})
	return l, nil
}

func compileLiteral(literal *pb.LiteralNodeProto, c *compilation) error {
	switch v := literal.Value.(type) {
	case *pb.LiteralNodeProto_NilValue:
		c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(nil)})
	case *pb.LiteralNodeProto_BoolValue:
		c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(v.BoolValue)})
	case *pb.LiteralNodeProto_StringValue:
		c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(v.StringValue)})
	case *pb.LiteralNodeProto_IntValue:
		c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(v.IntValue)})
	case *pb.LiteralNodeProto_FloatValue:
		c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(v.FloatValue)})
	case *pb.LiteralNodeProto_QueryValue:
		if q, err := b6.NewQueryFromProto(v.QueryValue); err == nil {
			c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(q)})
		} else {
			return err
		}
	case *pb.LiteralNodeProto_FeatureIDValue:
		id := b6.NewFeatureIDFromProto(v.FeatureIDValue)
		c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(id)})
	case *pb.LiteralNodeProto_PointValue:
		ll := s2.LatLng{Lat: s1.Angle(v.PointValue.LatE7) * s1.E7, Lng: s1.Angle(v.PointValue.LngE7) * s1.E7}
		c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(b6.PointFromLatLng(ll))})
	case *pb.LiteralNodeProto_PathValue:
		p := b6.PolylineProtoToS2Polyline(v.PathValue)
		c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(b6.PathFromS2Points(*p))})
	case *pb.LiteralNodeProto_AreaValue:
		m := b6.MultiPolygonProtoToS2MultiPolygon(v.AreaValue)
		c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(b6.AreaFromS2Polygons(m))})
	case *pb.LiteralNodeProto_TagValue:
		tag := b6.Tag{Key: v.TagValue.Key, Value: v.TagValue.Value}
		c.Append(Instruction{Op: OpPushValue, Value: reflect.ValueOf(tag)})
	default:
		return fmt.Errorf("Don't know how to compile literal %T", literal.Value)
	}
	return nil
}

const MaxArgs = 32

type VM struct {
	Instructions []Instruction
	PC           int
	Args         [MaxArgs]reflect.Value
	Stack        []reflect.Value
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
		vms[i].Stack = make([]reflect.Value, len(v.Stack))
		copy(vms[i].Stack, v.Stack)
	}
	return vms
}

func (v *VM) Execute(context *Context) (interface{}, error) {
	if err := v.execute(context); err != nil {
		return nil, err
	}
	result := v.Stack[len(v.Stack)-1].Interface()
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
			v.Stack = append(v.Stack, v.Instructions[v.PC].Value)
		case OpStore:
			v.Args[v.Instructions[v.PC].Args[0]] = v.Stack[len(v.Stack)-1]
			v.Stack = v.Stack[0 : len(v.Stack)-1]
		case OpLoad:
			v.Stack = append(v.Stack, v.Args[v.Instructions[v.PC].Args[0]])
		case OpCallValue:
			n := int(v.Instructions[v.PC].Args[ArgsNumArgs])
			if args, err = v.Instructions[v.PC].Callable.CallFromStack(n, args, context); err != nil {
				return err
			}
		case OpCallStack:
			n := int(v.Instructions[v.PC].Args[ArgsNumArgs])
			f := v.Stack[len(v.Stack)-1].Interface().(Callable)
			v.Stack = v.Stack[0 : len(v.Stack)-1]
			if args, err = f.CallFromStack(n, args, context); err != nil {
				return err
			}
		case OpReturn:
			done = true
		default:
			return fmt.Errorf("Can't execute instruction %+v", v.Instructions[v.PC])
		}
		v.PC++
	}
	return nil
}

type goCall struct {
	f    reflect.Value
	name string
}

func (g goCall) NumArgs() int {
	return g.f.Type().NumIn() - 1
}

func (g goCall) String() string {
	return "go: " + g.name
}

func (g goCall) CallFromStack(n int, scratch []reflect.Value, context *Context) ([]reflect.Value, error) {
	vm := context.VM
	t := g.f.Type()
	scratch = scratch[0:0]
	expected := g.NumArgs()
	if n == expected {
		for i := 0; i < n; i++ {
			arg := len(vm.Stack) - n + i
			if v, err := ConvertWithContext(vm.Stack[arg], t.In(i), context); err == nil {
				scratch = append(scratch, v)
			} else {
				return scratch, fmt.Errorf("%s: %s", g.name, err.Error())
			}
		}
		scratch = append(scratch, reflect.ValueOf(context))
		result := g.f.Call(scratch)
		vm.Stack = vm.Stack[0 : len(vm.Stack)-n]
		if len(result) < 1 {
			// TODO: Check return type during compilation
			panic(fmt.Sprintf("expected 2 results from %s", g.f.Type()))
		}
		if len(result) > 1 {
			if err, ok := result[1].Interface().(error); ok && err != nil {
				vm.Stack = append(vm.Stack, reflect.Value{})
				return nil, err
			}
		}
		if result[0].Kind() == reflect.Func {
			vm.Stack = append(vm.Stack, reflect.ValueOf(&goCall{f: result[0], name: "(result)"}))
		} else {
			vm.Stack = append(vm.Stack, result[0])
		}
	} else if n < expected {
		p := &partialCall{c: g, vmArgs: vm.Args}
		p.n = n
		for i := 0; i < n; i++ {
			p.args[i] = vm.Stack[len(vm.Stack)-n+i]
		}
		vm.Stack = vm.Stack[0 : len(vm.Stack)-n]
		vm.Stack = append(vm.Stack, reflect.ValueOf(p))
	} else {
		return scratch, fmt.Errorf("%s: expected at most %d args, found %d", g.name, expected, n)
	}
	return scratch, nil
}

func (g goCall) CallWithArgs(args []interface{}, scratch []reflect.Value, context *Context) (interface{}, []reflect.Value, error) {
	t := g.f.Type()
	if len(args) == t.NumIn()-1 { // Don't count context
		scratch = scratch[0:0]
		for i, arg := range args {
			if v, err := ConvertWithContext(reflect.ValueOf(arg), t.In(i), context); err == nil {
				scratch = append(scratch, v)
			} else {
				return nil, scratch, fmt.Errorf("%s: %s", g.name, err.Error())
			}
		}
		scratch = append(scratch, reflect.ValueOf(context))
		result := g.f.Call(scratch)
		if err, ok := result[1].Interface().(error); ok && err != nil {
			return nil, scratch, err
		}
		return result[0].Interface(), scratch, nil
	} else if len(args) < t.NumIn()-1 {
		p := &partialCall{c: g, vmArgs: context.VM.Args}
		p.n = len(args)
		for i, arg := range args {
			p.args[i] = reflect.ValueOf(arg)
		}
		return p.ToFunction(), scratch, nil
	} else {
		return nil, scratch, fmt.Errorf("%s: expected at most %d args, found %d", g.name, t.NumIn()-1, len(args))
	}
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
	if w, ok := context.FunctionWrappers[t]; ok {
		return w(g)
	}
	panic(fmt.Sprintf("Can't convert values of type %s", t)) // Checked in init()
}

type lambdaCall struct {
	pc   int
	args int
}

func (l *lambdaCall) NumArgs() int {
	return l.args
}

func (l *lambdaCall) String() string {
	return fmt.Sprintf("lambda: pc: %d args: %d", l.pc, l.args)
}

func (l *lambdaCall) CallFromStack(n int, scratch []reflect.Value, context *Context) ([]reflect.Value, error) {
	var err error
	vm := context.VM
	if n == l.args {
		opc := vm.PC
		vm.PC = l.pc
		err = vm.execute(context)
		vm.PC = opc
	} else if n < l.args {
		p := &partialCall{c: l, vmArgs: vm.Args}
		p.n = n
		for i := 0; i < n; i++ {
			p.args[i] = vm.Stack[len(vm.Stack)-n+i]
		}
		vm.Stack = vm.Stack[0 : len(vm.Stack)-n]
		vm.Stack = append(vm.Stack, reflect.ValueOf(p))
	} else {
		err = fmt.Errorf("lambda: expected at most %d args, found %d", l.args, n)
	}
	return scratch, err
}

func (l *lambdaCall) CallWithArgs(args []interface{}, scratch []reflect.Value, context *Context) (interface{}, []reflect.Value, error) {
	vm := context.VM
	for _, arg := range args {
		vm.Stack = append(vm.Stack, reflect.ValueOf(arg))
	}
	scratch, err := l.CallFromStack(len(args), scratch, context)
	var result interface{}
	if err == nil {
		result = vm.Stack[len(vm.Stack)-1].Interface()
	}
	vm.Stack = vm.Stack[0 : len(vm.Stack)-1]
	return result, scratch, err
}

func (l *lambdaCall) ToFunctionValue(t reflect.Type, context *Context) reflect.Value {
	if w, ok := context.FunctionWrappers[t]; ok {
		return w(l)
	}
	panic(fmt.Sprintf("Can't convert values of type %s", t)) // Checked in init()
}

type partialCall struct {
	c      Callable
	n      int
	args   [MaxArgs]reflect.Value
	vmArgs [MaxArgs]reflect.Value
}

func (p *partialCall) NumArgs() int {
	return p.c.NumArgs() - p.n
}

func (p *partialCall) CallFromStack(n int, scratch []reflect.Value, context *Context) ([]reflect.Value, error) {
	vm := context.VM
	if n+p.n == p.c.NumArgs() {
		vm.Stack = append(vm.Stack, p.args[0:p.n]...)
		oargs := vm.Args
		vm.Args = p.vmArgs
		scratch, err := p.c.CallFromStack(n+p.n, scratch, context)
		vm.Args = oargs
		return scratch, err
	} else if n+p.n < p.c.NumArgs() {
		pp := &partialCall{c: p, vmArgs: vm.Args}
		pp.n = n
		for i := 0; i < n; i++ {
			pp.args[i] = vm.Stack[len(vm.Stack)-n+i]
		}
		vm.Stack = vm.Stack[0 : len(vm.Stack)-n]
		vm.Stack = append(vm.Stack, reflect.ValueOf(pp))
	} else {
		return scratch, fmt.Errorf("(partial): expected at most %d args, found %d", p.c.NumArgs()-p.n, n)
	}
	return scratch, nil
}

func (p *partialCall) CallWithArgs(args []interface{}, scratch []reflect.Value, context *Context) (interface{}, []reflect.Value, error) {
	if len(args)+p.n == p.c.NumArgs() {
		added := make([]interface{}, len(args)+p.n)
		for i, arg := range args {
			added[i] = arg
		}
		for i := 0; i < p.n; i++ {
			added[i+len(args)] = p.args[i].Interface()
		}
		return p.c.CallWithArgs(added, scratch, context)
	} else if len(args)+p.n < p.c.NumArgs() {
		pp := &partialCall{c: p.c, vmArgs: context.VM.Args}
		pp.n = len(args)
		for i, arg := range args {
			pp.args[i] = reflect.ValueOf(arg)
		}
		return p.ToFunction(), scratch, nil
	} else {
		return nil, scratch, fmt.Errorf("(partial): expected at most %d args, found %d", p.c.NumArgs()-p.n, len(args))
	}
}

func (p *partialCall) ToFunctionValue(t reflect.Type, context *Context) reflect.Value {
	if w, ok := context.FunctionWrappers[t]; ok {
		return w(p)
	}
	panic(fmt.Sprintf("Can't convert values of type %s", t)) // Checked in init()
}

func (p *partialCall) ToFunction() interface{} {
	switch p.c.NumArgs() - p.n {
	case 0:
		return func(context *Context) (interface{}, error) {
			return Call0(p.c, context)
		}
	case 1:
		return func(arg0 interface{}, context *Context) (interface{}, error) {
			return Call1(arg0, p.c, context)
		}
	case 2:
		return func(arg0 interface{}, arg1 interface{}, context *Context) (interface{}, error) {
			return Call2(arg0, arg1, p.c, context)
		}
	default:
		panic(fmt.Sprintf("can't wrap partial with %d args", p.c.NumArgs()-p.n))
	}
}

func (p *partialCall) String() string {
	return fmt.Sprintf("partial/%d: %s", p.n, p.c.String())
}
