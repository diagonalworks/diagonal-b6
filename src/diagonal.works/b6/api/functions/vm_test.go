package functions

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"

	"github.com/golang/geo/s2"
	"github.com/google/go-cmp/cmp"
)

func TestEvaluate(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	e := `find (intersecting (find-area /area/openstreetmap.org/way/115912092))`
	if _, err := api.EvaluateString(e, NewContext(granarySquare)); err != nil {
		t.Error(err)
	}
}

func TestMapWithVM(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	e := `find (intersecting (find-area /area/openstreetmap.org/way/222021576)) | map {f -> get f "name"}`
	result, err := api.EvaluateString(e, NewContext(granarySquare))
	if err != nil {
		t.Fatal(err)
	}
	values := []string{}

	if err := api.FillSliceFromValues(result.(b6.UntypedCollection), &values); err != nil {
		t.Fatalf("Expected no error, found %q", err)
	}
	expected := []string{
		"Caravan", "", "Yumchaa", "", "", "", "", "",
	}
	if !reflect.DeepEqual(expected, values) {
		t.Errorf("Expected %q, found %q", expected, values)
	}
}

func TestMapParallelWithVM(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	e := `find (intersecting (find-area /area/openstreetmap.org/way/222021576)) | map-parallel {f -> get f "name"}`
	result, err := api.EvaluateString(e, NewContext(granarySquare))
	if err != nil {
		t.Fatal(err)
	}
	values := []string{}
	if err := api.FillSliceFromValues(result.(b6.UntypedCollection), &values); err != nil {
		t.Fatalf("Expected no error, found %q", err)
	}
	expected := []string{
		"Caravan", "", "Yumchaa", "", "", "", "", "",
	}
	if !reflect.DeepEqual(expected, values) {
		t.Errorf("Expected %q, found %q", expected, values)
	}
}

func TestMapWithVMAndPartialFunction(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	e := `find (intersecting (find-area /area/openstreetmap.org/way/222021576)) | map (get "name")`
	result, err := api.EvaluateString(e, NewContext(granarySquare))
	if err != nil {
		t.Fatalf("Expected no error, found %q", err)
	}
	values := []string{}
	if err := api.FillSliceFromValues(result.(b6.UntypedCollection), &values); err != nil {
		t.Fatalf("Expected no error, found %q", err)
	}
	expected := []string{
		"Caravan", "", "Yumchaa", "", "", "", "", "",
	}
	if !reflect.DeepEqual(expected, values) {
		t.Errorf("Expected %q, found %q", expected, values)
	}
}

func TestMapItems(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	e := `all-tags /n/6082053666 | map-items {p -> pair (second p) 1}`
	result, err := api.EvaluateString(e, NewContext(granarySquare))
	if err != nil {
		t.Fatal(err)
	}
	filled := make(map[b6.Tag]int)
	collection, ok := result.(b6.UntypedCollection)
	if !ok {
		t.Fatalf("Expected a collection, found %T", result)
	}
	if err := api.FillMap(collection, filled); err != nil {
		t.Fatal(err)
	}
	if len(filled) < 2 {
		t.Errorf("Expected at least two values, found %d", len(filled))
	}
	if count, ok := filled[b6.Tag{Key: "#amenity", Value: b6.NewStringExpression("cafe")}]; !ok || count != 1 {
		t.Errorf("Expected a count of 1 for amenity tag, found %+v", filled)
	}
}

func TestWithVMAndPipelineInLamba(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	e := `find [#building] | map {b -> area b | gt 1000.0} | count-values`
	result, err := api.EvaluateString(e, NewContext(granarySquare))
	if err != nil {
		t.Fatal(err)
	}
	collection, ok := result.(b6.UntypedCollection)
	if !ok {
		t.Fatalf("Expected a collection, found %T", result)
	}
	filled := make(map[bool]int)
	if err := api.FillMap(collection, filled); err != nil {
		t.Fatal(err)
	}
	if filled[true] < filled[false] {
		t.Error("Expected more buildings over 1000m2 than not")
	}
	total := 0
	for _, n := range filled {
		total += n
	}
	if total != camden.BuildingsInGranarySquare {
		t.Errorf("Expected values for %d buildings, found %d", camden.BuildingsInGranarySquare, total)
	}
}

func TestReturnFunctions(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	e := `find (keyed "#building") | to-geojson-collection | map-geometries (apply-to-area {a -> centroid a})`
	v, err := api.EvaluateString(e, NewContext(granarySquare))
	if err != nil {
		t.Fatal(err)
	}
	if g, ok := v.(*geojson.FeatureCollection); ok {
		cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLngFromDegrees(51.53531, -0.12521)), b6.MetersToAngle(1000))
		for _, f := range g.Features {
			if p, ok := f.Geometry.Coordinates.(geojson.Point); ok {
				if !cap.ContainsPoint(p.ToS2Point()) {
					t.Errorf("Expected %s to be within Granary Square", p.ToS2LatLng())
				}
			} else {
				t.Errorf("Expected geojson.Point, found %T", f.Geometry.Coordinates)
			}
		}
	} else {
		t.Errorf("Expected geojson.FeatureCollection, found %T", v)
	}
}

func TestPassSpecificFunctionWhereGenericNeeded(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	e := `find (keyed "#building") | map centroid`
	if _, err := api.EvaluateString(e, NewContext(granarySquare)); err != nil {
		t.Error(err)
	}
}

func TestInterfaceConversionForArguments(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)

	// find-feature returns a b6.Feature (that happens to implement
	// b6.Area), are area expects a b6.Area
	e := `find-feature /area/openstreetmap.org/way/427900370 | area`
	v, err := api.EvaluateString(e, NewContext(granarySquare))
	if err != nil {
		t.Fatal(err)
	}
	f, ok := v.(float64)
	if !ok {
		t.Fatalf("Expected a float, found %T", v)
	}
	if f < 300 || f > 400 {
		t.Error("Area outside expected range")
	}
}

func TestConvertQueryToFunctionReturningBool(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	e := `find [#amenity] | filter [addr:postcode]`
	v, err := api.EvaluateString(e, NewContext(granarySquare))
	if err != nil {
		t.Fatal(err)

	}
	n := 0
	i := v.(b6.UntypedCollection).BeginUntyped()
	for {
		ok, err := i.Next()
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			break
		}
		if !i.Value().(b6.Feature).Get("addr:postcode").IsValid() {
			t.Error("Expected a addr:postcode tag, found none")
		}
		n++
	}
	if n == 0 {
		t.Error("Expected at least 1 value")
	}
}

func TestConvertQueryToFunctionReturningBoolWithSpecificFeature(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	apply := func(c *api.Context, f func(*api.Context, b6.PhysicalFeature) (bool, error)) (bool, error) {
		return f(c, c.World.FindFeatureByID(ingest.FromOSMNodeID(camden.VermuteriaNode)).(b6.PhysicalFeature))
	}
	c := &api.Context{
		World: granarySquare,
		FunctionSymbols: api.FunctionSymbols{
			"apply-to-example-point": apply,
		},
		Adaptors: Adaptors(),
	}
	e := "apply-to-example-point [#amenity]"
	v, err := api.EvaluateString(e, c)
	if err != nil {
		t.Fatal(err)
	}
	if !v.(bool) {
		t.Error("Expected true, found false")
	}
}

func TestConvertIntAndFloat64ToNumber(t *testing.T) {
	w := ingest.NewBasicMutableWorld()
	increment := func(c *api.Context, n b6.Number) (int, error) {
		switch n := n.(type) {
		case b6.IntNumber:
			return int(n) + 1, nil
		case b6.FloatNumber:
			return int(n) + 1, nil
		}
		return 0, fmt.Errorf("Bad number")
	}
	c := &api.Context{
		World: w,
		FunctionSymbols: api.FunctionSymbols{
			"increment": increment,
		},
	}
	for _, e := range []string{"increment 1", "increment 1.0"} {
		v, err := api.EvaluateString(e, c)
		if err != nil {
			t.Fatal(err)
		}
		if v, ok := v.(int); !ok || v != 2 {
			t.Errorf("Expected 2, found %v", v)
		}
	}
}

func TestCallLambdaWithNoArguments(t *testing.T) {
	w := ingest.NewBasicMutableWorld()
	c := &api.Context{
		World: w,
		FunctionSymbols: api.FunctionSymbols{
			"call": func(c *api.Context, f func(*api.Context) (interface{}, error)) (interface{}, error) {
				return f(c)
			},
		},
		Adaptors: Adaptors(),
	}
	e := "call {-> 42}"
	v, err := api.EvaluateString(e, c)
	if err != nil {
		t.Fatal(err)
	}
	if i, ok := v.(int); !ok || i != 42 {
		t.Errorf("Expected 42, found %T %v", v, v)
	}
}

func TestIncorrectlyPassNumberAsFunction(t *testing.T) {
	w := ingest.NewBasicMutableWorld()
	c := &api.Context{
		World: w,
		FunctionSymbols: api.FunctionSymbols{
			"call": func(f func(*api.Context) (interface{}, error), c *api.Context) (interface{}, error) {
				return f(c)
			},
		},
	}
	e := "call 42"
	_, err := api.EvaluateString(e, c)
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestReturnAnErrorFromALambda(t *testing.T) {
	w := ingest.NewBasicMutableWorld()
	c := &api.Context{
		World: w,
		FunctionSymbols: api.FunctionSymbols{
			"call": func(c *api.Context, f func(*api.Context) (interface{}, error)) (interface{}, error) {
				return f(c)
			},
			"broken": func(c *api.Context, _ int) (interface{}, error) {
				return nil, fmt.Errorf("broken")
			},
		},
		Adaptors: Adaptors(),
	}
	e := "call {-> broken 42}"
	r, err := api.EvaluateString(e, c)
	if r != nil || err == nil || err.Error() != "broken" {
		t.Errorf("Expected an error, found: %+v", err)
	}
}

func TestMapLiteralCollection(t *testing.T) {
	w := ingest.NewBasicMutableWorld()
	e := `map {highway="motorway": 2, highway="primary": 6} (add 1)`
	result, err := api.EvaluateString(e, NewContext(w))
	if err != nil {
		t.Fatal(err)
	}
	collection := make(map[b6.Tag]int)
	if err := api.FillMap(result.(b6.UntypedCollection), collection); err != nil {
		t.Fatalf("Expected no error, found %q", err)
	}
	expected := map[b6.Tag]int{
		{Key: "highway", Value: b6.NewStringExpression("motorway")}: 3,
		{Key: "highway", Value: b6.NewStringExpression("primary")}:  7,
	}
	if diff := cmp.Diff(expected, collection); diff != "" {
		t.Errorf("Found diff (-want, +got):\n%s", diff)
	}
}

func TestMapLiteralCollectionWithImplicitKeys(t *testing.T) {
	w := ingest.NewBasicMutableWorld()
	e := `map {36, 42} (add 10)`
	result, err := api.EvaluateString(e, NewContext(w))
	if err != nil {
		t.Fatal(err)
	}
	collection := make(map[int]int)
	if err := api.FillMap(result.(b6.UntypedCollection), collection); err != nil {
		t.Fatalf("Expected no error, found %q", err)
	}
	expected := map[int]int{
		0: 46,
		1: 52,
	}
	if diff := cmp.Diff(expected, collection); diff != "" {
		t.Errorf("Found diff (-want, +got):\n%s", diff)
	}
}

func TestCallNestedLambda(t *testing.T) {
	w := ingest.NewBasicMutableWorld()
	e := `call {a -> call {b -> add a b} 10} 20`
	parsed, err := api.ParseExpression(e)
	if err != nil {
		t.Fatalf("Failed to parse expression: %s", err)
	}
	r, err := api.Evaluate(parsed, NewContext(w))
	if err != nil {
		t.Fatal(err)
	}
	if i, ok := b6.ToInt(r); !ok || i != 30 {
		t.Errorf("Unexpected result")
	}
}

func TestVMProvidesCurrentExpression(t *testing.T) {
	var expression b6.Expression
	c := &api.Context{
		World: ingest.NewBasicMutableWorld(),
		FunctionSymbols: api.FunctionSymbols{
			"sub": func(c *api.Context, a int, b int) (int, error) {
				expression = c.VM.Expression()
				return a - b, nil
			},
			"add": func(c *api.Context, a int, b int) (int, error) {
				return a + b, nil
			},
		},
		Adaptors: Adaptors(),
	}
	e := "sub (add 20 30) (add 10 20)"
	_, err := api.EvaluateString(e, c)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	expected := b6.NewCallExpression(
		b6.NewSymbolExpression("sub"),
		[]b6.Expression{
			b6.NewCallExpression(
				b6.NewSymbolExpression("add"),
				[]b6.Expression{
					b6.NewIntExpression(20),
					b6.NewIntExpression(30),
				},
			),
			b6.NewCallExpression(
				b6.NewSymbolExpression("add"),
				[]b6.Expression{
					b6.NewIntExpression(10),
					b6.NewIntExpression(20),
				},
			),
		},
	)

	if !expression.Equal(expected) {
		t.Errorf("expected: %s, found: %s", expected, expression)
	}
}

func TestVMProvidesCurrentExpressionWithMap(t *testing.T) {
	var expression b6.Expression
	fs := Functions()
	fs["sub"] = func(c *api.Context, a int, b int) (int, error) {
		expression = c.VM.Expression()
		return a - b, nil
	}

	c := &api.Context{
		World:           ingest.NewBasicMutableWorld(),
		FunctionSymbols: fs,
		Adaptors:        Adaptors(),
		Context:         context.Background(),
	}

	e := "map (collection (pair 0 (add 20 30))) (sub 10)"
	mapped, err := api.EvaluateString(e, c)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	i := mapped.(b6.UntypedCollection).BeginUntyped()
	for {
		ok, err := i.Next()
		if err != nil {
			t.Fatalf("expected no error, found %s", err)
		} else if !ok {
			break
		}
	}

	expected, err := api.ParseExpression("sub 10 (second (pair 0 (add 20 30))) 10")
	if err != nil {
		t.Fatalf("Failed to parse expected expression: %s", err)
	}

	expression = api.Simplify(expected, fs)

	if !expression.Equal(expected) {
		t.Errorf("expected: %s, found: %s", expected, expression)
	}
}

func TestVMProvidesCurrentExpressionWithPartialCalls(t *testing.T) {
	var expression b6.Expression
	fs := Functions()
	fs["sub"] = func(c *api.Context, a int, b int) (int, error) {
		expression = c.VM.Expression()
		return a - b, nil
	}

	c := &api.Context{
		World:           ingest.NewBasicMutableWorld(),
		FunctionSymbols: fs,
		Adaptors:        Adaptors(),
		Context:         context.Background(),
	}

	e := "add 20 30 | sub 10"
	_, err := api.EvaluateString(e, c)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	expected, err := api.ParseExpression("sub (add 20 30) 10")
	if err != nil {
		t.Fatalf("Failed to parse expected expression: %s", err)
	}

	if !expression.Equal(expected) {
		t.Errorf("expected: %s, found: %s", expected, expression)
	}
}

func TestCallExpressionFeature(t *testing.T) {
	w := ingest.NewBasicMutableWorld()
	id := b6.FeatureID{Type: b6.FeatureTypeExpression, Namespace: "diagonal.works/test", Value: 0}
	f := &ingest.GenericFeature{
		ID: id,
		Tags: []b6.Tag{{
			Key:   b6.ExpressionTag,
			Value: b6.NewCallExpression(b6.NewSymbolExpression("add"), []b6.Expression{b6.NewIntExpression(10)}),
		}},
	}
	if err := w.AddFeature(f); err != nil {
		t.Fatalf("Failed to add feature")
	}

	fs := Functions()
	fs["add"] = func(c *api.Context, a int, b int) (int, error) {
		return a + b, nil
	}

	c := &api.Context{
		World:           w,
		FunctionSymbols: fs,
		Adaptors:        Adaptors(),
		Context:         context.Background(),
	}

	r, err := api.EvaluateString("call (evaluate-feature /expression/diagonal.works/test/0) 20", c)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	if i, ok := r.(int); !ok || i != 30 {
		t.Errorf("Expected 30, found %T %v", r, r)
	}
}

func TestCallExpressionFeatureWithLambda(t *testing.T) {
	w := ingest.NewBasicMutableWorld()
	id := b6.FeatureID{Type: b6.FeatureTypeExpression, Namespace: "diagonal.works/test", Value: 0}
	f := &ingest.GenericFeature{
		ID: id,
		Tags: []b6.Tag{{
			Key: b6.ExpressionTag,
			Value: b6.NewLambdaExpression(
				[]string{"i"},
				b6.NewCallExpression(
					b6.NewSymbolExpression("add"),
					[]b6.Expression{
						b6.NewSymbolExpression("i"),
						b6.NewIntExpression(10),
					},
				),
			),
		}},
	}
	if err := w.AddFeature(f); err != nil {
		t.Fatalf("Failed to add feature")
	}

	fs := Functions()
	fs["add"] = func(c *api.Context, a int, b int) (int, error) {
		return a + b, nil
	}

	c := &api.Context{
		World:           w,
		FunctionSymbols: fs,
		Adaptors:        Adaptors(),
		Context:         context.Background(),
	}

	r, err := api.EvaluateString("call (evaluate-feature /expression/diagonal.works/test/0) 20", c)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	if i, ok := r.(int); !ok || i != 30 {
		t.Errorf("Expected 30, found %T %v", r, r)
	}
}
