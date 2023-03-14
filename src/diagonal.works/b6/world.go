package b6

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"diagonal.works/b6/geojson"
	"diagonal.works/b6/search"
	"diagonal.works/b6/units"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

type Renderable interface {
	ToGeoJSON() geojson.GeoJSON
}

type Tag struct {
	Key   string
	Value string
}

func (t Tag) IsValid() bool {
	return t.Key != ""
}

func (t Tag) IntValue() (int, bool) {
	if i, err := strconv.Atoi(t.Value); err == nil {
		return i, true
	}
	return 0, false
}

func (t Tag) FloatValue() (float64, bool) {
	if f, err := strconv.ParseFloat(t.Value, 64); err == nil {
		return f, true
	}
	return 0.0, false
}

func InvalidTag() Tag {
	return Tag{}
}

type Tagged interface {
	AllTags() []Tag
	Get(key string) Tag
}

type FeatureType int

const (
	FeatureTypePoint FeatureType = iota
	FeatureTypePath
	FeatureTypeArea
	FeatureTypeRelation
	FeatureTypeInvalid

	FeatureTypeBegin = FeatureTypePoint
	FeatureTypeEnd   = FeatureTypeInvalid
	FeatureTypeBits  = 2 // Bits necessary to represent all types except invalid
)

func (f FeatureType) String() string {
	switch f {
	case FeatureTypePoint:
		return "point"
	case FeatureTypePath:
		return "path"
	case FeatureTypeArea:
		return "area"
	case FeatureTypeRelation:
		return "relation"
	default:
		return "invalid"
	}
}

func FeatureTypeFromString(s string) FeatureType {
	for t := FeatureTypeBegin; t < FeatureTypeEnd; t++ {
		if s == t.String() {
			return t
		}
	}
	return FeatureTypeInvalid
}

type FeatureTypeIndex interface {
	RewriteFeatureTypeQuery(query FeatureTypeQuery) search.Query
}

type FeatureTypeQuery struct {
	Type  FeatureType
	Query search.Query
}

func (f FeatureTypeQuery) String() string {
	return fmt.Sprintf("(feature-type %s %s)", f.Type.String(), f.Query.String())
}

func (f FeatureTypeQuery) Compile(index search.Index) search.Iterator {
	if featureTypeIndex, ok := index.(FeatureTypeIndex); ok {
		return featureTypeIndex.RewriteFeatureTypeQuery(f).Compile(index)
	}
	log.Printf("Using expensive All query for %s", f.String())
	return search.All{Token: search.AllToken}.Compile(index)

}

type Namespace string

func (n Namespace) String() string {
	return string(n)
}

type Namespaces []Namespace

func (n Namespaces) Len() int           { return len(n) }
func (n Namespaces) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n Namespaces) Less(i, j int) bool { return n[i] < n[j] }

const (
	// Features from OSM
	NamespaceOSMNode     Namespace = "openstreetmap.org/node"
	NamespaceOSMWay      Namespace = "openstreetmap.org/way"
	NamespaceOSMRelation Namespace = "openstreetmap.org/relation"

	// Used internally
	NamespacePrivate Namespace = "diagonal.works/ns/private"
	NamespaceLatLng  Namespace = "diagonal.works/ns/ll"

	// Used when connecting features to the street network
	NamespaceDiagonalEntrances    Namespace = "diagonal.works/ns/entrance"
	NamespaceDiagonalAccessPaths  Namespace = "diagonal.works/ns/access-path"
	NamespaceDiagonalAccessPoints Namespace = "diagonal.works/ns/access-point"

	NamespaceUKONSBoundaries       Namespace = "statistics.gov.uk/datasets/regions"
	NamespaceGBUPRN                Namespace = "ordnancesurvey.co.uk/open-uprn"
	NamespaceGBOSTerrain50Contours Namespace = "ordnancesurvey.co.uk/terrain-50/contours"

	NamespaceInvalid Namespace = ""
)

var StandardNamespaces = []Namespace{
	NamespaceOSMNode,
	NamespaceOSMWay,
	NamespaceOSMRelation,
	NamespaceLatLng,
	NamespaceDiagonalEntrances,
	NamespaceDiagonalAccessPaths,
	NamespaceDiagonalAccessPoints,
}

var OSMNamespaces = []Namespace{
	NamespaceOSMNode,
	NamespaceOSMWay,
	NamespaceOSMRelation,
}

type Identifiable interface {
	FeatureID() FeatureID
}

type IdentifiablePoint interface {
	PointID() PointID
}

type FeatureID struct {
	Type      FeatureType
	Namespace Namespace
	Value     uint64
}

func (f FeatureID) IsValid() bool {
	return f.Namespace != NamespaceInvalid && f.Type != FeatureTypeInvalid
}

func (f FeatureID) Less(other FeatureID) bool {
	if f.Type == other.Type {
		if f.Namespace == other.Namespace {
			return f.Value < other.Value
		} else {
			return f.Namespace < other.Namespace
		}
	} else {
		return f.Type < other.Type
	}
}

func (f FeatureID) String() string {
	return fmt.Sprintf("%s/%s/%d", f.Type.String(), f.Namespace, f.Value)
}

func (f FeatureID) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

func (f *FeatureID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*f = FeatureIDFromString(s)
	return nil
}

func (f FeatureID) FeatureID() FeatureID {
	return f
}

func (f FeatureID) ToPointID() PointID {
	if f.Type == FeatureTypePoint || f.Type == FeatureTypeInvalid {
		return PointID{Namespace: f.Namespace, Value: f.Value}
	}
	panic("Not a point")
}

func (f FeatureID) ToPathID() PathID {
	if f.Type == FeatureTypePath || f.Type == FeatureTypeInvalid {
		return PathID{Namespace: f.Namespace, Value: f.Value}
	}
	panic("Not a path")
}

func (f FeatureID) ToAreaID() AreaID {
	if f.Type == FeatureTypeArea || f.Type == FeatureTypeInvalid {
		return AreaID{Namespace: f.Namespace, Value: f.Value}
	}
	panic("Not a area")
}

func (f FeatureID) ToRelationID() RelationID {
	if f.Type == FeatureTypeRelation || f.Type == FeatureTypeInvalid {
		return RelationID{Namespace: f.Namespace, Value: f.Value}
	}
	panic("Not a relation")
}

func FeatureIDFromString(s string) FeatureID {
	i := strings.Index(s, "/")
	j := strings.LastIndex(s, "/")
	if i < 0 || i == j {
		return FeatureIDInvalid
	}
	id := FeatureID{Type: FeatureTypeFromString(s[0:i])}
	if id.Type == FeatureTypeInvalid {
		return FeatureIDInvalid
	}
	id.Namespace = Namespace(s[i+1 : j])
	if v, err := strconv.ParseUint(s[j+1:], 10, 64); err == nil {
		id.Value = v
	} else {
		return FeatureIDInvalid
	}
	return id
}

var (
	FeatureIDInvalid = FeatureID{Type: FeatureTypeInvalid, Namespace: NamespaceInvalid}
	FeatureIDEnd     = FeatureID{Type: FeatureTypeEnd, Namespace: NamespaceInvalid} // For sentinels

	FeatureIDPointBegin    = FeatureID{Type: FeatureTypePoint, Namespace: NamespaceInvalid, Value: 0}
	FeatureIDPathBegin     = FeatureID{Type: FeatureTypePath, Namespace: NamespaceInvalid, Value: 0}
	FeatureIDPointEnd      = FeatureIDPathBegin
	FeatureIDAreaBegin     = FeatureID{Type: FeatureTypeArea, Namespace: NamespaceInvalid, Value: 0}
	FeatureIDPathEnd       = FeatureIDAreaBegin
	FeatureIDRelationBegin = FeatureID{Type: FeatureTypeRelation, Namespace: NamespaceInvalid, Value: 0}
	FeatureIDAreaEnd       = FeatureIDRelationBegin
	FeatureIDRelationEnd   = FeatureIDEnd
)

type FeatureIDs []FeatureID

func (f FeatureIDs) Len() int           { return len(f) }
func (f FeatureIDs) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f FeatureIDs) Less(i, j int) bool { return f[i].Less(f[j]) }

type PointID struct {
	Namespace Namespace
	Value     uint64
}

func MakePointID(ns Namespace, v uint64) PointID {
	return PointID{Namespace: ns, Value: v}
}

func (p PointID) FeatureID() FeatureID {
	return FeatureID{Type: FeatureTypePoint, Namespace: p.Namespace, Value: p.Value}
}

func (p PointID) IsValid() bool {
	return p.Namespace != NamespaceInvalid
}

func (p PointID) String() string {
	return p.FeatureID().String()
}

func (p PointID) Less(other PointID) bool {
	if p.Namespace == other.Namespace {
		return p.Value < other.Value
	} else {
		return p.Namespace < other.Namespace
	}
}

var PointIDInvalid = PointID{Namespace: NamespaceInvalid}

type PathID struct {
	Namespace Namespace
	Value     uint64
}

func MakePathID(ns Namespace, v uint64) PathID {
	return PathID{Namespace: ns, Value: v}
}

func (p PathID) FeatureID() FeatureID {
	return FeatureID{Type: FeatureTypePath, Namespace: p.Namespace, Value: p.Value}
}

func (p PathID) IsValid() bool {
	return p.Namespace != NamespaceInvalid
}

func (p PathID) String() string {
	return p.FeatureID().String()
}

func (p PathID) Less(other PathID) bool {
	if p.Namespace == other.Namespace {
		return p.Value < other.Value
	} else {
		return p.Namespace < other.Namespace
	}
}

var PathIDInvalid = PathID{Namespace: NamespaceInvalid}

type AreaID struct {
	Namespace Namespace
	Value     uint64
}

func MakeAreaID(ns Namespace, v uint64) AreaID {
	return AreaID{Namespace: ns, Value: v}
}

func (a AreaID) FeatureID() FeatureID {
	return FeatureID{Type: FeatureTypeArea, Namespace: a.Namespace, Value: a.Value}
}

func (a AreaID) IsValid() bool {
	return a.Namespace != NamespaceInvalid
}

func (a AreaID) String() string {
	return a.FeatureID().String()
}

func (a AreaID) Less(other PointID) bool {
	if a.Namespace == other.Namespace {
		return a.Value < other.Value
	} else {
		return a.Namespace < other.Namespace
	}
}

var AreaIDInvalid = AreaID{Namespace: NamespaceInvalid}

type RelationID struct {
	Namespace Namespace
	Value     uint64
}

func MakeRelationID(ns Namespace, v uint64) RelationID {
	return RelationID{Namespace: ns, Value: v}
}

func (r RelationID) FeatureID() FeatureID {
	return FeatureID{Type: FeatureTypeRelation, Namespace: r.Namespace, Value: r.Value}
}

func (r RelationID) IsValid() bool {
	return r.Namespace != NamespaceInvalid
}

func (r RelationID) String() string {
	return r.FeatureID().String()
}

func (r RelationID) Less(other PointID) bool {
	if r.Namespace == other.Namespace {
		return r.Value < other.Value
	} else {
		return r.Namespace < other.Namespace
	}
}

var RelationIDInvalid = RelationID{Namespace: NamespaceInvalid}

type Geometry interface {
	Covering(coverer s2.RegionCoverer) s2.CellUnion
	ToGeoJSON() geojson.GeoJSON
}

type Feature interface {
	FeatureID() FeatureID
	Tagged
	Renderable
}

type PhysicalFeature interface {
	Feature
	Geometry
}

func Center(feature PhysicalFeature) s2.Point {
	// TODO: Cache the values in precomputed indicies?
	switch feature := feature.(type) {
	case PointFeature:
		return feature.Point()
	case PathFeature:
		return s2.Point{Vector: feature.Polyline().Centroid().Normalize()}
	case AreaFeature:
		if feature.Len() == 1 {
			return s2.Point{Vector: feature.Polygon(0).Loop(0).Centroid().Normalize()}
		} else {
			query := s2.NewConvexHullQuery()
			for i := 0; i < feature.Len(); i++ {
				query.AddPolygon(feature.Polygon(i))
			}
			return s2.Point{Vector: query.ConvexHull().Centroid().Normalize()}
		}
	}
	return s2.Point{}
}

type Point interface {
	Geometry
	Point() s2.Point
	CellID() s2.CellID
}

type point struct {
	p s2.Point
}

func (p point) Point() s2.Point {
	return p.p
}

func (p point) CellID() s2.CellID {
	return s2.CellIDFromLatLng(s2.LatLngFromPoint(p.p))
}

func (p point) Covering(coverer s2.RegionCoverer) s2.CellUnion {
	return coverer.Covering(p.p)
}

func (p point) ToGeoJSON() geojson.GeoJSON {
	return PointToGeoJSON(p)
}

func PointFromS2Point(p s2.Point) Point {
	return point{p: p}
}

func PointFromLatLng(ll s2.LatLng) Point {
	return point{p: s2.PointFromLatLng(ll)}
}

func PointFromLatLngDegrees(lat float64, lng float64) Point {
	return point{p: s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng))}
}

type pointWithID struct {
	point
	id PointID
}

func (p pointWithID) FeatureID() FeatureID {
	return p.id.FeatureID()
}

func (p pointWithID) PointID() PointID {
	return p.id
}

func (p pointWithID) AllTags() []Tag {
	return []Tag{}
}

func (p pointWithID) Get(string) Tag {
	return Tag{}
}

func PointFromS2PointAndID(p s2.Point, id PointID) PointFeature {
	return pointWithID{point: point{p: p}, id: id}
}

type PointFeature interface {
	Feature
	Point
	PointID() PointID
}

type Path interface {
	Geometry
	Len() int
	Point(i int) s2.Point
	Polyline() *s2.Polyline
}

type path struct {
	ps []s2.Point
}

func (p path) Len() int {
	return len(p.ps)
}

func (p path) Point(i int) s2.Point {
	return p.ps[i]
}

func (p path) Polyline() *s2.Polyline {
	return (*s2.Polyline)(&p.ps)
}

func (p path) Covering(coverer s2.RegionCoverer) s2.CellUnion {
	return coverer.Covering((*s2.Polyline)(&p.ps))
}

func (p path) ToGeoJSON() geojson.GeoJSON {
	return PathToGeoJSON(p)
}

func PathFromS2Points(ps []s2.Point) Path {
	return path{ps: ps}
}

type PathFeature interface {
	Feature
	Path
	PathID() PathID
	Feature(i int) PointFeature
}

type Area interface {
	Geometry
	Len() int
	Polygon(i int) *s2.Polygon
}

type area struct {
	ps []*s2.Polygon
}

func (a area) Len() int {
	return len(a.ps)
}

func (a area) Polygon(i int) *s2.Polygon {
	return a.ps[i]
}

func (a area) Covering(coverer s2.RegionCoverer) s2.CellUnion {
	covering := make(s2.CellUnion, 0)
	for _, p := range a.ps {
		covering = s2.CellUnionFromUnion(covering, coverer.Covering(p))
	}
	return covering
}

func (a area) ToGeoJSON() geojson.GeoJSON {
	return AreaToGeoJSON(a)
}

func AreaFromS2Loop(l *s2.Loop) Area {
	return AreaFromS2Polygon(s2.PolygonFromLoops([]*s2.Loop{l}))
}

func AreaFromS2Polygon(p *s2.Polygon) Area {
	return area{ps: []*s2.Polygon{p}}
}

func AreaFromS2Polygons(ps []*s2.Polygon) Area {
	return area{ps: ps}
}

func AreaToS2Polygons(a Area) []*s2.Polygon {
	ps := make([]*s2.Polygon, a.Len())
	for i := 0; i < a.Len(); i++ {
		ps[i] = a.Polygon(i)
	}
	return ps
}

type AreaFeature interface {
	Feature
	Area
	AreaID() AreaID
	Feature(i int) []PathFeature
}

type RelationMember struct {
	ID   FeatureID
	Role string
}

type RelationFeature interface {
	Feature
	RelationID() RelationID
	Len() int
	Member(i int) RelationMember
	Covering(coverer s2.RegionCoverer) s2.CellUnion
}

type Features interface {
	Feature() Feature
	FeatureID() FeatureID
	Next() bool
}

type PathSegmentKey struct {
	ID    PathID
	First int
	Last  int
}

func (p *PathSegmentKey) ToPathSegment(path PathFeature) PathSegment {
	return PathSegment{path, p.First, p.Last}
}

type PathSegment struct {
	PathFeature // TODO: Rename to Feature, to match other structs?
	First       int
	Last        int
}

func (p PathSegment) Len() int {
	if p.First < p.Last {
		return p.Last - p.First + 1
	} else {
		return p.First - p.Last + 1
	}
}

func (p PathSegment) ToKey() PathSegmentKey {
	return PathSegmentKey{ID: p.PathFeature.PathID(), First: p.First, Last: p.Last}
}

func (p PathSegment) pathIndex(i int) int {
	if p.First < p.Last {
		if p.First+i <= p.Last {
			return p.First + i
		}
	} else if p.First-i >= p.Last {
		return p.First - i
	}
	panic(fmt.Sprintf("Segment point %d out of range (first: %d, last: %d)", i, p.First, p.Last))
}

func (p PathSegment) SegmentPoint(i int) s2.Point {
	return p.PathFeature.Point(p.pathIndex(i))
}

func (p PathSegment) SegmentFeature(i int) PointFeature {
	return p.PathFeature.Feature(p.pathIndex(i))
}

// TODO: Elsewhere this is called Feature, not Point - need to sort out the naming
// conventions
// TODO: Also consider adapting Segment to the Path interface, ie Point returns
// points relative to the segment (like SegmentPoint does now).
func (p PathSegment) FirstPoint() PointFeature {
	return p.PathFeature.Feature(p.First)
}

func (p PathSegment) LastPoint() PointFeature {
	return p.PathFeature.Feature(p.Last)
}

func ToPathSegment(path PathFeature) PathSegment {
	return PathSegment{path, 0, path.Len() - 1}
}

func (p PathSegment) Polyline() *s2.Polyline {
	polyline := *(p.PathFeature.Polyline())
	first, last := p.First, p.Last
	if first > last {
		first, last = last, first
	}
	segment := polyline[first : last+1]
	return &segment
}

var PathSegmentInvalid = PathSegment{PathFeature: nil}

type PathSegments interface {
	PathSegment() PathSegment
	Next() bool
}

type EmptyPathSegments struct{}

func (EmptyPathSegments) PathSegment() PathSegment {
	panic("No PathSegment")
}

func (EmptyPathSegments) Next() bool {
	return false
}

func AllPathSegments(p PathSegments) []PathSegment {
	segments := make([]PathSegment, 0, 8)
	if p != nil {
		for p.Next() {
			segments = append(segments, p.PathSegment())
		}
	}
	return segments
}

func FindPathSegmentByKey(key PathSegmentKey, w World) PathSegment {
	return PathSegment{
		PathFeature: FindPathByID(key.ID, w),
		First:       key.First,
		Last:        key.Last,
	}
}

type LocationsByID interface {
	FindLocationByID(id PointID) (s2.LatLng, bool)
}

type FeaturesByID interface {
	LocationsByID
	FindFeatureByID(id FeatureID) Feature
	HasFeatureWithID(id FeatureID) bool
}

type FeatureQuery interface {
	search.Query
	Matches(Feature) bool
}

type EachFeatureOptions struct {
	SkipPoints    bool
	SkipPaths     bool
	SkipAreas     bool
	SkipRelations bool
	Parallelism   int
}

type World interface {
	// TODO: Include transit once our use of it has stabalised
	FindFeatureByID(id FeatureID) Feature
	HasFeatureWithID(id FeatureID) bool
	FindLocationByID(id PointID) (s2.LatLng, bool)
	// TODO: make the query type more specific to Features, similar to the level in api.proto
	FindFeatures(query search.Query) Features
	FindRelationsByFeature(id FeatureID) RelationFeatures
	FindPathsByPoint(id PointID) PathSegments
	FindAreasByPoint(id PointID) AreaFeatures
	EachFeature(each func(f Feature, goroutine int) error, options *EachFeatureOptions) error

	// Returns a copy of all tokens known to this world's search index. The
	// order isn't defined.
	Tokens() []string
}

func AllFeatures(f Features) []Feature {
	features := make([]Feature, 0, 8)
	for f.Next() {
		features = append(features, f.Feature())
	}
	return features
}

func FindPointByID(id PointID, features FeaturesByID) PointFeature {
	if point := features.FindFeatureByID(id.FeatureID()); point != nil {
		return point.(PointFeature)
	}
	return nil
}

type pointFeatures struct {
	features Features
}

func (p pointFeatures) Next() bool {
	return p.features.Next()
}

func (p pointFeatures) Feature() PointFeature {
	if point, ok := p.features.Feature().(PointFeature); ok {
		return point
	}
	panic(fmt.Sprintf("Not a PointFeature: %T", p.features.Feature()))
}

func (p pointFeatures) FeatureID() FeatureID {
	return p.features.FeatureID()
}

type PointFeatures interface {
	Feature() PointFeature
	FeatureID() FeatureID
	Next() bool
}

func AllPoints(p PointFeatures) []PointFeature {
	features := make([]PointFeature, 0, 8)
	if p != nil {
		for p.Next() {
			features = append(features, p.Feature())
		}
	}
	return features
}

func NewPointFeatures(features Features) PointFeatures {
	return pointFeatures{features: features}
}

func FindPathByID(id PathID, features FeaturesByID) PathFeature {
	if path := features.FindFeatureByID(id.FeatureID()); path != nil {
		return path.(PathFeature)
	}
	return nil
}

type pathFeatures struct {
	features Features
}

func (p pathFeatures) Next() bool {
	return p.features.Next()
}

func (p pathFeatures) Feature() PathFeature {
	if path, ok := p.features.Feature().(PathFeature); ok {
		return path
	}
	panic(fmt.Sprintf("Not a PathFeature: %T", p.features.Feature()))
}

func (p pathFeatures) FeatureID() FeatureID {
	return p.features.FeatureID()
}

type PathFeatures interface {
	Feature() PathFeature
	FeatureID() FeatureID
	Next() bool
}

func AllPaths(p PathFeatures) []PathFeature {
	features := make([]PathFeature, 0, 8)
	if p != nil {
		for p.Next() {
			features = append(features, p.Feature())
		}
	}
	return features
}

func NewPathFeatures(features Features) PathFeatures {
	return pathFeatures{features: features}
}

func FindAreaByID(id AreaID, features FeaturesByID) AreaFeature {
	if area := features.FindFeatureByID(id.FeatureID()); area != nil {
		return area.(AreaFeature)
	}
	return nil
}

type areaFeatures struct {
	features Features
}

func (a areaFeatures) Next() bool {
	return a.features.Next()
}

func (a areaFeatures) Feature() AreaFeature {
	if area, ok := a.features.Feature().(AreaFeature); ok {
		return area
	}
	panic(fmt.Sprintf("Not an AreaFeature: %T", a.features.Feature()))
}

func (a areaFeatures) FeatureID() FeatureID {
	return a.features.FeatureID()
}

type AreaFeatures interface {
	Feature() AreaFeature
	FeatureID() FeatureID
	Next() bool
}

func AllAreas(a AreaFeatures) []AreaFeature {
	features := make([]AreaFeature, 0, 8)
	if a != nil {
		for a.Next() {
			features = append(features, a.Feature())
		}
	}
	return features
}

func NewAreaFeatures(features Features) AreaFeatures {
	return areaFeatures{features: features}
}

func FindRelationByID(id RelationID, features FeaturesByID) RelationFeature {
	if relation := features.FindFeatureByID(id.FeatureID()); relation != nil {
		return relation.(RelationFeature)
	}
	return nil
}

type relationFeatures struct {
	features Features
}

func (r relationFeatures) Next() bool {
	return r.features.Next()
}

type RelationFeatures interface {
	Feature() RelationFeature
	FeatureID() FeatureID
	RelationID() RelationID
	Next() bool
}

func (r relationFeatures) Feature() RelationFeature {
	if relation, ok := r.features.Feature().(RelationFeature); ok {
		return relation
	}
	panic(fmt.Sprintf("Not an RelationFeature: %T", r.features.Feature()))
}

func (r relationFeatures) FeatureID() FeatureID {
	return r.features.FeatureID()
}

func (r relationFeatures) RelationID() RelationID {
	return r.FeatureID().ToRelationID()
}

func AllRelations(r RelationFeatures) []RelationFeature {
	features := make([]RelationFeature, 0, 8)
	if r != nil {
		for r.Next() {
			features = append(features, r.Feature())
		}
	}
	return features
}

func NewRelationFeatures(features Features) RelationFeatures {
	return relationFeatures{features: features}
}

func FindPoints(q search.Query, w World) PointFeatures {
	q = FeatureTypeQuery{Type: FeatureTypePoint, Query: q}
	return NewPointFeatures(w.FindFeatures(q))
}

func FindPaths(q search.Query, w World) PathFeatures {
	q = FeatureTypeQuery{Type: FeatureTypePath, Query: q}
	return NewPathFeatures(w.FindFeatures(q))
}

func FindAreas(q search.Query, w World) AreaFeatures {
	q = FeatureTypeQuery{Type: FeatureTypeArea, Query: q}
	return NewAreaFeatures(w.FindFeatures(q))
}

func FindRelations(q search.Query, w World) RelationFeatures {
	q = FeatureTypeQuery{Type: FeatureTypeRelation, Query: q}
	return NewRelationFeatures(w.FindFeatures(q))
}

type FeatureValues interface {
	Feature(v search.Value) Feature
	ID(v search.Value) FeatureID
}

type FeatureIndex interface {
	search.Index
	FeatureValues
}

type FeatureIterator struct {
	i     search.Iterator
	index FeatureIndex
}

func NewFeatureIterator(i search.Iterator, index FeatureIndex) *FeatureIterator {
	return &FeatureIterator{i: i, index: index}
}

func (f *FeatureIterator) Next() bool {
	return f.i.Next()
}

func (f *FeatureIterator) Feature() Feature {
	return f.index.Feature(f.i.Value())
}

func (f *FeatureIterator) FeatureID() FeatureID {
	return f.index.ID(f.i.Value())
}

func fillPropertiesFromTags(tagged Tagged, feature *geojson.Feature) {
	for _, tag := range tagged.AllTags() {
		feature.Properties[tag.Key] = tag.Value
	}
}

func PointToGeoJSON(point Point) *geojson.Feature {
	return geojson.NewFeatureFromS2LatLng(s2.LatLngFromPoint(point.Point()))
}

func PointFeatureToGeoJSON(point PointFeature) *geojson.Feature {
	g := PointToGeoJSON(point)
	fillPropertiesFromTags(point, g)
	return g
}

func PathToGeoJSON(path Path) *geojson.Feature {
	polyline := path.Polyline()
	geometry := geojson.GeometryFromLineString(geojson.FromPolyline(polyline))
	return geojson.NewFeatureWithGeometry(geometry)
}

func PathFeatureToGeoJSON(path PathFeature) *geojson.Feature {
	g := PathToGeoJSON(path)
	fillPropertiesFromTags(path, g)
	return g
}

func AreaToGeoJSON(area Area) *geojson.Feature {
	coordinates := make([][][]geojson.Coordinate, area.Len())
	for i := 0; i < area.Len(); i++ {
		polygon := area.Polygon(i)
		coordinates[i] = geojson.FromPolygon(polygon)
	}
	var geometry geojson.Geometry
	if len(coordinates) == 1 {
		geometry = geojson.GeometryFromPolygon(coordinates[0])
	} else {
		geometry = geojson.GeometryFromMultiPolygon(coordinates)
	}
	return geojson.NewFeatureWithGeometry(geometry)
}

func AreaFeatureToGeoJSON(area AreaFeature) *geojson.Feature {
	g := AreaToGeoJSON(area)
	fillPropertiesFromTags(area, g)
	return g
}

func RelationFeatureToGeoJSON(relation RelationFeature, byID FeaturesByID) *geojson.FeatureCollection {
	collection := geojson.NewFeatureCollection()
	for i := 0; i < relation.Len(); i++ {
		if f := byID.FindFeatureByID(relation.Member(i).ID); f != nil {
			collection.Add(f.ToGeoJSON())
		}
	}
	return collection
}

func AngleToMeters(angle s1.Angle) float64 {
	return units.AngleToMeters(angle)
}

func MetersToAngle(meters float64) s1.Angle {
	return units.MetersToAngle(meters)
}

func AreaToMeters2(area float64) float64 {
	return units.AreaToMeters2(area)
}

func Meters2ToArea(m2 float64) float64 {
	return units.Meters2ToArea(m2)
}
