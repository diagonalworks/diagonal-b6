package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"diagonal.works/b6/api/functions"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/renderer"
	"diagonal.works/b6/test/camden"
)

func TestMatchingFunctions(t *testing.T) {
	base := camden.BuildGranarySquareForTests(t)
	w := ingest.NewMutableOverlayWorld(base)

	handler := UIHandler{
		World:            w,
		RenderRules:      renderer.BasemapRenderRules,
		Cores:            1,
		FunctionSymbols:  functions.Functions(),
		FunctionWrappers: functions.Wrappers(),
	}

	e := "find-feature /a/427900370"
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
	functions := make([]string, 0)
	for _, s := range uiResponse.Proto.Stack.Substacks {
		for _, l := range s.Lines {
			if shell := l.GetShell(); shell != nil {
				for _, f := range shell.Functions {
					functions = append(functions, f)
				}
			}
		}
	}

	for _, e := range []string{"to-geojson", "closest", "get-string", "reachable"} {
		if !contains(e, functions) {
			t.Errorf("Function %q not included in area features: %v", e, functions)
		}
	}
}

func contains(item string, items []string) bool {
	for _, i := range items {
		if i == item {
			return true
		}
	}
	return false
}
