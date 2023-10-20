package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"diagonal.works/b6"
)

func TestMapPositionFilledFromStartupQuery(t *testing.T) {
	handler := StartupHandler{
		World:    b6.EmptyWorld{},
		Renderer: NewDefaultUIRenderer(b6.EmptyWorld{}),
	}

	url := "http://b6.diagonal.works/startup?ll=51.5321489,-0.1253271&z=18"
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
}
