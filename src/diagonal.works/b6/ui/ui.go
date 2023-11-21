package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
	"diagonal.works/b6/ingest"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/renderer"
	"github.com/golang/geo/s2"
	"google.golang.org/protobuf/encoding/protojson"
)

type Options struct {
	StaticPath        string
	JavaScriptPath    string
	BasemapRules      renderer.RenderRules
	Renderer          UIRenderer
	Cores             int
	World             ingest.MutableWorld
	APIOptions        api.Options
	InstrumentHandler func(handler http.Handler, name string) http.Handler
}

func RegisterWebInterface(root *http.ServeMux, options *Options) error {
	root.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(options.StaticPath, "index.html"))
		} else {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	}))
	for _, filename := range []string{"b6.css", "palette.html"} {
		filename := filename
		root.Handle("/"+filename, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(options.StaticPath, filename))
		}))
	}
	root.Handle("/bundle.js", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(options.JavaScriptPath, "bundle.js"))
	}))
	root.Handle("/images/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := strings.LastIndex(r.URL.Path, "/")
		http.ServeFile(w, r, filepath.Join(options.StaticPath, "images", r.URL.Path[i+1:]))
	}))

	var uiRenderer UIRenderer
	if options.Renderer != nil {
		uiRenderer = options.Renderer
	} else {
		uiRenderer = NewDefaultUIRenderer(options.World)
	}
	startup := http.Handler(&StartupHandler{World: options.World, Renderer: uiRenderer})
	if options.InstrumentHandler != nil {
		startup = options.InstrumentHandler(startup, "startup")
	}
	root.Handle("/startup", startup)
	ui := http.Handler(NewUIHandler(uiRenderer, options.World, options.APIOptions))
	if options.InstrumentHandler != nil {
		ui = options.InstrumentHandler(ui, "ui")
	}
	root.Handle("/ui", ui)

	return nil
}

func NewDefaultUIRenderer(w b6.World) UIRenderer {
	return &DefaultUIRenderer{
		World:           w,
		FunctionSymbols: functions.Functions(),
		BasemapRules:    renderer.BasemapRenderRules,
	}
}

func RegisterTiles(root *http.ServeMux, options *Options) {
	rules := renderer.BasemapRenderRules
	if options.BasemapRules != nil {
		rules = options.BasemapRules
	}
	base := http.Handler(&renderer.TileHandler{Renderer: &renderer.BasemapRenderer{RenderRules: rules, World: options.World}})
	if options.InstrumentHandler != nil {
		base = options.InstrumentHandler(base, "tiles_base")
	}
	root.Handle("/tiles/base/", base)
	query := http.Handler(&renderer.TileHandler{Renderer: renderer.NewQueryRenderer(options.World, options.Cores)})
	if options.InstrumentHandler != nil {
		query = options.InstrumentHandler(query, "tiles_query")
	}
	root.Handle("/tiles/query/", query)
}

type UIRenderer interface {
	Render(response *UIResponseJSON, value interface{}, context b6.CollectionFeature, locked bool) error
}

type FeatureIDProtoJSON pb.FeatureIDProto

func (b *FeatureIDProtoJSON) MarshalJSON() ([]byte, error) {
	return protojson.Marshal((*pb.FeatureIDProto)(b))
}

func (b *FeatureIDProtoJSON) UnmarshalJSON(buffer []byte) error {
	return protojson.Unmarshal(buffer, (*pb.FeatureIDProto)(b))
}

type LatLngJSON struct {
	LatE7 int `json:"latE7"`
	LngE7 int `json:"lngE7"`
}

type StartupResponseJSON struct {
	Version       string              `json:"version,omitempty"`
	Docked        []*UIResponseJSON   `json:"docked,omitempty"`
	OpenDockIndex *int                `json:"openDockIndex,omitempty"`
	MapCenter     *LatLngJSON         `json:"mapCenter,omitempty"`
	MapZoom       int                 `json:"mapZoom,omitempty"`
	Context       *FeatureIDProtoJSON `json:"context,omitempty"`
	Expression    string              `json:"expression,omitempty"`
	Error         string              `json:"error,omitempty"`
}

const DefaultMapZoom = 16

type StartupHandler struct {
	World    b6.World
	Renderer UIRenderer
}

func (s *StartupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := &StartupResponseJSON{
		Version: b6.BackendVersion,
	}

	if r := r.URL.Query().Get("r"); len(r) > 0 {
		renderContext := b6.FeatureIDFromString(r[1:])
		if renderContext.IsValid() && renderContext.Type == b6.FeatureTypeCollection {
			if err := s.fillStartupResponseFromRenderContext(response, renderContext.ToCollectionID()); err != nil {
				response.Error = err.Error()
				output, _ := json.Marshal(response)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(output))
				return
			}
			response.Context = (*FeatureIDProtoJSON)(b6.NewProtoFromFeatureID(renderContext))
		}
	}

	if ll := r.URL.Query().Get("ll"); len(ll) > 0 {
		if lll, err := b6.LatLngFromString(ll); err == nil {
			response.MapCenter = &LatLngJSON{
				LatE7: int(lll.Lat.E7()),
				LngE7: int(lll.Lng.E7()),
			}
		}
	}

	if z := r.URL.Query().Get("z"); len(z) > 0 {
		if zi, err := strconv.ParseInt(z, 10, 64); err == nil {
			response.MapZoom = int(zi)
		}
	}

	if d := r.URL.Query().Get("d"); len(d) > 0 {
		if di, err := strconv.ParseInt(d, 10, 64); err == nil {
			i := int(di)
			response.OpenDockIndex = &i
		}
	}

	if e := r.URL.Query().Get("e"); len(e) > 0 {
		response.Expression = e
	}

	output, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(output))
}

func (s *StartupHandler) fillStartupResponseFromRenderContext(response *StartupResponseJSON, id b6.CollectionID) error {
	if context := b6.FindCollectionByID(id, s.World); context != nil {
		if context, ok := context.(b6.CollectionFeature); ok {
			c := b6.AdaptCollection[string, b6.FeatureID](context)
			i := c.Begin()
			for {
				ok, err := i.Next()
				if err != nil {
					return fmt.Errorf("%s: %w", id, err)
				} else if !ok {
					break
				}
				if i.Key() == "centroid" {
					if centroid := s.World.FindFeatureByID(i.Value()); centroid != nil {
						if p, ok := centroid.(b6.PhysicalFeature); ok {
							ll := s2.LatLngFromPoint(b6.Centroid(p))
							response.MapCenter = &LatLngJSON{
								LatE7: int(ll.Lat.E7()),
								LngE7: int(ll.Lng.E7()),
							}
							response.MapZoom = DefaultMapZoom
						}
					}
				} else if i.Key() == "docked" {
					if docked := s.World.FindFeatureByID(i.Value()); docked != nil {
						uiResponse := NewUIResponseJSON()
						if err := s.Renderer.Render(uiResponse, docked, context, true); err == nil {
							response.Docked = append(response.Docked, uiResponse)
						} else {
							return fmt.Errorf("%s: %w", i.Value(), err)
						}
					}
				}
			}
		}
	}
	return nil
}
