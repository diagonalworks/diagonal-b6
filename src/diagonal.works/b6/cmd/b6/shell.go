package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/ingest"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/renderer"
	"github.com/golang/geo/s2"
)

var titleTags = []b6.Tag{
	{Key: "#amenity"},
	{Key: "#shop"},
	{Key: "#building"},
	{Key: "entrance"},
	{Key: "#bridge"},
	{Key: "#highway"},
	{Key: "footway"},
	{Key: "#route"},
	{Key: "#railway"},
	{Key: "#leisure"},
	{Key: "#landuse"},
	{Key: "barrier"},
	{Key: "waterway"},
	{Key: "#water"},
	{Key: "#boundary"},
	{Key: "public_transport"},
}

type SpanJSON struct {
	Class string
	Text  string
}

type LineJSON []SpanJSON

func fillLines(v interface{}, lines []LineJSON) []LineJSON {
	switch v := v.(type) {
	case b6.Feature:
		return fillLinesFromFeature(v, lines)
	case api.Collection:
		return fillLinesFromCollection(v, lines)
	default:
		return append(lines, fillSpans(v, LineJSON{}))
	}
	return lines
}

func fillLinesFromFeature(f b6.Feature, lines []LineJSON) []LineJSON {
	line := LineJSON{}
	for _, key := range []string{"name", "addr:housename"} {
		if t := f.Get(key); t.IsValid() {
			line = append(line, SpanJSON{Text: t.Value + " ", Class: "name"})
			break
		}
	}
	line = fillSpansFromFeatureID(f.FeatureID(), false, line)
	lines = append(lines, line)
	for _, tag := range f.AllTags() {
		lines = append(lines, fillSpansFromTag(tag, LineJSON{}))
	}
	switch f := f.(type) {
	case b6.PathFeature:
		for i := 0; i < f.Len(); i++ {
			id := f.Feature(i).FeatureID()
			lines = append(lines, LineJSON{SpanJSON{Text: strconv.Itoa(i)}, SpanJSON{Text: ": "}, SpanJSON{Text: "/" + id.String()}})
		}
	}
	return lines
}

func fillLinesFromCollection(c api.Collection, lines []LineJSON) []LineJSON {
	i := c.Begin()
	for {
		if ok, err := i.Next(); err != nil {
			lines = fillLinesFromError(err, lines)
			break
		} else if !ok {
			break
		}
		line := LineJSON{}
		line = fillSpans(i.Key(), line)
		line = append(line, SpanJSON{Text: ": "})
		line = fillSpans(i.Value(), line)
		lines = append(lines, line)

	}
	return lines
}

func fillLinesFromError(err error, lines []LineJSON) []LineJSON {
	return append(lines, LineJSON{SpanJSON{Text: err.Error()}})
}

func fillSpans(v interface{}, spans []SpanJSON) []SpanJSON {
	switch v := v.(type) {
	case b6.Feature:
		spans = fillSpansFromFeature(v, spans)
	case b6.FeatureID:
		spans = fillSpansFromFeatureID(v, true, spans)
	case b6.Tag:
		spans = fillSpansFromTag(v, spans)
	case b6.Point:
		ll := s2.LatLngFromPoint(v.Point())
		spans = append(spans, SpanJSON{Text: fmt.Sprintf("%f, %f", ll.Lat.Degrees(), ll.Lng.Degrees())})
	case geojson.GeoJSON:
		spans = fillSpansFromGeoJSON(v, spans)
	default:
		spans = append(spans, SpanJSON{Text: fmt.Sprintf("%v", v)})
	}
	return spans
}

func fillSpansFromFeature(f b6.Feature, spans []SpanJSON) []SpanJSON {
	for _, key := range []string{"name", "addr:housename"} {
		if t := f.Get(key); t.IsValid() {
			spans = append(spans, SpanJSON{Text: t.Value + " ", Class: "name"})
			break
		}
	}
	spans = fillSpansFromFeatureID(f.FeatureID(), true, spans)
	spans = append(spans, SpanJSON{Text: " "})
	for _, t := range titleTags {
		if tt := f.Get(t.Key); tt.IsValid() && (t.Value == "" || t.Value == tt.Value) {
			spans = fillSpansFromTag(tt, spans)
			spans = append(spans, SpanJSON{Text: " "})
			break
		}
	}
	return spans
}

func fillSpansFromFeatureID(id b6.FeatureID, abbreviate bool, spans []SpanJSON) []SpanJSON {
	return append(spans, SpanJSON{Text: api.FeatureIDToExpression(id, abbreviate), Class: "literal"})
}

func fillSpansFromTag(t b6.Tag, spans []SpanJSON) []SpanJSON {
	if t.IsValid() {
		return append(spans, SpanJSON{Text: api.TagToExpression(t), Class: "literal"})
	}
	return append(spans, SpanJSON{Text: "tag", Class: "invalid"})
}

func fillSpansFromGeoJSON(g geojson.GeoJSON, spans []SpanJSON) []SpanJSON {
	switch g := g.(type) {
	case *geojson.Feature:
		spans = append(spans, SpanJSON{Text: "GeoJSON Feature"})
	case *geojson.FeatureCollection:
		label := "Features"
		if len(g.Features) == 1 {
			label = "Feature"
		}
		spans = append(spans, SpanJSON{Text: fmt.Sprintf("GeoJSON FeatureCollection, %d %s", len(g.Features), label)})
	default:
		spans = append(spans, SpanJSON{Text: "GeoJSON"})
	}
	return spans
}

type UIChange interface {
	Apply(r *ShellResponseJSON) error
}

type FeatureIDJSON [3]string

func MakeFeatureIDJSON(id b6.FeatureID) FeatureIDJSON {
	return FeatureIDJSON{id.Type.String(), id.Namespace.String(), strconv.FormatUint(id.Value, 16)}
}

func (f FeatureIDJSON) FeatureID() b6.FeatureID {
	v, err := strconv.ParseUint(f[2], 16, 64)
	if err != nil {
		return b6.FeatureIDInvalid
	}
	return b6.FeatureID{
		Type:      b6.FeatureTypeFromString(f[0]),
		Namespace: b6.Namespace(f[1]),
		Value:     v,
	}
}

type FeatureColourJSON struct {
	ID     FeatureIDJSON
	Colour string
}

func (f *FeatureColourJSON) MarshalJSON() ([]byte, error) {
	return json.Marshal([2]interface{}{f.ID, f.Colour})
}

type MapLayer uint8
type MapLayers uint8

const (
	MapLayerBasemap    MapLayer = 1
	MapLayerBoundaries          = 2
	MapLayerPoints              = 4
)

func (m MapLayers) Has(l MapLayer) bool {
	return (m & MapLayers(l)) != 0
}

func mapLayersForQuery(q b6.Query) MapLayers {
	switch q := q.(type) {
	case b6.Tagged:
		if q.Key == "#boundary" {
			return MapLayers(MapLayerBoundaries)
		} else if q.Key == "#place" && q.Value == "uprn" {
			return MapLayers(MapLayerPoints)
		}
	case b6.Keyed:
		if q.Key == "#boundary" {
			return MapLayers(MapLayerBoundaries)
		}
	case b6.Intersection:
		var l MapLayers
		for _, qq := range q {
			l |= mapLayersForQuery(qq)
		}
		return l
	case b6.Union:
		var l MapLayers
		for _, qq := range q {
			l |= mapLayersForQuery(qq)
		}
		return l
	}
	return MapLayers(MapLayerBasemap)
}

type showChange struct {
	GeoJSON         geojson.GeoJSON
	BoundariesLayer string
	PointsLayer     string
	FeatureColours  []FeatureColourJSON
	Lines           []LineJSON
}

func (c *showChange) Apply(r *ShellResponseJSON) error {
	r.GeoJSON = c.GeoJSON
	if c.GeoJSON != nil {
		centroid := c.GeoJSON.Centroid()
		r.Center = &centroid
	}
	r.UpdateLayers = true
	r.BoundariesLayer = c.BoundariesLayer
	r.PointsLayer = c.PointsLayer
	r.FeatureColours = c.FeatureColours
	if c.Lines != nil {
		r.Lines = append(r.Lines, c.Lines...)
	} else {
		r.Lines = []LineJSON{}
	}
	return nil
}

func show(v interface{}, c *api.Context) (UIChange, error) {
	change := &showChange{Lines: fillLines(v, []LineJSON{})}
	switch v := v.(type) {
	case b6.Query:
		var boundaries string
		var points string
		if layer := mapLayersForQuery(v); layer.Has(MapLayerBoundaries) {
			if e, ok := api.QueryToExpression(v); ok {
				boundaries = e
			}
		} else if layer.Has(MapLayerPoints) {
			if e, ok := api.QueryToExpression(v); ok {
				points = e
			}
		}
		change.BoundariesLayer = boundaries
		change.PointsLayer = points
	case b6.Feature:
		change.GeoJSON = v.ToGeoJSON()
	case b6.Renderable:
		change.GeoJSON = v.ToGeoJSON()
	case geojson.GeoJSON:
		change.GeoJSON = v
	case b6.FeatureID:
		f := c.World.FindFeatureByID(v)
		if f != nil {
			return show(f, c)
		}
	case api.Collection:
		i := v.Begin()
		g := geojson.NewFeatureCollection()
		for {
			ok, err := i.Next()
			if err != nil {
				return nil, err
			}
			if !ok {
				break
			}
			switch vv := i.Value().(type) {
			case b6.Renderable:
				g.Add(vv.ToGeoJSON())
			case b6.FeatureID:
				f := c.World.FindFeatureByID(vv)
				if f != nil {
					g.Add(f.ToGeoJSON())
				}
			}
		}
		change.GeoJSON = g
	}
	return change, nil
}

func showColours(collection api.FeatureIDAnyCollection, c *api.Context) (UIChange, error) {

	min := math.Inf(1)
	max := math.Inf(-1)
	ids := make([]b6.FeatureID, 0)
	values := make([]float64, 0)

	i := collection.Begin()
	n := 0
	for {
		if ok, err := i.Next(); err != nil {
			return nil, err
		} else if !ok {
			break
		}
		if id, ok := i.Key().(b6.FeatureID); ok {
			if v, ok := api.ToFloat(i.Value()); ok {
				if v < min {
					min = v
				}
				if v > max {
					max = v
				}
				ids = append(ids, id)
				values = append(values, v)
			}
		}
		n++
	}

	line := LineJSON{SpanJSON{Text: fmt.Sprintf("n: %d values: %d min: %f max: %f", n, len(ids), min, max)}}

	gradient := renderer.Gradient{
		{Value: 0.0, Colour: renderer.ColourFromHexString("#796bbe")},
		{Value: 0.143, Colour: renderer.ColourFromHexString("#a673a0")},
		{Value: 0.286, Colour: renderer.ColourFromHexString("#cc7a87")},
		{Value: 0.429, Colour: renderer.ColourFromHexString("#ee8171")},
		{Value: 0.571, Colour: renderer.ColourFromHexString("#ff9068")},
		{Value: 0.714, Colour: renderer.ColourFromHexString("#ffa86d")},
		{Value: 0.857, Colour: renderer.ColourFromHexString("#ffbe71")},
		{Value: 1.00, Colour: renderer.ColourFromHexString("#ffd476")},
	}

	var colours []FeatureColourJSON
	if min < max {
		for i := range values {
			colour := gradient.Interpolate((values[i] - min) / (max - min))
			colours = append(colours, FeatureColourJSON{ID: MakeFeatureIDJSON(ids[i]), Colour: colour.ToHexString()})
		}
	}
	return &showChange{
		Lines:          []LineJSON{line},
		FeatureColours: colours,
	}, nil
}

type ShellHandler struct {
	World      *ingest.MutableOverlayWorld
	Cores      int
	Symbols    api.FunctionSymbols
	Convertors api.FunctionConvertors

	worldLock sync.RWMutex
}

func NewShellHandler(w *ingest.MutableOverlayWorld, cores int) (*ShellHandler, error) {
	local := make(api.FunctionSymbols)
	for name, f := range functions.Functions() {
		local[name] = f
	}
	local["show"] = show
	local["show-colours"] = showColours
	for name, f := range local {
		if err := functions.Validate(f, name); err != nil {
			return nil, err
		}
	}
	return &ShellHandler{
		World:      w,
		Cores:      cores,
		Symbols:    local,
		Convertors: functions.FunctionConvertors(),
	}, nil
}

type ShellRequestJSON struct {
	Expression string `json:"e,omitempty"`
}

type ShellResponseJSON struct {
	Lines            []LineJSON
	GeoJSON          geojson.GeoJSON
	HighlightGeoJSON geojson.GeoJSON
	Center           *geojson.Point
	BoundariesLayer  string
	PointsLayer      string
	FeatureColours   []FeatureColourJSON
	UpdateLayers     bool
	UpdateBuildings  bool
}

func (s *ShellHandler) evaluate(request *ShellRequestJSON, response *ShellResponseJSON) error {
	s.worldLock.RLock()
	defer s.worldLock.RUnlock()

	w := s.World
	if parsed, err := api.ParseExpression(request.Expression); err == nil {
		response.Lines = append(response.Lines, fillSpansFromExpression(request.Expression, api.OrderTokens(parsed), LineJSON{SpanJSON{Class: "prompt", Text: "b6"}}))
	} else {
		response.Lines = append(response.Lines, LineJSON{SpanJSON{Class: "prompt", Text: "b6"}, SpanJSON{Text: request.Expression}})
	}

	result, err := api.EvaluateString(request.Expression, w, s.Symbols, s.Convertors)
	if err != nil {
		return err
	} else if change, ok := result.(ingest.Change); ok {
		return s.applyChange(change, response)
	} else if c, ok := result.(UIChange); ok {
		return c.Apply(response)
	} else if p, ok := result.(b6.Point); ok {
		center := geojson.FromS2Point(p.Point())
		response.Center = &center
	} else if g, ok := result.(geojson.GeoJSON); ok {
		response.GeoJSON = g
		center := g.Centroid()
		response.Center = &center
	}
	response.Lines = fillLines(result, response.Lines)
	if renderable, ok := result.(b6.Renderable); ok {
		response.GeoJSON = renderable.ToGeoJSON()
	}
	return nil
}

func (s *ShellHandler) applyChange(change ingest.Change, response *ShellResponseJSON) error {
	s.worldLock.RUnlock()
	s.worldLock.Lock()
	ids, err := change.Apply(s.World)
	response.Lines = append(response.Lines, LineJSON{SpanJSON{Text: fmt.Sprintf("Changed %d features", len(ids))}})
	s.worldLock.Unlock()
	s.worldLock.RLock()
	return err
}

func fillSpansFromExpression(expression string, tokens []*pb.NodeProto, line LineJSON) LineJSON {
	i := int32(0)
	for _, t := range tokens {
		if t.Begin-i > 0 {
			line = append(line, SpanJSON{Text: expression[i:t.Begin]})
			i = t.Begin
		}
		if _, ok := t.Node.(*pb.NodeProto_Literal); ok {
			line = append(line, SpanJSON{Text: expression[t.Begin:t.End], Class: "literal"})
		} else {
			line = append(line, SpanJSON{Text: expression[t.Begin:t.End]})
		}
		i = t.End
	}
	if i < int32(len(expression)) {
		line = append(line, SpanJSON{Text: expression[i:]})
	}
	return line
}

func (s *ShellHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request ShellRequestJSON
	var response ShellResponseJSON
	if r.Method == "GET" {
		request.Expression = r.URL.Query().Get("e")
	} else if r.Method == "POST" {
		d := json.NewDecoder(r.Body)
		d.Decode(&request)
		r.Body.Close()
	} else {
		http.Error(w, "Bad method", http.StatusMethodNotAllowed)
		return
	}
	if request.Expression == "" {
		http.Error(w, "No expression", http.StatusBadRequest)
		return
	}

	if err := s.evaluate(&request, &response); err != nil {
		response.Lines = append(response.Lines, LineJSON{SpanJSON{Class: "error", Text: err.Error()}})
	}

	output, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(output))
}
