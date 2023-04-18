package renderer

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBadTilePathsReturnHTTPBadRequest(t *testing.T) {
	handler := &TileHandler{}
	paths := []string{
		// Tile paths with the wrong structure
		"/tile/broken",
		"/tile/one/two/three",
		"/tile/1/2/3",
		"/tile/1",
		"/tile/1/two/three.mvt",
		// Tile paths with the wrong prefix
		"/broken/1/2/3.mvt",
		// Tile paths with the correct structure, but an invalid coordinate
		"/tile/32/1/2.mvt",
		"/tile/2/1/5.mvt",
		"/tile/2/5/1.mvt",
	}
	for _, path := range paths {
		request := httptest.NewRequest("GET", path, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Errorf("Expected http.StatusBadRequest, found %d", response.Code)
		}
	}
}
