package functions

import (
	"fmt"
	"reflect"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"

	"github.com/golang/geo/s2"
)

func TestEvaluate(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	e := `find (intersecting (find-area /area/openstreetmap.org/way/115912092))`
	if _, err := api.EvaluateString(e, NewContext(granarySquare)); err != nil {
		t.Error(err)
	}
}

func TestMap(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	e := `find (intersecting (find-area /area/openstreetmap.org/way/222021576)) | map {f -> get f "name"}`
	if result, err := api.EvaluateString(e, NewContext(granarySquare)); err != nil {
		t.Error(err)
		return
	} else {
		values := []string{}
		if err := api.FillSliceFromValues(result.(api.Collection), &values); err != nil {
			t.Errorf("Expected no error, found %q", err)
		} else {
			expected := []string{
				"Caravan", "", "Yumchaa", "", "", "", "", "",
			}
			if !reflect.DeepEqual(expected, values) {
				t.Errorf("Expected %q, found %q", expected, values)
			}
		}
	}
}

func TestMapWithPartialFunction(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	e := `find (intersecting (find-area /area/openstreetmap.org/way/222021576)) | map (get "name")`
	if result, err := api.EvaluateString(e, NewContext(granarySquare)); err != nil {
		t.Errorf("Expected no error, found %q", err)
		return
	} else {
		values := []string{}
		if err := api.FillSliceFromValues(result.(api.Collection), &values); err != nil {
			t.Errorf("Expected no error, found %q", err)
		} else {
			expected := []string{
				"Caravan", "", "Yumchaa", "", "", "", "", "",
			}
			if !reflect.DeepEqual(expected, values) {
				t.Errorf("Expected %q, found %q", expected, values)
			}
		}
	}
}

func TestMapItems(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	e := `all-tags /n/6082053666 | map-items {p -> pair (second p) 1}`
	if result, err := api.EvaluateString(e, NewContext(granarySquare)); err != nil {
		t.Error(err)
	} else {
		filled := make(map[b6.Tag]int)
		collection, ok := result.(api.Collection)
		if !ok {
			t.Errorf("Expected a collection, found %T", result)
			return
		}
		if err := api.FillMap(collection, filled); err != nil {
			t.Error(err)
			return
		}
		if len(filled) < 2 {
			t.Errorf("Expected at least two values, found %d", len(filled))
		}
		if count, ok := filled[b6.Tag{Key: "#amenity", Value: "cafe"}]; !ok || count != 1 {
			t.Errorf("Expected a count of 1 for amenity tag, found %+v", filled)
		}
	}
}

func TestPipelineInLamba(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	e := `find [#building] | map {b -> area b | gt 1000.0} | count-values`
	if result, err := api.EvaluateString(e, NewContext(granarySquare)); err != nil {
		t.Error(err)
	} else {
		collection, ok := result.(api.Collection)
		if !ok {
			t.Errorf("Expected a collection, found %T", result)
			return
		}
		filled := make(map[bool]int)
		if err := api.FillMap(collection, filled); err != nil {
			t.Error(err)
			return
		}
		if filled[true] < filled[false] {
			t.Errorf("Expected more buildings over 1000m2 than not")
		}
		total := 0
		for _, n := range filled {
			total += n
		}
		if total != camden.BuildingsInGranarySquare {
			t.Errorf("Expected values for %d buildings, found %d", camden.BuildingsInGranarySquare, total)
		}
	}
}

func TestReturnFunctions(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}

	e := `find (keyed "#building") | to-geojson-collection | map-geometries (apply-to-area {a -> centroid a})`
	if v, err := api.EvaluateString(e, NewContext(granarySquare)); err != nil {
		t.Error(err)
		return
	} else {
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
}

func TestPassSpecificFunctionWhereGenericNeeded(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}
	e := `find (keyed "#building") | map centroid`
	if _, err := api.EvaluateString(e, NewContext(granarySquare)); err != nil {
		t.Error(err)
		return
	}
}

func TestInterfaceConversionForArguments(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}
	// find-feature returns a b6.Feature (that happens to implement
	// b6.Area), are area expects a b6.Area
	e := `find-feature /area/openstreetmap.org/way/427900370 | area`
	if v, err := api.EvaluateString(e, NewContext(granarySquare)); err != nil {
		t.Error(err)
		return
	} else {
		if f, ok := v.(float64); ok {
			if f < 300 || f > 400 {
				t.Errorf("Area outside expected range")
			}
		} else {
			t.Errorf("Expected a float, found %T", v)
		}
	}
}

func TestConvertQueryToFunctionReturningBool(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}
	e := `find [#amenity] | filter [addr:postcode]`
	if v, err := api.EvaluateString(e, NewContext(granarySquare)); err != nil {
		t.Error(err)
		return
	} else {
		n := 0
		i := v.(api.Collection).Begin()
		for {
			ok, err := i.Next()
			if err != nil {
				t.Error(err)
				return
			} else if !ok {
				break
			}
			if !i.Value().(b6.Feature).Get("addr:postcode").IsValid() {
				t.Errorf("Expected a addr:postcode tag, found none")
			}
			n++
		}
		if n == 0 {
			t.Errorf("Expected at least 1 value")
		}
	}
}

func TestConvertQueryToFunctionReturningBoolWithSpecificFeature(t *testing.T) {
	granarySquare := camden.BuildGranarySquareForTests(t)
	if granarySquare == nil {
		return
	}
	apply := func(f func(b6.PointFeature, *api.Context) (bool, error), c *api.Context) (bool, error) {
		return f(b6.FindPointByID(ingest.FromOSMNodeID(camden.VermuteriaNode), c.World), c)
	}
	c := &api.Context{
		World: granarySquare,
		FunctionSymbols: api.FunctionSymbols{
			"apply-to-example-point": apply,
		},
		FunctionWrappers: Wrappers(),
	}
	e := "apply-to-example-point [#amenity]"
	if v, err := api.EvaluateString(e, c); err != nil {
		t.Error(err)
	} else if !v.(bool) {
		t.Error("Expected true, found false")
	}
}

func TestConvertIntAndFloat64ToNumber(t *testing.T) {
	w := ingest.NewBasicMutableWorld()
	increment := func(n api.Number, c *api.Context) (int, error) {
		switch n := n.(type) {
		case api.IntNumber:
			return int(n) + 1, nil
		case api.FloatNumber:
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
		if v, err := api.EvaluateString(e, c); err != nil {
			t.Error(err)
		} else if v, ok := v.(int); !ok || v != 2 {
			t.Errorf("Expected 2, found %v", v)
		}
	}
}

func TestCallLambdaWithNoArguments(t *testing.T) {
	w := ingest.NewBasicMutableWorld()
	c := &api.Context{
		World: w,
		FunctionSymbols: api.FunctionSymbols{
			"call": func(f func(*api.Context) (interface{}, error), c *api.Context) (interface{}, error) {
				return f(c)
			},
		},
		FunctionWrappers: Wrappers(),
	}
	e := "call {-> 42}"
	if v, err := api.EvaluateString(e, c); err != nil {
		t.Error(err)
	} else if i, ok := v.(int64); !ok || i != 42 {
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
		t.Errorf("Expected an error")
	}
}

func TestReturnAnErrorFromALambda(t *testing.T) {
	w := ingest.NewBasicMutableWorld()
	c := &api.Context{
		World: w,
		FunctionSymbols: api.FunctionSymbols{
			"call": func(f func(*api.Context) (interface{}, error), c *api.Context) (interface{}, error) {
				return f(c)
			},
			"broken": func(_ int, c *api.Context) (interface{}, error) {
				return nil, fmt.Errorf("broken")
			},
		},
		FunctionWrappers: Wrappers(),
	}
	e := "call {-> broken 42}"
	r, err := api.EvaluateString(e, c)
	if r != nil || err == nil || err.Error() != "broken" {
		t.Errorf("Expected an error")
	}
}
