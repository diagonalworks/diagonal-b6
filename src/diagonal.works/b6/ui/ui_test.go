package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api/functions"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test/camden"
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
			Worlds:          worlds,
			FunctionSymbols: functions.Functions(),
			Lock:            &lock,
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
