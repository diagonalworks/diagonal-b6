package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/exp/slices"
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
	base := camden.BuildGranarySquareForTests(t)
	w := ingest.NewMutableOverlayWorld(base)

	handler := StackHandler{
		UI: NewDefaultUI(w),
	}

	url := fmt.Sprintf("http://b6.diagonal.works/blocks?e=%s", url.QueryEscape(e))
	request := httptest.NewRequest("GET", url, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	result := response.Result()
	if result.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, found %d", http.StatusOK, result.StatusCode)
	}

	var uiResponse UIResponseJSON
	d := json.NewDecoder(result.Body)
	if err := d.Decode(&uiResponse); err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

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
