package geojson

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	pb "diagonal.works/b6/proto"
	"github.com/golang/geo/s2"
)

// GeoJSON is implemented by all GeoJSON objects, specifically, Feature,
// FeatureCollection and Geometry.

type GeometryMapFunction func(c Coordinates) (Coordinates, error)

func MapS2Polygons(f func(p *s2.Polygon) ([]*s2.Polygon, error)) GeometryMapFunction {
	return func(c Coordinates) (Coordinates, error) {
		switch cc := c.(type) {
		case Polygon:
			if ps, err := f(cc.ToS2Polygon()); err == nil {
				if len(ps) == 1 {
					return FromPolygon(ps[0]), nil
				} else {
					return FromPolygons(ps), nil
				}
			} else {
				return nil, err
			}
		case MultiPolygon:
			mapped := make([]*s2.Polygon, 0, len(cc))
			for _, p := range cc.ToS2Polygons() {
				if ps, err := f(p); err == nil {
					for _, p := range ps {
						mapped = append(mapped, p)
					}
				} else {
					return nil, err
				}
			}
			return FromPolygons(mapped), nil
		default:
			return cc, nil
		}
	}
}

type GeoJSON interface {
	MapGeometries(f GeometryMapFunction) (GeoJSON, error)
	ToS2Polygons() []*s2.Polygon
	Centroid() Point
}

type Coordinate struct {
	Lat float64
	Lng float64
}

func CoordinateFromS2LatLng(ll s2.LatLng) Coordinate {
	return Coordinate{Lat: ll.Lat.Degrees(), Lng: ll.Lng.Degrees()}
}

func CoordinateFromS2Point(point s2.Point) Coordinate {
	return CoordinateFromS2LatLng(s2.LatLngFromPoint(point))
}

func (c *Coordinate) ToS2Point() s2.Point {
	return s2.PointFromLatLng(s2.LatLngFromDegrees(c.Lat, c.Lng))
}

func CoordinatesFromLoop(loop *s2.Loop) []Coordinate {
	coordinates := make([]Coordinate, loop.NumVertices()+1)
	for i := range coordinates {
		coordinates[i] = CoordinateFromS2Point(loop.Vertex(i % loop.NumVertices()))
	}
	return coordinates
}

func FromPolygon(polygon *s2.Polygon) Polygon {
	coordinates := make(Polygon, polygon.NumLoops())
	for i, loop := range polygon.Loops() {
		coordinates[i] = make([]Coordinate, loop.NumVertices()+1)
		if !loop.IsHole() {
			for j := range coordinates[i] {
				coordinates[i][j] = CoordinateFromS2Point(loop.Vertex(j % loop.NumVertices()))
			}
		} else {
			for j := range coordinates[i] {
				coordinates[i][len(coordinates[i])-j-1] = CoordinateFromS2Point(loop.Vertex(j % loop.NumVertices()))
			}
		}
	}
	return coordinates
}

func FromPolygons(polygons []*s2.Polygon) MultiPolygon {
	coordinates := make(MultiPolygon, len(polygons))
	for i, p := range polygons {
		coordinates[i] = FromPolygon(p)
	}
	return coordinates
}

func (c Coordinate) ToS2LatLng() s2.LatLng {
	return s2.LatLngFromDegrees(c.Lat, c.Lng)
}

func (c Coordinate) MarshalJSON() ([]byte, error) {
	return json.Marshal([]float64{c.Lng, c.Lat})
}

func (c Coordinate) String() string {
	marshalled, _ := c.MarshalJSON()
	return string(marshalled)
}

func (c *Coordinate) UnmarshalJSON(buffer []byte) error {
	var points []float64
	if err := json.Unmarshal(buffer, &points); err != nil {
		return err
	}
	if len(points) != 2 {
		return fmt.Errorf("Expected 2 floats, found: %q", buffer)
	}
	c.Lng = points[0]
	c.Lat = points[1]
	return nil
}

type Coordinates interface {
	EachCoordinate(func(Coordinate))
	Centroid() Point
}

type Point Coordinate

func FromS2LatLng(ll s2.LatLng) Point {
	return Point{Lat: ll.Lat.Degrees(), Lng: ll.Lng.Degrees()}
}

func FromS2Point(point s2.Point) Point {
	return FromS2LatLng(s2.LatLngFromPoint(point))
}

func (p Point) ToS2LatLng() s2.LatLng {
	return Coordinate(p).ToS2LatLng()
}

func (p Point) ToS2Point() s2.Point {
	return s2.PointFromLatLng(p.ToS2LatLng())
}

func (p Point) MarshalJSON() ([]byte, error) {
	return Coordinate(p).MarshalJSON()
}

func (p Point) EachCoordinate(f func(Coordinate)) {
	f(Coordinate(p))
}

func (p *Point) UnmarshalJSON(buffer []byte) error {
	return ((*Coordinate)(p)).UnmarshalJSON(buffer)
}

func (p Point) Centroid() Point {
	return p
}

type MultiPoint []Coordinate

func (m MultiPoint) EachCoordinate(f func(Coordinate)) {
	for _, coordinate := range m {
		f(coordinate)
	}
}

func (m MultiPoint) Centroid() Point {
	query := s2.NewConvexHullQuery()
	for _, coordinate := range m {
		query.AddPoint(coordinate.ToS2Point())
	}
	return FromS2Point(s2.Point{Vector: query.ConvexHull().Centroid().Normalize()})
}

type LineString []Coordinate

func FromPolyline(polyline *s2.Polyline) LineString {
	coordinates := make(LineString, len(*polyline))
	for i, point := range *polyline {
		coordinates[i] = CoordinateFromS2Point(point)
	}
	return coordinates
}

func (l LineString) EachCoordinate(f func(Coordinate)) {
	for _, coordinate := range l {
		f(coordinate)
	}
}

func (l LineString) ToS2Polyline() s2.Polyline {
	points := make(s2.Polyline, len(l))
	for i, point := range l {
		points[i] = point.ToS2Point()
	}
	return points
}

func (l LineString) ToS2Loop() *s2.Loop {
	return s2.LoopFromPoints(l.ToS2Polyline())
}

func (l LineString) Centroid() Point {
	p := l.ToS2Polyline()
	return FromS2Point(p.Centroid())
}

type Polygon [][]Coordinate

func (p Polygon) EachCoordinate(f func(Coordinate)) {
	for _, loop := range p {
		for _, coordinate := range loop {
			f(coordinate)
		}
	}
}

func (p Polygon) ToS2Polygon() *s2.Polygon {
	loops := make([]*s2.Loop, len(p))
	for i, loop := range p {
		loops[i] = LineString(loop).ToS2Loop()
		if i%2 == 1 {
			loops[i].Invert()
		}
	}
	return s2.PolygonFromLoops(loops)
}

func (p Polygon) Centroid() Point {
	if len(p) > 0 {
		return FromS2Point(LineString(p[0]).ToS2Loop().Centroid())
	}
	return Point{}
}

type MultiLineString [][]Coordinate

func (m MultiLineString) EachCoordinate(f func(Coordinate)) {
	for _, line := range m {
		for _, coordinate := range line {
			f(coordinate)
		}
	}
}

func (m MultiLineString) Centroid() Point {
	query := s2.NewConvexHullQuery()
	for _, line := range m {
		for _, coordinate := range line {
			query.AddPoint(coordinate.ToS2Point())
		}
	}
	return FromS2Point(s2.Point{Vector: query.ConvexHull().Centroid().Normalize()})
}

type MultiPolygon [][][]Coordinate

func (m MultiPolygon) EachCoordinate(f func(Coordinate)) {
	for _, polygon := range m {
		for _, loop := range polygon {
			for _, coordinate := range loop {
				f(coordinate)
			}
		}
	}
}

func (m MultiPolygon) ToS2Polygons() []*s2.Polygon {
	r := make([]*s2.Polygon, len(m))
	for i, polygon := range m {
		r[i] = Polygon(polygon).ToS2Polygon()
	}
	return r
}

func (m MultiPolygon) Centroid() Point {
	query := s2.NewConvexHullQuery()
	for _, polygon := range m {
		for _, loop := range polygon {
			for _, coordinate := range loop {
				query.AddPoint(coordinate.ToS2Point())
			}
		}
	}
	return FromS2Point(s2.Point{Vector: query.ConvexHull().Centroid().Normalize()})
}

type Geometry struct {
	Type        string      `json:"type"`
	Coordinates Coordinates `json:"coordinates"`
}

func GeometryFromCoordinates(c Coordinates) Geometry {
	switch c.(type) {
	case Point:
		return Geometry{Type: "Point", Coordinates: c}
	case MultiPoint:
		return Geometry{Type: "MultiPoint", Coordinates: c}
	case LineString:
		return Geometry{Type: "LineString", Coordinates: c}
	case MultiLineString:
		return Geometry{Type: "MultiLineString", Coordinates: c}
	case Polygon:
		return Geometry{Type: "Polygon", Coordinates: c}
	case MultiPolygon:
		return Geometry{Type: "MultiPolygon", Coordinates: c}
	}
	panic(fmt.Sprintf("unknown geometry type %T", c))
}

func (g *Geometry) UnmarshalJSON(buffer []byte) error {
	t := struct{ Type string }{}
	json.Unmarshal(buffer, &t)
	g.Type = t.Type
	switch g.Type {
	case "Point":
		point := struct{ Coordinates Point }{}
		if err := json.Unmarshal(buffer, &point); err != nil {
			return err
		}
		g.Coordinates = point.Coordinates
	case "LineString":
		lineString := struct{ Coordinates LineString }{}
		if err := json.Unmarshal(buffer, &lineString); err != nil {
			return err
		}
		g.Coordinates = lineString.Coordinates
	case "Polygon":
		polygon := struct{ Coordinates Polygon }{}
		if err := json.Unmarshal(buffer, &polygon); err != nil {
			return err
		}
		g.Coordinates = polygon.Coordinates
	case "MultiPoint":
		multiPoint := struct{ Coordinates MultiPoint }{}
		if err := json.Unmarshal(buffer, &multiPoint); err != nil {
			return err
		}
		g.Coordinates = multiPoint.Coordinates
	case "MultiLineString":
		multiLineString := struct{ Coordinates MultiLineString }{}
		if err := json.Unmarshal(buffer, &multiLineString); err != nil {
			return err
		}
		g.Coordinates = multiLineString.Coordinates
	case "MultiPolygon":
		multiPolygon := struct{ Coordinates MultiPolygon }{}
		if err := json.Unmarshal(buffer, &multiPolygon); err != nil {
			return err
		}
		g.Coordinates = multiPolygon.Coordinates
	default:
		return fmt.Errorf("Can't unmarshal geometry with type %q", g.Type)
	}
	return nil
}

func (g Geometry) MapGeometryGeometries(f GeometryMapFunction) (Geometry, error) {
	if mapped, err := f(g.Coordinates); err == nil {
		return GeometryFromCoordinates(mapped), nil
	} else {
		return Geometry{}, err
	}
}

func (g Geometry) MapGeometries(f GeometryMapFunction) (GeoJSON, error) {
	return g.MapGeometryGeometries(f)
}

func (g Geometry) ToS2Polygons() []*s2.Polygon {
	switch c := g.Coordinates.(type) {
	case Polygon:
		return []*s2.Polygon{c.ToS2Polygon()}
	case MultiPolygon:
		return c.ToS2Polygons()
	default:
		return []*s2.Polygon{}
	}
}

func (g Geometry) Centroid() Point {
	return g.Coordinates.Centroid()
}

var _ GeoJSON = &Geometry{}

func GeometryFromPoint(point Point) Geometry {
	return Geometry{Type: "Point", Coordinates: point}
}

func GeometryFromS2Point(point s2.Point) Geometry {
	return Geometry{Type: "Point", Coordinates: Point(CoordinateFromS2Point(point))}
}

func GeometryFromS2LatLng(ll s2.LatLng) Geometry {
	return Geometry{Type: "Point", Coordinates: Point(CoordinateFromS2LatLng(ll))}
}

func GeometryFromLineString(lineString []Coordinate) Geometry {
	return Geometry{Type: "LineString", Coordinates: LineString(lineString)}
}

func GeometryFromS2Edge(edge s2.Edge) Geometry {
	coordinates := make([]Coordinate, 2)
	coordinates[0] = CoordinateFromS2Point(edge.V0)
	coordinates[1] = CoordinateFromS2Point(edge.V1)
	return GeometryFromLineString(coordinates)
}

func GeometryFromS2Polyline(polyline s2.Polyline) Geometry {
	coordinates := make([]Coordinate, len(polyline))
	for i, point := range polyline {
		coordinates[i] = CoordinateFromS2Point(point)
	}
	return GeometryFromLineString(coordinates)
}

func GeometryFromS2Loop(loop *s2.Loop) Geometry {
	return Geometry{Type: "Polygon", Coordinates: Polygon([][]Coordinate{CoordinatesFromLoop(loop)})}
}

func GeometryFromPolygon(polygon Polygon) Geometry {
	return Geometry{Type: "Polygon", Coordinates: Polygon(polygon)}
}

func GeometryFromS2Polygon(polygon *s2.Polygon) Geometry {
	return GeometryFromPolygon(FromPolygon(polygon))
}

func GeometryFromS2Polygons(polygons []*s2.Polygon) Geometry {
	return GeometryFromMultiPolygon(FromPolygons(polygons))
}

func GeometryFromMultiLineString(multiLineString [][]Coordinate) Geometry {
	return Geometry{Type: "MultiLineString", Coordinates: MultiLineString(multiLineString)}
}

func GeometryFromMultiPolygon(multiPolygon MultiPolygon) Geometry {
	return Geometry{Type: "MultiPolygon", Coordinates: MultiPolygon(multiPolygon)}
}

type Feature struct {
	Type       string            `json:"type"`
	Geometry   Geometry          `json:"geometry"`
	Properties map[string]string `json:"properties"`
}

func NewFeature() *Feature {
	return &Feature{Type: "Feature", Properties: make(map[string]string)}
}

func NewFeatureFromPoint(point Point) *Feature {
	return NewFeatureWithGeometry(GeometryFromPoint(point))
}

func NewFeatureFromS2Point(point s2.Point) *Feature {
	return NewFeatureWithGeometry(GeometryFromS2Point(point))
}

func NewFeatureFromS2Edge(edge s2.Edge) *Feature {
	return NewFeatureWithGeometry(GeometryFromS2Edge(edge))
}

func NewFeatureFromS2Polyline(polyline s2.Polyline) *Feature {
	return NewFeatureWithGeometry(GeometryFromS2Polyline(polyline))
}

func NewFeatureFromS2Loop(loop *s2.Loop) *Feature {
	return NewFeatureWithGeometry(GeometryFromS2Loop(loop))
}

func NewFeatureFromS2Polygon(polygon *s2.Polygon) *Feature {
	return NewFeatureWithGeometry(GeometryFromS2Polygon(polygon))
}

func NewFeatureFromS2LatLng(ll s2.LatLng) *Feature {
	return NewFeatureWithGeometry(GeometryFromS2LatLng(ll))
}

func NewFeatureFromS2Cell(cell s2.Cell) *Feature {
	loop := s2.LoopFromPoints([]s2.Point{cell.Vertex(0), cell.Vertex(1), cell.Vertex(2), cell.Vertex(3)})
	return NewFeatureWithGeometry(GeometryFromS2Loop(loop))
}

func NewFeatureFromS2CellID(id s2.CellID) *Feature {
	return NewFeatureFromS2Cell(s2.CellFromCellID(id))
}

func NewFeatureWithGeometry(geometry Geometry) *Feature {
	feature := NewFeature()
	feature.Geometry = geometry
	return feature
}

func (f *Feature) String() string {
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Sprintf("json.MarshalIndent: %s", err)
	}
	return string(b)
}

func (f *Feature) MapFeatureGeometries(m GeometryMapFunction) (*Feature, error) {
	if mapped, err := f.Geometry.MapGeometryGeometries(m); err == nil {
		return &Feature{
			Type:       "Feature",
			Geometry:   mapped,
			Properties: f.Properties,
		}, nil
	} else {
		return nil, err
	}
}

func (f *Feature) MapGeometries(m GeometryMapFunction) (GeoJSON, error) {
	return f.MapFeatureGeometries(m)
}

func (f *Feature) ToS2Polygons() []*s2.Polygon {
	return f.Geometry.ToS2Polygons()
}

func (f *Feature) Centroid() Point {
	return f.Geometry.Centroid()
}

var _ GeoJSON = &Feature{}

type FeatureCollection struct {
	Type     string     `json:"type"`
	Features []*Feature `json:"features"`
}

func NewFeatureCollection() *FeatureCollection {
	return &FeatureCollection{Type: "FeatureCollection", Features: make([]*Feature, 0, 1)}
}

func (c *FeatureCollection) Add(g GeoJSON) {
	switch g := g.(type) {
	case *Geometry:
		c.AddFeature(NewFeatureWithGeometry(*g))
	case *Feature:
		c.AddFeature(g)
	case *FeatureCollection:
		for _, f := range g.Features {
			c.AddFeature(f)
		}
	default:
		panic(fmt.Sprintf("not a GeoJSON object: %T", g))
	}
}

func (c *FeatureCollection) AddFeature(feature *Feature) {
	c.Features = append(c.Features, feature)
}

func (c *FeatureCollection) WriteToFile(filename string) error {
	output, err := json.Marshal(c)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	f.Write(output)
	return f.Close()
}

func (c *FeatureCollection) MapFeatureCollectionGeometries(f GeometryMapFunction) (*FeatureCollection, error) {
	features := make([]*Feature, len(c.Features))
	for i, feature := range c.Features {
		if mapped, err := feature.MapFeatureGeometries(f); err == nil {
			features[i] = mapped
		} else {
			return nil, err
		}
	}
	return &FeatureCollection{Type: "FeatureCollection", Features: features}, nil
}

func (c *FeatureCollection) MapGeometries(f GeometryMapFunction) (GeoJSON, error) {
	return c.MapFeatureCollectionGeometries(f)
}

func (c *FeatureCollection) Centroid() Point {
	query := s2.NewConvexHullQuery()
	for _, feature := range c.Features {
		feature.Geometry.Coordinates.EachCoordinate(func(c Coordinate) {
			query.AddPoint(c.ToS2Point())
		})
	}
	return FromS2Point(s2.Point{Vector: query.ConvexHull().Centroid().Normalize()})
}

func (f *FeatureCollection) ToS2Polygons() []*s2.Polygon {
	ps := make([]*s2.Polygon, 0, len(f.Features))
	for _, feature := range f.Features {
		ps = append(ps, feature.ToS2Polygons()...)
	}
	return ps
}

var _ GeoJSON = &FeatureCollection{}

func ReadFromFile(filename string) (*FeatureCollection, error) {
	var f FeatureCollection
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &f)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func Unmarshal(buffer []byte) (GeoJSON, error) {
	t := struct{ Type string }{}
	json.Unmarshal(buffer, &t)
	switch t.Type {
	case "FeatureCollection":
		var collection FeatureCollection
		if err := json.Unmarshal(buffer, &collection); err != nil {
			return nil, err
		}
		return &collection, nil
	case "Feature":
		var feature Feature
		if err := json.Unmarshal(buffer, &feature); err != nil {
			return nil, err
		}
		return &feature, nil
	case "Point", "MultiPoint", "LineString", "Polygon", "MultiPolygon":
		var geometry Geometry
		if err := json.Unmarshal(buffer, &geometry); err != nil {
			return nil, err
		}
		return &geometry, nil
	}
	return nil, fmt.Errorf("Can't unmarshal GeoJSON type %q", t.Type)
}

func PolygonProtoToGeoJSON(polygon *pb.PolygonProto) *Feature {
	if polygon == nil {
		polygon = &pb.PolygonProto{Loops: []*pb.LoopProto{}}
	}
	loops := make([][]Coordinate, len(polygon.Loops), len(polygon.Loops))
	for i, loop := range polygon.Loops {
		loops[i] = LoopProtoToGeoJSON(loop)
	}
	return NewFeatureWithGeometry(GeometryFromPolygon(loops))
}

func LoopProtoToGeoJSON(loop *pb.LoopProto) []Coordinate {
	coordinates := make([]Coordinate, len(loop.Points)+1, len(loop.Points)+1)
	for i := range coordinates {
		coordinates[i] = PointProtoToGeoJSON(loop.Points[i%len(loop.Points)])
	}
	return coordinates
}

func PointProtoToGeoJSON(point *pb.PointProto) Coordinate {
	return Coordinate{
		Lat: float64(point.LatE7) / 1e7,
		Lng: float64(point.LngE7) / 1e7,
	}
}
