package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
	"diagonal.works/b6/ingest"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/test/camden"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestStateFilledFromStartupQuery(t *testing.T) {
	var lock sync.RWMutex
	handler := StartupHandler{
		UI: &OpenSourceUI{
			Worlds: &ingest.MutableWorlds{
				Base: b6.EmptyWorld{},
			},
			Lock: &lock,
		},
	}

	url := "http://b6.diagonal.works/startup?ll=51.5321489,-0.1253271&z=18&d=2&e=find-feature+/n/3501612811"
	request := httptest.NewRequest("GET", url, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	result := response.Result()
	if result.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, found %d", http.StatusOK, result.StatusCode)
	}

	var startupResponse StartupResponseJSON
	d := json.NewDecoder(result.Body)
	if err := d.Decode(&startupResponse); err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

	if expected := 515321489; startupResponse.MapCenter.LatE7 != expected {
		t.Errorf("Expected lat %d, found %d", expected, startupResponse.MapCenter.LatE7)
	}

	if expected := -1253271; startupResponse.MapCenter.LngE7 != expected {
		t.Errorf("Expected lng %d, found %d", expected, startupResponse.MapCenter.LatE7)
	}

	if expected := 18; startupResponse.MapZoom != expected {
		t.Errorf("Expected lng %d, found %d", expected, startupResponse.MapZoom)
	}

	if expected := 2; *startupResponse.OpenDockIndex != expected {
		t.Errorf("Expected open dock index %d, found %d", expected, *startupResponse.OpenDockIndex)
	}

	if expected := "find-feature /n/3501612811"; startupResponse.Expression != expected {
		t.Errorf("Expected expression %q, found %q", expected, startupResponse.Expression)
	}
}

func TestEvaluateFunctionThatChangesWorld(t *testing.T) {
	worlds := &ingest.MutableWorlds{
		Base: camden.BuildGranarySquareForTests(t),
	}
	var lock sync.RWMutex
	handler := StackHandler{
		UI: &OpenSourceUI{
			Worlds: worlds,
			Lock:   &lock,
			Evaluator: api.Evaluator{
				Worlds:          worlds,
				FunctionSymbols: functions.Functions(),
				Lock:            &lock,
			},
		},
	}

	root := b6.FeatureID{Type: b6.FeatureTypeCollection, Namespace: "diagonal.works/test/world", Value: 0}

	url := "http://b6.diagonal.works/stack"
	j := map[string]interface{}{
		"expression": fmt.Sprintf("add-tag /%s building:levels=\"25\"", camden.LightermanID),
		"root": map[string]interface{}{
			"type":      root.Type,
			"namespace": root.Namespace,
			"value":     root.Value,
		},
	}
	body, _ := json.Marshal(j)

	request := httptest.NewRequest("POST", url, bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	result := response.Result()
	if result.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, found %d", http.StatusOK, result.StatusCode)
	}

	modified := worlds.FindOrCreateWorld(root)
	f := modified.FindFeatureByID(camden.LightermanID.FeatureID())
	if f == nil {
		t.Fatal("failed to find expected feature")
	}
	expected := b6.Tag{Key: "building:levels", Value: b6.String("25")}
	if levels := f.Get(expected.Key); levels != expected {
		t.Errorf("expected tag %s, found %s", expected, levels)
	}
}
func TestEvaluateFunctionViaEvaluateEndpoint(t *testing.T) {
	worlds := &ingest.MutableWorlds{
		Base: camden.BuildGranarySquareForTests(t),
	}
	var lock sync.RWMutex
	handler := EvaluateHandler{
		Evaluator: api.Evaluator{
			Worlds:          worlds,
			FunctionSymbols: functions.Functions(),
			Adaptors:        functions.Adaptors(),
			Lock:            &lock,
		},
	}

	url := "http://b6.diagonal.works/evaluate"
	j := map[string]interface{}{
		"request": map[string]interface{}{
			"call": map[string]interface{}{
				"function": map[string]interface{}{
					"symbol": "add-ints",
				},
				"args": []map[string]interface{}{
					{
						"literal": map[string]interface{}{
							"intValue": 22,
						},
					},
					{
						"literal": map[string]interface{}{
							"intValue": 20,
						},
					},
				},
			},
		},
		"root": map[string]interface{}{
			"type":      b6.FeatureTypeCollection,
			"namespace": "diagonal.works/test/world",
			"value":     0,
		},
	}
	body, _ := json.Marshal(j)
	request := httptest.NewRequest("POST", url, bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	result := response.Result()
	if result.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, found %d", http.StatusOK, result.StatusCode)
	}
	body, err := io.ReadAll(response.Result().Body)
	if err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

	expected := map[string]interface{}{
		"result": map[string]interface{}{
			"literal": map[string]interface{}{
				"intValue": "42",
			},
		},
	}
	var actual map[string]interface{}
	json.Unmarshal(body, &actual)
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Expected no difference, found: %s", diff)
	}
}

func TestCompareScenarios(t *testing.T) {
	worlds := &ingest.MutableWorlds{
		Base: camden.BuildGranarySquareForTests(t),
	}

	var lock sync.RWMutex

	evaluator := api.Evaluator{
		Worlds:          worlds,
		FunctionSymbols: functions.Functions(),
		Adaptors:        functions.Adaptors(),
		Lock:            &lock,
	}

	handler := CompareHandler{
		Evaluator: evaluator,
		Worlds:    worlds,
	}

	analysis := b6.FeatureID{Type: b6.FeatureTypeCollection, Namespace: "diagonal.works/test/analysis", Value: 0}
	baseline := b6.FeatureID{Type: b6.FeatureTypeCollection, Namespace: "diagonal.works/test/world", Value: 0}
	scenario := b6.FeatureID{Type: b6.FeatureTypeCollection, Namespace: "diagonal.works/test/world", Value: 1}

	lock.RLock()
	e := "find [#amenity=restaurant] | map (get \"cuisine\") | histogram-with-id /" + analysis.String()
	if _, err := evaluator.EvaluateString(e, baseline); err != nil {
		t.Fatalf("Failed to setup baseline analysis: %s", err)
	}
	lock.RUnlock()

	w := worlds.FindOrCreateWorld(scenario)
	// The horror
	w.AddTag(camden.DishoomID, b6.Tag{Key: "#amenity", Value: b6.String("dentist")})

	url := "http://b6.diagonal.works/compare"
	j := map[string]interface{}{
		"analysis": map[string]interface{}{
			"type":      analysis.Type,
			"namespace": analysis.Namespace,
			"value":     analysis.Value,
		},
		"baseline": map[string]interface{}{
			"type":      baseline.Type,
			"namespace": baseline.Namespace,
			"value":     baseline.Value,
		},
		"scenarios": []map[string]interface{}{
			{
				"type":      scenario.Type,
				"namespace": scenario.Namespace,
				"value":     scenario.Value,
			},
		},
	}
	body, _ := json.Marshal(j)

	request := httptest.NewRequest("POST", url, bytes.NewBuffer(body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	result := response.Result()
	if result.StatusCode != http.StatusOK {
		m, _ := io.ReadAll(result.Body)
		t.Fatalf("Expected status %d, found %d: %s", http.StatusOK, result.StatusCode, string(m))
	}

	body, _ = io.ReadAll(result.Body)
	var comparison pb.ComparisonLineProto
	if err := protojson.Unmarshal(body, &comparison); err != nil {
		t.Fatalf("Failed to unmarshal response: %s", err)
	}

	if l := len(comparison.Scenarios); l != 1 {
		t.Fatalf("Expected one scenario, found %d", l)
	}

	if len(comparison.Baseline.Bars) != len(comparison.Scenarios[0].Bars) {
		t.Fatalf("Expected baseline and scenario to have same number of bars")
	}

	different := 0
	for i := range comparison.Baseline.Bars {
		if comparison.Baseline.Bars[i].Value != comparison.Scenarios[0].Bars[i].Value {
			different++
		}
	}

	if different != 1 {
		t.Errorf("Expected one different bar, found %d", different)
	}
}
