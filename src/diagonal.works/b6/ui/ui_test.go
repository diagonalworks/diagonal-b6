package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"diagonal.works/b6"
)

func TestStateFilledFromStartupQuery(t *testing.T) {
	handler := StartupHandler{
		UI: &OpenSourceUI{
			World: b6.EmptyWorld{},
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
