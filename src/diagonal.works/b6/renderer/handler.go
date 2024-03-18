package renderer

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"diagonal.works/b6"

	"github.com/golang/geo/s2"
	"google.golang.org/protobuf/proto"
)

type TileArgs struct {
	Q string
	V string
	R b6.FeatureID
}

type TileHandler struct {
	Renderer Renderer
}

func (h *TileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tile, err := b6.TileFromURLPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Bad tile path request", http.StatusBadRequest)
		return
	}
	bounds := tile.RectBound()
	if !bounds.IsValid() {
		http.Error(w, "Invalid tile", http.StatusBadRequest)
		return
	}
	query := r.URL.Query()
	args := TileArgs{Q: query.Get("q"), V: query.Get("v"), R: b6.FeatureIDFromString(query.Get("r"))}
	rendered, err := h.Renderer.Render(tile, &args)
	if err != nil {
		log.Printf("Failed to render tile: %v", err)
		http.Error(w, "Failed to render tile", http.StatusInternalServerError)
		return
	}
	if r.URL.Query().Get("o") == "d" {
		outputPlainTextTile(tile, rendered, w, r)
	} else {
		outputEncodedTile(tile, rendered, w, r)
	}
}

func outputPlainTextTile(tile b6.Tile, rendered *Tile, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Plain", "text/plain")
	w.WriteHeader(200)
	b := &strings.Builder{}
	fmt.Fprintf(b, "Tile: %s (%s)\n", tile.String(), tile.RectBound().Center().String())
	for _, layer := range rendered.Layers {
		fmt.Fprintf(b, "Layer: %q\n", layer.Name)
		for _, feature := range layer.Features {
			fmt.Fprintf(b, "  ID: %d\n", feature.ID)
			switch g := feature.Geometry.(type) {
			case *Point:
				fmt.Fprintf(b, "    Point: %s\n", s2.LatLngFromPoint(g.ToS2Point()))
			case *LineString:
				fmt.Fprintf(b, "    LineString: %d vertices\n", len(*g.ToS2Polyline()))
			case *Polygon:
				fmt.Fprintf(b, "    Polygon: %d loops\n", g.ToS2Polygon().NumLoops())
			default:
				fmt.Fprintf(b, "    Unknown geometry\n")
			}
			for key, value := range feature.Tags {
				fmt.Fprintf(b, "    Tag: %q=%q\n", key, value)
			}
		}
	}
	w.Write([]byte(b.String()))
}

func outputEncodedTile(tile b6.Tile, rendered *Tile, w http.ResponseWriter, r *http.Request) {
	encoded := EncodeTile(tile, rendered)
	marshaled, err := proto.Marshal(encoded)
	if err != nil {
		log.Printf("Failed to marshal tile: %v", err)
		http.Error(w, "Failed to render tile", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.Header().Add("Content-Length", fmt.Sprintf("%d", len(marshaled)))
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(200)
	w.Write(marshaled)
}
