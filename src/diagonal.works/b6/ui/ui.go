package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
	"diagonal.works/b6/geojson"
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
	UI                UI
	Cores             int
	World             ingest.MutableWorld
	APIOptions        api.Options
	InstrumentHandler func(handler http.Handler, name string) http.Handler
}

type filesystem []string

func (f filesystem) Open(filename string) (http.File, error) {
	for _, path := range f {
		full := filepath.Join(path, filename)
		if _, err := os.Stat(full); err == nil {
			return os.Open(full)
		}
	}
	return nil, os.ErrExist
}

func RegisterWebInterface(root *http.ServeMux, options *Options) error {
	staticPaths := strings.Split(options.StaticPath, ",")
	root.Handle("/", http.FileServer(filesystem(staticPaths)))

	root.Handle("/bundle.js", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(options.JavaScriptPath, "bundle.js"))
	}))

	var ui UI
	if options.UI != nil {
		ui = options.UI
	} else {
		ui = NewDefaultUI(options.World)
	}
	startup := http.Handler(&StartupHandler{UI: ui})
	if options.InstrumentHandler != nil {
		startup = options.InstrumentHandler(startup, "startup")
	}
	root.Handle("/startup", startup)
	stack := http.Handler(&StackHandler{UI: ui})
	if options.InstrumentHandler != nil {
		stack = options.InstrumentHandler(stack, "ui")
	}
	root.Handle("/stack", stack)

	return nil
}

func NewDefaultUI(w b6.World) UI {
	return &OpenSourceUI{
		World:           w,
		FunctionSymbols: functions.Functions(),
		Adaptors:        functions.Adaptors(),
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

type StartupRequest struct {
	RenderContext b6.CollectionID
	MapCenter     *LatLngJSON
	MapZoom       *int
	OpenDockIndex *int
	Expression    string
}

func (s *StartupRequest) FillFromURL(url *url.URL) {
	if r := url.Query().Get("r"); len(r) > 0 {
		if id := b6.FeatureIDFromString(r[1:]); id.IsValid() && id.Type == b6.FeatureTypeCollection {
			s.RenderContext = id.ToCollectionID()
		}
	}

	if ll := url.Query().Get("ll"); len(ll) > 0 {
		if lll, err := b6.LatLngFromString(ll); err == nil {
			s.MapCenter = &LatLngJSON{
				LatE7: int(lll.Lat.E7()),
				LngE7: int(lll.Lng.E7()),
			}
		}
	}

	if z := url.Query().Get("z"); len(z) > 0 {
		if zi, err := strconv.ParseInt(z, 10, 64); err == nil {
			s.MapZoom = new(int)
			*s.MapZoom = int(zi)
		}
	}

	if d := url.Query().Get("d"); len(d) > 0 {
		if di, err := strconv.ParseInt(d, 10, 64); err == nil {
			s.OpenDockIndex = new(int)
			*s.OpenDockIndex = int(di)
		}
	}

	if e := url.Query().Get("e"); len(e) > 0 {
		s.Expression = e
	}
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

type UIResponseProtoJSON pb.UIResponseProto

func (b *UIResponseProtoJSON) MarshalJSON() ([]byte, error) {
	return protojson.Marshal((*pb.UIResponseProto)(b))
}

func (b *UIResponseProtoJSON) UnmarshalJSON(buffer []byte) error {
	return protojson.Unmarshal(buffer, (*pb.UIResponseProto)(b))
}

type UIResponseJSON struct {
	Proto   *UIResponseProtoJSON `json:"proto,omitempty"`
	GeoJSON []geojson.GeoJSON    `json:"geoJSON,omitempty"`
}

func NewUIResponseJSON() *UIResponseJSON {
	return &UIResponseJSON{
		Proto: &UIResponseProtoJSON{
			Stack: &pb.StackProto{},
		},
	}
}

func (u *UIResponseJSON) AddGeoJSON(g geojson.GeoJSON) {
	u.GeoJSON = append(u.GeoJSON, g)
	u.Proto.GeoJSON = append(u.Proto.GeoJSON, &pb.GeoJSONProto{
		Index: int32(len(u.GeoJSON) - 1),
	})
}

type UI interface {
	ServeStartup(request *StartupRequest, response *StartupResponseJSON, ui UI) error
	ServeStack(request *pb.UIRequestProto, response *UIResponseJSON, ui UI) error
	Render(response *UIResponseJSON, value interface{}, context b6.CollectionFeature, locked bool, ui UI) error
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

const DefaultMapZoom = 16

type StartupHandler struct {
	UI UI
}

func (s *StartupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := StartupRequest{}
	request.FillFromURL(r.URL)
	response := StartupResponseJSON{
		Version: b6.BackendVersion,
	}
	if err := s.UI.ServeStartup(&request, &response, s.UI); err != nil {
		response.Error = err.Error()
	}

	output, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(output))
}

type StackHandler struct {
	UI UI
}

func (s *StackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := &pb.UIRequestProto{}
	response := NewUIResponseJSON()

	if r.Method == "GET" {
		request.Expression = r.URL.Query().Get("e")
	} else if r.Method == "POST" {
		var err error
		var body []byte
		if body, err = io.ReadAll(r.Body); err == nil {
			r.Body.Close()
			err = protojson.Unmarshal(body, request)
		}
		if err != nil {
			log.Printf("Bad request body")
			http.Error(w, "Bad request body", http.StatusBadRequest)
			return
		}
	} else {
		log.Printf("Bad method")
		http.Error(w, "Bad method", http.StatusMethodNotAllowed)
		return
	}

	if request.Expression == "" && request.Node == nil {
		log.Printf("No expression")
		http.Error(w, "No expression", http.StatusBadRequest)
		return
	}

	if err := s.UI.ServeStack(request, response, s.UI); err == nil {
		output, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(output))
	} else {
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

type OpenSourceUI struct {
	BasemapRules    renderer.RenderRules
	World           b6.World
	Options         api.Options
	FunctionSymbols api.FunctionSymbols
	Adaptors        api.Adaptors
}

func (o *OpenSourceUI) ServeStartup(request *StartupRequest, response *StartupResponseJSON, ui UI) error {
	if context := b6.FindCollectionByID(request.RenderContext, o.World); context != nil {
		if context, ok := context.(b6.CollectionFeature); ok {
			c := b6.AdaptCollection[string, b6.FeatureID](context)
			i := c.Begin()
			for {
				ok, err := i.Next()
				if err != nil {
					return fmt.Errorf("%s: %w", request.RenderContext, err)
				} else if !ok {
					break
				}
				if i.Key() == "centroid" {
					if centroid := o.World.FindFeatureByID(i.Value()); centroid != nil {
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
					if docked := o.World.FindFeatureByID(i.Value()); docked != nil {
						uiResponse := NewUIResponseJSON()
						if err := ui.Render(uiResponse, docked, context, true, ui); err == nil {
							response.Docked = append(response.Docked, uiResponse)
						} else {
							return fmt.Errorf("%s: %w", i.Value(), err)
						}
					}
				}
			}
		}
	}

	if request.RenderContext.IsValid() {
		id := b6.NewProtoFromFeatureID(request.RenderContext.FeatureID())
		response.Context = (*FeatureIDProtoJSON)(id)
	}

	if request.MapCenter != nil {
		response.MapCenter = request.MapCenter
	}
	if request.MapZoom != nil {
		response.MapZoom = *request.MapZoom
	}
	if request.OpenDockIndex != nil {
		response.OpenDockIndex = request.OpenDockIndex
	}
	response.Expression = request.Expression
	return nil
}

func (o *OpenSourceUI) ServeStack(request *pb.UIRequestProto, response *UIResponseJSON, ui UI) error {
	var root b6.CollectionFeature
	if request.Context != nil && request.Context.Type == pb.FeatureType_FeatureTypeCollection {
		root = b6.FindCollectionByID(b6.NewFeatureIDFromProto(request.Context).ToCollectionID(), o.World)
	}

	var expression b6.Expression
	if request.Expression != "" {
		var err error
		if request.Node == nil {
			expression, err = api.ParseExpression(request.Expression)
		} else {
			var lhs b6.Expression
			if err = lhs.FromProto(request.Node); err == nil {
				expression, err = api.ParseExpressionWithLHS(request.Expression, lhs)
			}
		}
		if err != nil {
			ui.Render(response, err, root, request.Locked, ui)
			var substack pb.SubstackProto
			fillSubstackFromError(&substack, err)
			response.Proto.Stack.Substacks = append(response.Proto.Stack.Substacks, &substack)
			return nil
		}
	} else {
		expression.FromProto(request.Node)
	}
	expression = api.Simplify(expression, o.FunctionSymbols)

	if !request.Locked {
		substack := &pb.SubstackProto{}
		fillSubstackFromExpression(substack, expression, true)
		if len(substack.Lines) > 0 {
			response.Proto.Stack.Substacks = append(response.Proto.Stack.Substacks, substack)
		}
	}

	if unparsed, ok := api.UnparseExpression(expression); ok {
		response.Proto.Expression = unparsed
	}

	var err error
	if response.Proto.Node, err = expression.ToProto(); err != nil {
		ui.Render(response, err, root, request.Locked, ui)
		return nil
	}

	vmContext := api.Context{
		World:           o.World,
		FunctionSymbols: o.FunctionSymbols,
		Adaptors:        o.Adaptors,
		Context:         context.Background(),
	}
	vmContext.FillFromOptions(&o.Options)

	result, err := api.Evaluate(expression, &vmContext)
	if err == nil {
		err = ui.Render(response, result, root, request.Locked, ui)
	}
	if err != nil {
		ui.Render(response, err, root, request.Locked, ui)
	}
	return err
}

func (o *OpenSourceUI) Render(response *UIResponseJSON, value interface{}, context b6.CollectionFeature, locked bool, ui UI) error {
	if err := o.fillResponseFromResult(response, value); err == nil {
		shell := &pb.ShellLineProto{
			Functions: make([]string, 0),
		}
		shell.Functions = fillMatchingFunctionSymbols(shell.Functions, value, o.FunctionSymbols)
		response.Proto.Stack.Substacks = append(response.Proto.Stack.Substacks, &pb.SubstackProto{
			Lines: []*pb.LineProto{{Line: &pb.LineProto_Shell{Shell: shell}}},
		})
		return nil
	} else {
		return o.fillResponseFromResult(response, err)
	}
}

func (o *OpenSourceUI) fillResponseFromResult(response *UIResponseJSON, result interface{}) error {
	p := (*pb.UIResponseProto)(response.Proto)
	switch r := result.(type) {
	case error:
		var substack pb.SubstackProto
		fillSubstackFromError(&substack, r)
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
	case string:
		var substack pb.SubstackProto
		fillSubstackFromAtom(&substack, AtomFromString(r))
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
	case b6.ExpressionFeature:
		// This is not perfect, as it makes original expression that
		// returned the ExpressionFeature, and the expression from the
		// feature itself, look like part of the same stack.
		// TODO: improve the UX for expression features
		substack := &pb.SubstackProto{}
		expression := api.AddPipelines(api.Simplify(b6.NewCallExpression(r.Expression(), []b6.Expression{}), o.FunctionSymbols))
		fillSubstackFromExpression(substack, expression, true)
		if len(substack.Lines) > 0 {
			response.Proto.Stack.Substacks = append(response.Proto.Stack.Substacks, substack)
		}
		id := b6.MakeCollectionID(r.ExpressionID().Namespace, r.ExpressionID().Value)
		if c := b6.FindCollectionByID(id, o.World); c != nil {
			substack := &pb.SubstackProto{}
			if err := fillSubstackFromCollection(substack, c, p, o.World); err == nil {
				p.Stack.Substacks = append(p.Stack.Substacks, substack)
			} else {
				return err
			}
		}
	case b6.Feature:
		p.Stack.Substacks = fillSubstacksFromFeature(p.Stack.Substacks, r, o.World)
		highlightInResponse(p, r.FeatureID())
	case b6.Query:
		if q, ok := api.UnparseQuery(r); ok {
			var substack pb.SubstackProto
			fillSubstackFromAtom(&substack, AtomFromString(q))
			p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		} else {
			// TODO: Improve the rendering of queries
			var substack pb.SubstackProto
			fillSubstackFromAtom(&substack, AtomFromString("Query"))
			p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		}
	case b6.Tag:
		var substack pb.SubstackProto
		fillSubstackFromAtom(&substack, AtomFromValue(r, o.World))
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		if !o.BasemapRules.IsRendered(r) {
			if q, ok := api.UnparseQuery(b6.Tagged(r)); ok {
				before := pb.MapLayerPosition_MapLayerPositionEnd
				if r.Key == "#boundary" {
					before = pb.MapLayerPosition_MapLayerPositionBuildings
				}
				p.Layers = append(p.Layers, &pb.MapLayerProto{Query: q, Before: before})
			}
		}
	case *api.HistogramCollection:
		if err := fillResponseFromHistogram(p, r, o.World); err != nil {
			return err
		}
	case b6.UntypedCollection:
		substack := &pb.SubstackProto{}
		if err := fillSubstackFromCollection(substack, r, p, o.World); err == nil {
			p.Stack.Substacks = append(p.Stack.Substacks, substack)
		} else {
			return err
		}
	case b6.Area:
		dimension := 0.0
		for i := 0; i < r.Len(); i++ {
			dimension += b6.AreaToMeters2(r.Polygon(i).Area())
		}
		atom := &pb.AtomProto{
			Atom: &pb.AtomProto_Download{
				Download: fmt.Sprintf("%.2fm² area", dimension),
			},
		}
		var substack pb.SubstackProto
		fillSubstackFromAtom(&substack, atom)
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		response.AddGeoJSON(r.ToGeoJSON())
	case b6.Path:
		dimension := b6.AngleToMeters(r.Polyline().Length())
		atom := &pb.AtomProto{
			Atom: &pb.AtomProto_Download{
				Download: fmt.Sprintf("%.2fm path", dimension),
			},
		}
		var substack pb.SubstackProto
		fillSubstackFromAtom(&substack, atom)
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		response.AddGeoJSON(r.ToGeoJSON())
	case *geojson.FeatureCollection:
		var label string
		if n := len(r.Features); n == 1 {
			label = "1 GeoJSON feature"
		} else {
			label = fmt.Sprintf("%d GeoJSON features", n)
		}
		atom := &pb.AtomProto{
			Atom: &pb.AtomProto_Download{
				Download: fmt.Sprintf(label),
			},
		}
		var substack pb.SubstackProto
		fillSubstackFromAtom(&substack, atom)
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		response.AddGeoJSON(r)
	case *geojson.Feature:
		atom := &pb.AtomProto{
			Atom: &pb.AtomProto_Download{
				Download: "GeoJSON feature",
			},
		}
		var substack pb.SubstackProto
		fillSubstackFromAtom(&substack, atom)
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		response.AddGeoJSON(r)
	case *geojson.Geometry:
		atom := &pb.AtomProto{
			Atom: &pb.AtomProto_Download{
				Download: "GeoJSON geometry",
			},
		}
		var substack pb.SubstackProto
		fillSubstackFromAtom(&substack, atom)
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		response.AddGeoJSON(geojson.NewFeatureWithGeometry(*r))
	default:
		substack := &pb.SubstackProto{
			Lines: []*pb.LineProto{{
				Line: &pb.LineProto_Value{
					Value: &pb.ValueLineProto{
						Atom: AtomFromValue(r, o.World),
					},
				},
			}},
		}
		p.Stack.Substacks = append(p.Stack.Substacks, substack)
	}
	switch r := result.(type) {
	case b6.Geometry:
		response.Proto.MapCenter = b6.NewPointProtoFromS2Point(b6.Centroid(r))
	case geojson.GeoJSON:
		response.Proto.MapCenter = b6.NewPointProtoFromS2Point(r.Centroid().ToS2Point())
	}
	return nil
}
