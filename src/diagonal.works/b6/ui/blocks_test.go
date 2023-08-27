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

	handler := BlockHandler{
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

	// TODO: Use typing - see the comment for the BlocksJSON definition
	var blocks map[string]interface{}
	d := json.NewDecoder(result.Body)
	if err := d.Decode(&blocks); err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}
	functions := blocks["Functions"].([]interface{})
	var functionNames []string
	for _, f := range functions {
		functionNames = append(functionNames, f.(string))
	}

	for _, e := range []string{"to-geojson", "closest", "get-string", "reachable"} {
		if !contains(e, functionNames) {
			t.Errorf("Function %q not included in area features: %v", e, functionNames)
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
