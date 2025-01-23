package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"sync"
	"testing"

	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestMatchingFunctions(t *testing.T) {
	response := sendExpressionToTestUI("find-feature /a/427900370", t)
	functions := make([]string, 0)
	for _, s := range response.Proto.Stack.Substacks {
		for _, l := range s.Lines {
			if shell := l.GetShell(); shell != nil {
				for _, f := range shell.Functions {
					functions = append(functions, f)
				}
			}
		}
	}

	for _, expected := range []string{"to-geojson", "closest", "get-string", "reachable"} {
		if !slices.Contains(functions, expected) {
			t.Errorf("Function %q not included in area features: %v", expected, functions)
		}
	}
}

func sendExpressionToTestUI(e string, t *testing.T) *UIResponseJSON {
	w := &ingest.MutableWorlds{
		Base: camden.BuildGranarySquareForTests(t),
	}

	var lock sync.RWMutex
	handler := StackHandler{
		UI: &OpenSourceUI{
			Worlds:          w,
			FunctionSymbols: functions.Functions(),
			Evaluator: api.Evaluator{
				Worlds:          w,
				FunctionSymbols: functions.Functions(),
				Adaptors:        functions.Adaptors(),
				Options: api.Options{
					Cores: 2,
				},
				Lock: &lock,
			},
			Lock: &lock,
		},
	}

	url := fmt.Sprintf("http://b6.diagonal.works/stack?e=%s", url.QueryEscape(e))
	request := httptest.NewRequest("GET", url, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	result := response.Result()
	if result.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, found %d", http.StatusOK, result.StatusCode)
	}

	// NOTE: This is a hack around Go's ... interesting ... json capabilities.
	// We need to provide a "decodable" version of `UIResponseJSON`; which is
	// itself not decodable properly as one of the fields is an interface, not a
	// concrete type.
	//
	// Ideally, we would just decode the `UIResponseJSON` type itself; but it is
	// not decodable because of the interface (`geojson.GeoJSON`) field. As a
	// result, we just define a new thing, decode it into that, and set that
	// field of an actual `UIResponseJSON`; leaving the other part of that
	// object undefined (!!). As this is only a test, it doesn't actually
	// matter.
	type UIResponseJSON_Decoadable struct {
		Proto *UIResponseProtoJSON `json:"proto,omitempty"`
	}

	var decodable UIResponseJSON_Decoadable
	d := json.NewDecoder(result.Body)
	if err := d.Decode(&decodable); err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

	var uiResponse UIResponseJSON
	uiResponse.Proto = decodable.Proto
	// NOTE: uiResponse.GeoJSON here is <nil>; but we don't actually look at
	// that in the tests; so we don't worry about it.

	return &uiResponse
}

func TestHistogramWithTagKeys(t *testing.T) {
	response := sendExpressionToTestUI(`find [#building] | map (get "#building") | histogram`, t)

	atoms := make([]string, 0)
	for _, s := range response.Proto.Stack.Substacks {
		for _, l := range s.Lines {
			if histogramBar := l.GetHistogramBar(); histogramBar != nil {
				if value := histogramBar.Range.GetValue(); value != "" {
					atoms = append(atoms, value)
				}
			}
		}
	}

	expected := []string{
		"#building=yes",
		"#building=university",
		"#building=commercial",
		"#building=apartments",
		"#building=construction",
	}
	less := func(a, b string) bool { return a < b }
	if diff := cmp.Diff(expected, atoms, cmpopts.SortSlices(less)); diff != "" {
		t.Errorf("Found diff (-want, +got):\n%s", diff)
	}
}
