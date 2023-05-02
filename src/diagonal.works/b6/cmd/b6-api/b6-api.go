package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
)

type Interface struct {
	Name       string
	Implements []string
}

type Function struct {
	Name   string
	Args   []string
	Result string
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
		return "any"
	} else if t.Kind() == reflect.Func {
		name := "Function"
		for i := 0; i < t.NumIn()-1; i++ {
			name += strings.Title(nameForType(t.In(i)))
		}
		name += strings.Title(nameForType(t.Out(0)))
		return name
	} else if i := strings.LastIndex(t.Name(), "."); i >= 0 {
		return t.Name()[i+1:]
	}
	return t.Name()
}

func isBuiltin(t reflect.Type) bool {
	return t.Kind() == reflect.Int || t.Kind() == reflect.Float64 || t.Kind() == reflect.String
}

func collectionForType(t reflect.Type) Collection {
	// TODO: Replace with generics
	switch n := nameForType(t); n {
	case "Collection":
		return Collection{Name: n, Key: "any", Value: "any"}
	case "PointCollection":
		return Collection{Name: n, Key: "any", Value: "Point"}
	case "PathCollection":
		return Collection{Name: n, Key: "any", Value: "Path"}
	case "AreaCollection":
		return Collection{Name: n, Key: "any", Value: "Area"}
	case "StringPointCollection":
		return Collection{Name: n, Key: "string", Value: "Point"}
	case "StringAreaCollection":
		return Collection{Name: n, Key: "string", Value: "Area"}
	case "FeatureCollection":
		return Collection{Name: n, Key: "any", Value: "Feature"}
	case "PointFeatureCollection":
		return Collection{Name: n, Key: "any", Value: "PointFeature"}
	case "PathFeatureCollection":
		return Collection{Name: n, Key: "any", Value: "PathFeature"}
	case "AreaFeatureCollection":
		return Collection{Name: n, Key: "any", Value: "AreaFeature"}
	case "RelationFeatureCollection":
		return Collection{Name: n, Key: "any", Value: "RelationFeature"}
	case "IntStringCollection":
		return Collection{Name: n, Key: "int", Value: "string"}
	case "IntTagCollection":
		return Collection{Name: n, Key: "int", Value: "Tag"}
	case "StringStringCollection":
		return Collection{Name: n, Key: "string", Value: "string"}
	case "FeatureIDIntCollection":
		return Collection{Name: n, Key: "FeatureID", Value: "int"}
	case "AnyFloatCollection":
		return Collection{Name: n, Key: "any", Value: "float64"}
	case "AnyRenderableCollection":
		return Collection{Name: n, Key: "any", Value: "Renderable"}
	case "FeatureIDStringCollection":
		return Collection{Name: n, Key: "FeatureID", Value: "string"}
	case "FeatureIDTagCollection":
		return Collection{Name: n, Key: "FeatureID", Value: "Tag"}
	case "FeatureIDStringStringPairCollection":
		return Collection{Name: n, Key: "FeatureID", Value: "StringStringPair"}
	case "FeatureIDFeatureIDCollection":
		return Collection{Name: n, Key: "FeatureID", Value: "FeatureID"}
	}
	panic(fmt.Sprintf("Can't handle collection %s", t))
}

var AnyType = reflect.TypeOf((*interface{})(nil)).Elem()
var CollectionType = reflect.TypeOf((*api.Collection)(nil)).Elem()

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
			for i := 0; i < t.NumIn()-1; i++ {
				ff.Args = append(ff.Args, nameForType(t.In(i)))
				types[t.In(i)] = struct{}{}
			}
			ff.Result = nameForType(t.Out(0))
			types[t.Out(0)] = struct{}{}
			output.Functions = append(output.Functions, ff)
		}
	}

	// Force inclusion of parent types that aren't used in functions
	types[reflect.TypeOf((*api.AreaCollection)(nil)).Elem()] = struct{}{}

	for t := range types {
		if t.Implements(CollectionType) {
			output.Collections = append(output.Collections, collectionForType(t))
		} else if t.Kind() == reflect.Func {
			f := Function{Name: nameForType(t), Args: []string{}}
			for i := 0; i < t.NumIn()-1; i++ {
				f.Args = append(f.Args, nameForType(t.In(i)))
			}
			f.Result = nameForType(t.Out(0))
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
