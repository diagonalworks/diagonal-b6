package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"reflect"
	"sort"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/api/functions"
)

type Interface struct {
	Name       string
	Implements []string
}

type Function struct {
	Name       string
	ArgTypes   []string
	ArgNames   []string
	Result     string
	IsVariadic bool
	Doc        string
}

type Collection struct {
	Name  string
	Key   string
	Value string
}

type API struct {
	Version      string
	Interfaces   []Interface
	Functions    []Function
	FunctionArgs []Function
	Collections  []Collection
}

func nameForType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return nameForType(t.Elem())
	} else if t == AnyType {
		return "Any"
	} else if t == UntypedCollectionType {
		return "Collection[Any,Any]"
	} else if t.Implements(UntypedCollectionType) && t != CollectionFeatureType {
		return collectionForType(t).Name
	} else if t.Kind() == reflect.Func {
		name := "Function"
		for i := 1; i < t.NumIn(); i++ {
			name += strings.Title(nameForType(t.In(i)))
		}
		name += strings.Title(nameForType(t.Out(0)))
		return name
	} else if i := strings.LastIndex(t.Name(), "."); i >= 0 {
		return t.Name()[i+1:]
	} else if t.Name() == "" {
		panic(fmt.Sprintf("no name for %+v", t))
	}
	return t.Name()
}

func isBuiltin(t reflect.Type) bool {
	var a any
	return t.Kind() == reflect.Int || t.Kind() == reflect.Float64 || t.Kind() == reflect.String || t == reflect.TypeOf(a)
}

func collectionForType(t reflect.Type) Collection {
	if t == UntypedCollectionType {
		return Collection{
			Name:  "Collection[Any,Any]",
			Key:   "Any",
			Value: "Any",
		}
	}
	begin, ok := t.MethodByName("Begin")
	if !ok {
		panic(fmt.Sprintf("No begin for collection %s", t))
	}
	key, ok := begin.Type.Out(0).MethodByName("Key")
	if !ok {
		panic(fmt.Sprintf("No key for collection %s", t))
	}
	value, ok := begin.Type.Out(0).MethodByName("Value")
	if !ok {
		panic(fmt.Sprintf("No value for collection %s", t))
	}
	return Collection{
		Name:  fmt.Sprintf("Collection[%s,%s]", strings.Title(nameForType(key.Type.Out(0))), strings.Title(nameForType(value.Type.Out(0)))),
		Key:   nameForType(key.Type.Out(0)),
		Value: nameForType(value.Type.Out(0)),
	}
}

var AnyType = reflect.TypeOf((*interface{})(nil)).Elem()
var UntypedCollectionType = reflect.TypeOf((*b6.UntypedCollection)(nil)).Elem()
var CollectionFeatureType = reflect.TypeOf((*b6.CollectionFeature)(nil)).Elem()

func outputFunctions() error {
	var output API
	var err error
	output.Version, err = b6.AdvanceVersionFromGit()
	if err != nil {
		return err
	}

	docs := functions.FunctionDocs()

	types := make(map[reflect.Type]struct{})
	for name, f := range functions.Functions() {
		t := reflect.TypeOf(f)
		if t.Kind() == reflect.Func {
			ff := Function{Name: name, ArgTypes: []string{}}
			if doc, ok := docs[name]; ok {
				ff.ArgNames = doc.ArgNames
				ff.Doc = doc.Doc
			}
			for i := 1; i < t.NumIn(); i++ {
				var in reflect.Type
				if i == t.NumIn()-1 && t.IsVariadic() {
					in = t.In(i).Elem()
					ff.IsVariadic = true
				} else {
					in = t.In(i)
				}
				ff.ArgTypes = append(ff.ArgTypes, nameForType(in))
				types[in] = struct{}{}
			}
			ff.Result = nameForType(t.Out(0))
			types[t.Out(0)] = struct{}{}
			output.Functions = append(output.Functions, ff)
		}
	}

	for t := range types {
		if t.Implements(UntypedCollectionType) && t != CollectionFeatureType {
			output.Collections = append(output.Collections, collectionForType(t))
		} else if t.Kind() == reflect.Func {
			f := Function{Name: nameForType(t), ArgTypes: []string{}}
			for i := 1; i < t.NumIn(); i++ {
				f.ArgTypes = append(f.ArgTypes, nameForType(t.In(i)))
			}
			f.Result = nameForType(t.Out(0))
			f.IsVariadic = t.IsVariadic()
			output.FunctionArgs = append(output.FunctionArgs, f)
		} else if !isBuiltin(t) {
			ts := make([]reflect.Type, 0)
			for tt := range types {
				// TODO: Handle Collections and Values properly when we sort out the interface
				if t != tt && tt != AnyType && tt.Kind() == reflect.Interface && t.Implements(tt) {
					ts = append(ts, tt)
				}
			}
			for i := range ts {
				for j := i + 1; j < len(ts); j++ {
					if ts[i] != nil && ts[j] != nil {
						if ts[i].Implements(ts[j]) {
							ts[j] = nil
						} else if ts[j].Implements(ts[i]) {
							ts[i] = nil
						}
					}
				}
			}
			i := Interface{Name: nameForType(t), Implements: []string{}}
			for _, tt := range ts {
				if tt != nil {
					i.Implements = append(i.Implements, nameForType(tt))
				}
			}
			output.Interfaces = append(output.Interfaces, i)
		}
	}
	b, err := json.MarshalIndent(&output, "", "  ")
	if err == nil {
		os.Stdout.Write(b)
		os.Stderr.Write([]byte{'\n'})
	}
	return err
}

const SourceRoot = "src/diagonal.works/b6/api/functions"

func outputDocs() error {
	noTests := func(f fs.FileInfo) bool {
		return !strings.HasSuffix(f.Name(), "_test.go")
	}
	files := token.NewFileSet()
	packages, err := parser.ParseDir(files, SourceRoot, noTests, parser.ParseComments)
	if err != nil {
		return err
	}
	fs := make(map[string]string)
	fillExportedFunctionNames(packages, fs)

	docs := make(map[string]functions.Doc)
	fillFunctionDocs(packages, docs)

	fmt.Println("package functions")
	fmt.Println("")
	fmt.Println("// Code generated by b6-api. DO NOT EDIT.")
	fmt.Println("")
	fmt.Println("var functionDocs = map[string]Doc{")

	b6Names := make([]string, 0, len(fs))
	for b6Name := range fs {
		b6Names = append(b6Names, b6Name)
	}
	sort.Strings(b6Names)

	for _, b6Name := range b6Names {
		if doc, ok := docs[fs[b6Name]]; ok {
			fmt.Printf("\t%q: %s,\n", b6Name, doc.ToGoLiteral())
		}
	}
	fmt.Println("}")
	return nil
}

func fillExportedFunctionNames(packages map[string]*ast.Package, functions map[string]string) {
	for _, p := range packages {
		for _, file := range p.Files {
			for _, decl := range file.Decls {
				if g, ok := decl.(*ast.GenDecl); ok && g.Tok == token.VAR {
					if v, ok := g.Specs[0].(*ast.ValueSpec); ok {
						if v.Names[0].Name == "functions" && len(v.Values) == 1 {
							fillMapFromExpr(v.Values[0], functions)
						}
					}
				}
			}
		}
	}
}

func fillMapFromExpr(expr ast.Expr, m map[string]string) {
	if c, ok := expr.(*ast.CompositeLit); ok {
		for _, e := range c.Elts {
			if kv, ok := e.(*ast.KeyValueExpr); ok {
				if k, ok := kv.Key.(*ast.BasicLit); ok && k.Kind == token.STRING {
					if v, ok := kv.Value.(*ast.Ident); ok {
						m[k.Value[1:len(k.Value)-1]] = v.Name
					}
				}
			}
		}
	}
}

func fillFunctionDocs(packages map[string]*ast.Package, docs map[string]functions.Doc) {
	for _, p := range packages {
		for _, file := range p.Files {
			for _, decl := range file.Decls {
				if f, ok := decl.(*ast.FuncDecl); ok {
					var doc functions.Doc
					for _, field := range f.Type.Params.List {
						if len(field.Names) > 0 {
							doc.ArgNames = append(doc.ArgNames, field.Names[0].Name)
						} else {
							doc.ArgNames = append(doc.ArgNames, "")
						}
					}
					if len(doc.ArgNames) > 0 {
						doc.ArgNames = doc.ArgNames[1:] // Remove *api.Context
					}
					if f.Doc != nil {
						doc.Doc = f.Doc.Text()
					}
					docs[f.Name.Name] = doc
				}
			}
		}
	}
}

func main() {
	versionFlag := flag.Bool("version", false, "Output a package version based on the API version and git.")
	pipVersionFlag := flag.Bool("pip-version", false, "Like --version, but formatted for PIP.")
	outputFuctionsFlag := flag.Bool("functions", false, "Output function definitions")
	outputDocsFlag := flag.Bool("docs", false, "Output arg names")
	flag.Parse()

	var err error
	if *versionFlag {
		var v string
		v, err = b6.AdvanceVersionFromGit()
		if err == nil {
			fmt.Fprintf(os.Stdout, "%s\n", v)
		}
	} else if *pipVersionFlag {
		var v string
		v, err = b6.AdvancePythonVersionFromGit()
		if err == nil {
			fmt.Fprintf(os.Stdout, "%s\n", v)
		}
	} else if *outputFuctionsFlag {
		err = outputFunctions()
	} else if *outputDocsFlag {
		err = outputDocs()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
