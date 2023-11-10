package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
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
	Args       []string
	Result     string
	IsVariadic bool
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
		return "AnyAnyCollection"
	} else if t.Implements(UntypedCollectionType) {
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
			Name:  "AnyAnyCollection",
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
		Name:  fmt.Sprintf("%s%sCollection", strings.Title(nameForType(key.Type.Out(0))), strings.Title(nameForType(value.Type.Out(0)))),
		Key:   nameForType(key.Type.Out(0)),
		Value: nameForType(value.Type.Out(0)),
	}
}

var AnyType = reflect.TypeOf((*interface{})(nil)).Elem()
var UntypedCollectionType = reflect.TypeOf((*b6.UntypedCollection)(nil)).Elem()

func generateAPI() error {
	var output API
	var err error
	output.Version, err = b6.AdvanceVersionFromGit()
	if err != nil {
		return err
	}

	types := make(map[reflect.Type]struct{})
	for name, f := range functions.Functions() {
		t := reflect.TypeOf(f)
		if t.Kind() == reflect.Func {
			ff := Function{Name: name, Args: []string{}}
			for i := 1; i < t.NumIn(); i++ {
				var in reflect.Type
				if i == t.NumIn()-1 && t.IsVariadic() {
					in = t.In(i).Elem()
					ff.IsVariadic = true
				} else {
					in = t.In(i)
				}
				ff.Args = append(ff.Args, nameForType(in))
				types[in] = struct{}{}
			}
			ff.Result = nameForType(t.Out(0))
			types[t.Out(0)] = struct{}{}
			output.Functions = append(output.Functions, ff)
		}
	}

	for t := range types {
		if t.Implements(UntypedCollectionType) {
			output.Collections = append(output.Collections, collectionForType(t))
		} else if t.Kind() == reflect.Func {
			f := Function{Name: nameForType(t), Args: []string{}}
			for i := 1; i < t.NumIn(); i++ {
				f.Args = append(f.Args, nameForType(t.In(i)))
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

func main() {
	version := flag.Bool("version", false, "Output a package version based on the API version and git.")
	pipVersion := flag.Bool("pip-version", false, "Like --version, but formatted for PIP.")
	flag.Parse()

	var err error
	if *version {
		var v string
		v, err = b6.AdvanceVersionFromGit()
		if err == nil {
			fmt.Fprintf(os.Stdout, "%s\n", v)
		}
	} else if *pipVersion {
		var v string
		v, err = b6.AdvancePythonVersionFromGit()
		if err == nil {
			fmt.Fprintf(os.Stdout, "%s\n", v)
		}
	} else {
		err = generateAPI()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
