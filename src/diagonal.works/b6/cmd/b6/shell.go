package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/graph"
	"diagonal.works/b6/ingest"
	pb "diagonal.works/b6/proto"

	"github.com/golang/geo/s2"
)

var titleTags = []b6.Tag{
	{Key: "#amenity"},
	{Key: "#shop"},
	{Key: "#building"},
	{Key: "public_transport", Value: "stop_position"},
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
}

func fillSpansFromTagProto(t *pb.TagProto, line LineJSON) LineJSON {
	return append(line, SpanJSON{Text: t.Key}, SpanJSON{Text: " = "}, SpanJSON{Text: t.Value})
}

func fillTitleFromTagProtos(ts []*pb.TagProto, line LineJSON) LineJSON {
	for _, t := range titleTags {
		for _, tt := range ts {
			if t.Key == tt.Key && (t.Value == "" || t.Value == tt.Value) {
				// TODO: Find a more principled way of add spaces
				return fillSpansFromTagProto(tt, append(line, SpanJSON{Text: " "}))
			}
		}
	}
	return line
}

type SpanJSON struct {
	Class string
	Text  string
}

type LineJSON []SpanJSON

func fillSpansFromFeatureIDProto(p *pb.FeatureIDProto, line LineJSON) LineJSON {
	id := b6.NewFeatureIDFromProto(p)
	return append(line, SpanJSON{Text: "/" + id.String()})
}

func fillSpansFromFeatureProto(p *pb.FeatureProto, line LineJSON) LineJSON {
	switch f := p.Feature.(type) {
	case *pb.FeatureProto_Point:
		line = fillSpansFromFeatureIDProto(f.Point.Id, line)
		return fillTitleFromTagProtos(f.Point.Tags, line)
	case *pb.FeatureProto_Path:
		line = fillSpansFromFeatureIDProto(f.Path.Id, line)
		return fillTitleFromTagProtos(f.Path.Tags, line)
	case *pb.FeatureProto_Area:
		line = fillSpansFromFeatureIDProto(f.Area.Id, line)
		return fillTitleFromTagProtos(f.Area.Tags, line)
	case *pb.FeatureProto_Relation:
		line = fillSpansFromFeatureIDProto(f.Relation.Id, LineJSON{})
		return fillTitleFromTagProtos(f.Relation.Tags, line)
	}
	return line
}

func fillSpansFromLiteralProto(l *pb.LiteralNodeProto, line LineJSON) LineJSON {
	switch v := l.Value.(type) {
	case *pb.LiteralNodeProto_NilValue:
		return append(line, SpanJSON{Text: "nil"})
	case *pb.LiteralNodeProto_StringValue:
		return append(line, SpanJSON{Text: v.StringValue})
	case *pb.LiteralNodeProto_IntValue:
		return append(line, SpanJSON{Text: fmt.Sprintf("%d", v.IntValue)})
	case *pb.LiteralNodeProto_FloatValue:
		return append(line, SpanJSON{Text: fmt.Sprintf("%f", v.FloatValue)})
	case *pb.LiteralNodeProto_PointValue:
		ll := s2.LatLngFromDegrees(float64(v.PointValue.LatE7)/1e7, float64(v.PointValue.LngE7)/1e7)
		return append(line, SpanJSON{Text: fmt.Sprintf("%f, %f", ll.Lat.Degrees(), ll.Lng.Degrees())})
	case *pb.LiteralNodeProto_FeatureIDValue:
		return fillSpansFromFeatureIDProto(v.FeatureIDValue, line)
	case *pb.LiteralNodeProto_FeatureValue:
		return fillSpansFromFeatureProto(v.FeatureValue, line)
	case *pb.LiteralNodeProto_CollectionValue:
		n := len(v.CollectionValue.Keys)
		entries := "entries"
		if n == 1 {
			entries = "entry"
		}
		return append(line, SpanJSON{Text: fmt.Sprintf("Collection, %d %s", n, entries)})
	case *pb.LiteralNodeProto_PairValue:
		line = append(line, SpanJSON{Text: "pair "})
		line = fillSpansFromLiteralProto(v.PairValue.First, line)
		line = append(line, SpanJSON{Text: " "})
		return fillSpansFromLiteralProto(v.PairValue.Second, line)
	case *pb.LiteralNodeProto_GeoJSONValue:
		line = append(line, SpanJSON{Text: "GeoJSON "})
		if g, err := unmarshalGeoJSON(v); err == nil {
			switch g := g.(type) {
			case *geojson.FeatureCollection:
				if len(g.Features) == 1 {
					line = append(line, SpanJSON{Text: "FeatureCollection, 1 feature"})
				} else {
					line = append(line, SpanJSON{Text: fmt.Sprintf("FeatureCollection, %d features", len(g.Features))})
				}
			case *geojson.Feature:
				line = append(line, SpanJSON{Text: fmt.Sprintf("Feature %s", g.Geometry.Type)})
			case *geojson.Geometry:
				line = append(line, SpanJSON{Text: fmt.Sprintf("Geometry %s", g.Type)})
			}
		} else {
			line = append(line, SpanJSON{Class: "error", Text: err.Error()})
		}
		return line
	}
	return append(line, SpanJSON{Text: fmt.Sprintf("%+v", l)})
}

func fillLinesFromTagProtos(ts []*pb.TagProto, lines []LineJSON) []LineJSON {
	for _, t := range ts {
		lines = append(lines, fillSpansFromTagProto(t, LineJSON{}))
	}
	return lines
}

func fillLinesFromLiteralProto(l *pb.LiteralNodeProto, lines []LineJSON) []LineJSON {
	switch v := l.Value.(type) {
	case *pb.LiteralNodeProto_CollectionValue:
		lines = append(lines, fillSpansFromLiteralProto(l, LineJSON{})) // Header
		for i, k := range v.CollectionValue.Keys {
			line := LineJSON{}
			line = fillSpansFromLiteralProto(k, line)
			line = append(line, SpanJSON{Text: ": "})
			line = fillSpansFromLiteralProto(v.CollectionValue.Values[i], line)
			lines = append(lines, line)
		}
		return lines
	case *pb.LiteralNodeProto_FeatureValue:
		return fillLinesFromFeatureProto(v.FeatureValue, lines)
	}
	return append(lines, fillSpansFromLiteralProto(l, LineJSON{}))
}

func fillLinesFromFeatureProto(f *pb.FeatureProto, lines []LineJSON) []LineJSON {
	switch f := f.Feature.(type) {
	case *pb.FeatureProto_Point:
		lines = append(lines, fillSpansFromFeatureIDProto(f.Point.Id, LineJSON{}))
		lines = fillLinesFromTagProtos(f.Point.Tags, lines)
	case *pb.FeatureProto_Path:
		lines = append(lines, fillSpansFromFeatureIDProto(f.Path.Id, LineJSON{}))
		lines = fillLinesFromTagProtos(f.Path.Tags, lines)
	case *pb.FeatureProto_Area:
		lines = append(lines, fillSpansFromFeatureIDProto(f.Area.Id, LineJSON{}))
		lines = fillLinesFromTagProtos(f.Area.Tags, lines)
	case *pb.FeatureProto_Relation:
		lines = append(lines, fillSpansFromFeatureIDProto(f.Relation.Id, LineJSON{}))
		lines = fillLinesFromTagProtos(f.Relation.Tags, lines)
	}
	return lines
}

func fillLinesFromProto(p *pb.NodeProto, lines []LineJSON) []LineJSON {
	switch n := p.Node.(type) {
	case *pb.NodeProto_Symbol:
		return append(lines, LineJSON{SpanJSON{Text: n.Symbol}})
	case *pb.NodeProto_Call:
		return append(lines, LineJSON{SpanJSON{Text: fmt.Sprintf("%+v", p)}})
	case *pb.NodeProto_Literal:
		return fillLinesFromLiteralProto(n.Literal, lines)
	}
	return append(lines, LineJSON{SpanJSON{Text: fmt.Sprintf("%+v", p)}})
}

func fillLinesFromFeature(f b6.Feature, lines []LineJSON) []LineJSON {
	lines = append(lines, LineJSON{SpanJSON{Text: "/" + f.FeatureID().String()}})
	for _, tag := range f.AllTags() {
		lines = append(lines, LineJSON{SpanJSON{Text: tag.Key}, SpanJSON{Text: " = "}, SpanJSON{Text: tag.Value}})
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

func unmarshalGeoJSON(p *pb.LiteralNodeProto_GeoJSONValue) (geojson.GeoJSON, error) {
	r, err := gzip.NewReader(bytes.NewBuffer(p.GeoJSONValue))
	if err != nil {
		return nil, err
	}
	var j []byte
	j, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return geojson.Unmarshal(j)
}

var Replacements = []string{"†", "‡", "§"}

func compressLine(line LineJSON) LineJSON {
	seen := make(map[SpanJSON]struct{})
	replacements := make(map[SpanJSON]int)
	for _, span := range line {
		if _, ok := seen[span]; ok {
			if len(replacements)+1 >= len(Replacements) {
				break
			}
			replacements[span] = len(replacements)
		} else {
			seen[span] = struct{}{}
		}
	}
	if len(replacements) == 0 {
		return line
	}

	seen = make(map[SpanJSON]struct{})
	replaced := make(LineJSON, 0, len(line)+len(replacements))
	for _, span := range line {
		if r, ok := replacements[span]; ok {
			if _, ok = seen[span]; ok {
				replaced = append(replaced, SpanJSON{Class: "reference", Text: Replacements[r]})
			} else {
				replaced = append(replaced, span, SpanJSON{Class: "marker", Text: Replacements[r]})
			}
			seen[span] = struct{}{}
		} else {
			replaced = append(replaced, span)
		}
	}
	return replaced
}

func compressLines(lines []LineJSON) []LineJSON {
	compressed := make([]LineJSON, len(lines))
	for i, line := range lines {
		compressed[i] = compressLine(line)
	}
	return compressed
}

func showConnectivity(id b6.Identifiable, c *api.Context) (geojson.GeoJSON, error) {
	f := api.Resolve(id, c.World)
	if f == nil {
		return nil, fmt.Errorf("No feature with ID %s", id)
	}
	features := geojson.NewFeatureCollection()
	var s *graph.ShortestPathSearch
	weights := graph.SimpleHighwayWeights{}
	switch f := f.(type) {
	case b6.PointFeature:
		s = graph.NewShortestPathSearchFromPoint(f.PointID())
		features.AddFeature(b6.PointFeatureToGeoJSON(f))
	case b6.AreaFeature:
		s = graph.NewShortestPathSearchFromBuilding(f, weights, c.World)
		features.AddFeature(b6.AreaFeatureToGeoJSON(f))
	default:
		return features, nil
	}
	s.ExpandSearch(1000.0, weights, graph.PointsAndAreas, c.World)
	for key, state := range s.PathStates() {
		segment := b6.FindPathSegmentByKey(key, c.World)
		polyline := segment.Polyline()
		geometry := geojson.GeometryFromLineString(geojson.FromPolyline(polyline))
		shape := geojson.NewFeatureWithGeometry(geometry)
		features.AddFeature(shape)
		label := geojson.NewFeatureFromS2Point(polyline.Centroid())
		switch state {
		case graph.PathStateTraversed:
			shape.Properties["-diagonal-stroke"] = "#00ff00"
		case graph.PathStateTooFar:
			shape.Properties["-diagonal-stroke"] = "#ff0000"
			features.AddFeature(label)
		case graph.PathStateNotUseable:
			shape.Properties["-diagonal-stroke"] = "#ff0000"
			features.AddFeature(label)
		}
		features.AddFeature(shape)
	}
	return features, nil
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
	if c.Lines != nil {
		r.Lines = append(r.Lines, c.Lines...)
	} else {
		r.Lines = []LineJSON{}
	}
	return nil
}

func show(v interface{}, c *api.Context) (UIChange, error) {
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
		return &showChange{
			BoundariesLayer: boundaries,
			PointsLayer:     points,
		}, nil
	case b6.Feature:
		return &showChange{
			GeoJSON: v.ToGeoJSON(),
			Lines:   compressLines(fillLinesFromFeature(v, []LineJSON{})),
		}, nil
	case b6.Renderable:
		return &showChange{
			GeoJSON: v.ToGeoJSON(),
		}, nil
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
		return &showChange{
			GeoJSON: g,
		}, nil
	}
	return &showChange{}, nil
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
		gp := geojson.FromS2Point(p.Point())
		response.Center = &gp
	}

	if node, err := api.ToProto(result); err == nil {
		if l := node.GetLiteral(); l != nil {
			if g, ok := l.Value.(*pb.LiteralNodeProto_GeoJSONValue); ok {
				response.GeoJSON, _ = unmarshalGeoJSON(g)
			}
		}
		response.Lines = compressLines(fillLinesFromProto(node, response.Lines))
	} else {
		response.Lines = append(response.Lines, LineJSON{SpanJSON{Class: "error", Text: err.Error()}})
	}
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
