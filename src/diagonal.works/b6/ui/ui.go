package ui

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/renderer"
)

type Options struct {
	StaticPath     string
	JavaScriptPath string
	Cores          int
	World          ingest.MutableWorld
}

func RegisterWebInterface(root *http.ServeMux, options *Options) error {
	root.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(options.StaticPath, "index.html"))
		} else {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	}))
	root.Handle("/b6.css", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(options.StaticPath, "b6.css"))
	}))
	root.Handle("/bundle.js", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(options.JavaScriptPath, "bundle.js"))
	}))
	root.Handle("/images/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := strings.LastIndex(r.URL.Path, "/")
		http.ServeFile(w, r, filepath.Join(options.StaticPath, "images", r.URL.Path[i+1:]))
	}))

	root.Handle("/bootstrap", http.HandlerFunc(serveBootstrap))

	shell, err := NewShellHandler(options.World, options.Cores)
	if err != nil {
		return err
	}
	root.Handle("/shell", shell)

	return nil
}

func RegisterTiles(root *http.ServeMux, w b6.World) {
	tiles := &renderer.TileHandler{Renderer: &renderer.BasemapRenderer{World: w}}
	root.Handle("/tiles/base/", tiles)
}

type BootstrapResponseJSON struct {
	Version string
}

func serveBootstrap(w http.ResponseWriter, r *http.Request) {
	response := BootstrapResponseJSON{Version: b6.BackendVersion}
	output, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(output))
}
