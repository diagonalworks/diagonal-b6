package ui

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

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

// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MAX_SAFE_INTEGER
const MaxSafeJavaScriptInteger = (1 << 53) - 1

type Options struct {
	StaticPath        string
	JavaScriptPath    string
	StaticV2Path      string
	StorybookPath     string
	EnableStorybook   bool
	BasemapRules      renderer.RenderRules
	UI                UI
	Worlds            ingest.Worlds
	APIOptions        api.Options
	InstrumentHandler func(handler http.Handler, name string) http.Handler
	Lock              *sync.RWMutex
}

type DropPrefixFilesystem struct {
	Prefix string
	Next   http.FileSystem
}

func (d *DropPrefixFilesystem) Open(filename string) (http.File, error) {
	if strings.HasPrefix(filename, d.Prefix) {
		return d.Next.Open(filename[len(d.Prefix):])
	}
	return nil, fs.ErrNotExist
}

type MergedFilesystem []string

func (m MergedFilesystem) Open(filename string) (http.File, error) {
	for _, path := range m {
		full := filepath.Join(path, filename)
		if _, err := os.Stat(full); err == nil {
			return os.Open(full)
		}
	}
	return nil, fs.ErrNotExist
}

func RegisterWebInterface(root *http.ServeMux, options *Options) error {
	staticPaths := strings.Split(options.StaticPath, ",")
	root.Handle("/", http.FileServer(MergedFilesystem(staticPaths)))

	root.Handle("/bundle.js", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(options.JavaScriptPath, "bundle.js"))
	}))

	staticV2Paths := strings.Split(options.StaticV2Path, ",")
	root.Handle("/assets/", http.FileServer(MergedFilesystem(staticV2Paths)))
	if len(staticV2Paths) > 0 {
		root.Handle("/v2.html", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(staticV2Paths[0], "index.html"))
		}))
	}

	if options.EnableStorybook {
		storybookPaths := strings.Split(options.StorybookPath, ",")
		root.Handle("/storybook/", http.FileServer(&DropPrefixFilesystem{
			Prefix: "/storybook",
			Next:   MergedFilesystem(storybookPaths),
		}))
	}

	var ui UI
	if options.UI != nil {
		ui = options.UI
	} else {
		ui = &OpenSourceUI{
			Worlds:          options.Worlds,
			Options:         options.APIOptions,
			FunctionSymbols: functions.Functions(),
			Adaptors:        functions.Adaptors(),
			BasemapRules:    renderer.BasemapRenderRules,
			Lock:            options.Lock,
		}
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

type lockedHandler struct {
	handler http.Handler
	lock    *sync.RWMutex
}

func (l *lockedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	l.handler.ServeHTTP(w, r)
}

func lockHandler(handler http.Handler, lock *sync.RWMutex) http.Handler {
	return &lockedHandler{handler: handler, lock: lock}
}

func RegisterTiles(root *http.ServeMux, options *Options) {
	rules := renderer.BasemapRenderRules
	if options.BasemapRules != nil {
		rules = options.BasemapRules
	}
	base := http.Handler(lockHandler(&renderer.TileHandler{Renderer: &renderer.BasemapRenderer{RenderRules: rules, Worlds: options.Worlds}}, options.Lock))
	if options.InstrumentHandler != nil {
		base = options.InstrumentHandler(base, "tiles_base")
	}
	root.Handle("/tiles/base/", base)
	query := http.Handler(lockHandler(&renderer.TileHandler{Renderer: renderer.NewQueryRenderer(options.Worlds, options.APIOptions.Cores)}, options.Lock))
	if options.InstrumentHandler != nil {
		query = options.InstrumentHandler(query, "tiles_query")
	}
	root.Handle("/tiles/query/", query)
	histogram := http.Handler(lockHandler(&renderer.TileHandler{Renderer: renderer.NewHistogramRenderer(rules, options.Worlds)}, options.Lock))
	if options.InstrumentHandler != nil {
		histogram = options.InstrumentHandler(histogram, "tiles_histogram")
	}
	root.Handle("/tiles/histogram/", histogram)
}

type StartupRequest struct {
	Root          b6.CollectionID
	MapCenter     *LatLngJSON
	MapZoom       *int
	OpenDockIndex *int
	Expression    string
}

func (s *StartupRequest) FillFromURL(url *url.URL) {
	if r := url.Query().Get("r"); len(r) > 0 {
		if id := b6.FeatureIDFromString(r[1:]); id.IsValid() && id.Type == b6.FeatureTypeCollection {
			s.Root = id.ToCollectionID()
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
	Root          *FeatureIDProtoJSON `json:"root,omitempty"`
	Expression    string              `json:"expression,omitempty"`
	Error         string              `json:"error,omitempty"`
	Session       uint64              `json:"session,omitempty"`
	Locked        bool                `json:"locked,omitempty"`
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
	Render(response *UIResponseJSON, value interface{}, root b6.CollectionID, locked bool, ui UI) error
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
	SendJSON(response, w, r)
}

type StackHandler struct {
	UI UI
}

func (s *StackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := &pb.UIRequestProto{}
	if !FillStackRequest(request, w, r) {
		return
	}

	response := NewUIResponseJSON()

	if err := s.UI.ServeStack(request, response, s.UI); err == nil {
		SendJSON(response, w, r)
	} else {
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func FillStackRequest(request *pb.UIRequestProto, w http.ResponseWriter, r *http.Request) bool {
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
			log.Println(err.Error())
			log.Printf("Bad request body")
			http.Error(w, "Bad request body", http.StatusBadRequest)
			return false
		}
	} else {
		log.Printf("Bad method")
		http.Error(w, "Bad method", http.StatusMethodNotAllowed)
		return false
	}

	if request.Expression == "" && request.Node == nil {
		log.Printf("No expression")
		http.Error(w, "No expression", http.StatusBadRequest)
		return false
	}

	return true
}

func SendJSON(value interface{}, w http.ResponseWriter, r *http.Request) {
	var output bytes.Buffer
	var encoder *json.Encoder
	var toClose io.Closer
	if strings.Index(r.Header.Get("Accept-Encoding"), "gzip") >= 0 {
		compresor := gzip.NewWriter(&output)
		encoder = json.NewEncoder(compresor)
		w.Header().Set("Content-Encoding", "gzip")
		toClose = compresor
	} else {
		encoder = json.NewEncoder(&output)
	}
	err := encoder.Encode(value)
	if err == nil && toClose != nil {
		err = toClose.Close()
	}
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(output.Bytes())
}

type OpenSourceUI struct {
	BasemapRules    renderer.RenderRules
	Worlds          ingest.Worlds
	Options         api.Options
	FunctionSymbols api.FunctionSymbols
	Adaptors        api.Adaptors
	Lock            *sync.RWMutex
}

func (o *OpenSourceUI) ServeStartup(request *StartupRequest, response *StartupResponseJSON, ui UI) error {
	o.Lock.RLock()
	defer o.Lock.RUnlock()
	w := o.Worlds.FindOrCreateWorld(request.Root.FeatureID())
	if root := b6.FindCollectionByID(request.Root, w); root != nil {
		response.Locked = root.Get("locked").String() == "yes"
		c := b6.AdaptCollection[string, b6.FeatureID](root)
		i := c.Begin()
		for {
			ok, err := i.Next()
			if err != nil {
				return fmt.Errorf("%s: %w", request.Root, err)
			} else if !ok {
				break
			}
			if i.Key() == "centroid" {
				if centroid := w.FindFeatureByID(i.Value()); centroid != nil {
					if p, ok := centroid.(b6.PhysicalFeature); ok {
						ll := s2.LatLngFromPoint(p.Point())
						response.MapCenter = &LatLngJSON{
							LatE7: int(ll.Lat.E7()),
							LngE7: int(ll.Lng.E7()),
						}
						response.MapZoom = DefaultMapZoom
					}
				}
			} else if i.Key() == "docked" {
				if docked := w.FindFeatureByID(i.Value()); docked != nil {
					uiResponse := NewUIResponseJSON()
					if err := ui.Render(uiResponse, docked, request.Root, true, ui); err == nil {
						response.Docked = append(response.Docked, uiResponse)
					} else {
						return fmt.Errorf("%s: %w", i.Value(), err)
					}
				}
			}
		}
	}

	if request.Root.IsValid() {
		id := b6.NewProtoFromFeatureID(request.Root.FeatureID())
		response.Root = (*FeatureIDProtoJSON)(id)
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
	response.Session = uint64(rand.Int63n(MaxSafeJavaScriptInteger))
	return nil
}

func (o *OpenSourceUI) ServeStack(request *pb.UIRequestProto, response *UIResponseJSON, ui UI) error {
	o.Lock.RLock()
	defer o.Lock.RUnlock()
	root := b6.NewFeatureIDFromProto(request.Root)
	w := o.Worlds.FindOrCreateWorld(root)

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
			ui.Render(response, err, root.ToCollectionID(), request.Locked, ui)
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
		ui.Render(response, err, root.ToCollectionID(), request.Locked, ui)
		return nil
	}

	vmContext := api.Context{
		World:           w,
		FunctionSymbols: o.FunctionSymbols,
		Adaptors:        o.Adaptors,
		Context:         context.Background(),
	}
	vmContext.FillFromOptions(&o.Options)

	result, err := api.Evaluate(expression, &vmContext)
	if err == nil {
		if change, ok := result.(ingest.Change); ok {
			o.Lock.RUnlock()
			o.Lock.Lock()
			var changed b6.Collection[b6.FeatureID, b6.FeatureID]
			if changed, err = change.Apply(w); err == nil {
				if first, ok := firstValueIfOnlyItem(changed); ok {
					err = ui.Render(response, first, root.ToCollectionID(), request.Locked, ui)
				} else {
					err = ui.Render(response, changed, root.ToCollectionID(), request.Locked, ui)
				}
				response.Proto.TilesChanged = true
			}
			o.Lock.Unlock()
			o.Lock.RLock()
		} else {
			err = ui.Render(response, result, root.ToCollectionID(), request.Locked, ui)
		}
	}
	if err != nil {
		ui.Render(response, err, root.ToCollectionID(), request.Locked, ui)
	}
	return nil
}

func firstValueIfOnlyItem(c b6.UntypedCollection) (any, bool) {
	i := c.BeginUntyped()
	ok, err := i.Next()
	if !ok || err != nil {
		return nil, false
	}
	first := i.Value()
	ok, err = i.Next()
	if !ok && err == nil {
		return first, true
	}
	return nil, false
}

func (o *OpenSourceUI) Render(response *UIResponseJSON, value interface{}, root b6.CollectionID, locked bool, ui UI) error {
	if err := o.fillResponseFromResult(response, value, o.Worlds.FindOrCreateWorld(root.FeatureID())); err == nil {
		shell := &pb.ShellLineProto{
			Functions: make([]string, 0),
		}
		shell.Functions = fillMatchingFunctionSymbols(shell.Functions, value, o.FunctionSymbols)
		response.Proto.Stack.Substacks = append(response.Proto.Stack.Substacks, &pb.SubstackProto{
			Lines: []*pb.LineProto{{Line: &pb.LineProto_Shell{Shell: shell}}},
		})
		return nil
	} else {
		return o.fillResponseFromResult(response, err, o.Worlds.FindOrCreateWorld(root.FeatureID()))
	}
}

func (o *OpenSourceUI) fillResponseFromResult(response *UIResponseJSON, result interface{}, w b6.World) error {
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
		if c := b6.FindCollectionByID(id, w); c != nil {
			substack := &pb.SubstackProto{}
			if err := fillSubstackFromCollection(substack, c, p, w); err == nil {
				p.Stack.Substacks = append(p.Stack.Substacks, substack)
			} else {
				return err
			}
		}
	case b6.Feature:
		if c, ok := r.(b6.CollectionFeature); ok {
			if b6 := c.Get("b6"); b6.Value.String() == "histogram" {
				return fillResponseFromHistogramFeature(response, c, w)
			}
		}
		p.Stack.Substacks = fillSubstacksFromFeature(p.Stack.Substacks, r, w)
		highlightInResponse(p, r.FeatureID())
	case b6.FeatureID:
		if f := w.FindFeatureByID(r); f != nil {
			return o.fillResponseFromResult(response, f, w)
		} else {
			return o.fillResponseFromResult(response, r.String(), w)
		}
	case b6.Query:
		if q, ok := api.UnparseQuery(r); ok {
			var substack pb.SubstackProto
			fillSubstackFromAtom(&substack, AtomFromString(q))
			p.Stack.Substacks = append(p.Stack.Substacks, &substack)
			p.Layers = append(p.Layers, &pb.MapLayerProto{
				Path:   "query",
				Q:      q,
				Before: pb.MapLayerPosition_MapLayerPositionEnd,
			})
		} else {
			// TODO: Improve the rendering of queries
			var substack pb.SubstackProto
			fillSubstackFromAtom(&substack, AtomFromString("Query"))
			p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		}
	case b6.Tag:
		var substack pb.SubstackProto
		fillSubstackFromAtom(&substack, AtomFromValue(r, w))
		p.Stack.Substacks = append(p.Stack.Substacks, &substack)
		if !o.BasemapRules.IsRendered(r) {
			if q, ok := api.UnparseQuery(b6.Tagged(r)); ok {
				before := pb.MapLayerPosition_MapLayerPositionEnd
				if r.Key == "#boundary" {
					before = pb.MapLayerPosition_MapLayerPositionBuildings
				}
				p.Layers = append(p.Layers, &pb.MapLayerProto{
					Path:   "query",
					Q:      q,
					Before: before,
				})
			}
		}
	case b6.UntypedCollection:
		substack := &pb.SubstackProto{}
		if err := fillSubstackFromCollection(substack, r, p, w); err == nil {
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
	case b6.Geometry:
		switch r.GeometryType() {
		case b6.GeometryTypePoint:
			ll := s2.LatLngFromPoint(r.Point())
			atom := &pb.AtomProto{
				Atom: &pb.AtomProto_Value{
					Value: fmt.Sprintf("%f, %f", ll.Lat.Degrees(), ll.Lng.Degrees()),
				},
			}
			var substack1 pb.SubstackProto
			fillSubstackFromAtom(&substack1, atom)
			p.Stack.Substacks = append(p.Stack.Substacks, &substack1)
			response.AddGeoJSON(r.ToGeoJSON())
		case b6.GeometryTypePath:
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
		default:
			return o.fillResponseFromResult(response, r.ToGeoJSON(), w)
		}
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
						Atom: AtomFromValue(r, w),
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
