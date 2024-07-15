package b6

import (
	"fmt"
	"reflect"
	"strings"

	"diagonal.works/b6/geometry"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/search"
	"github.com/golang/geo/s2"
)

// Spatial search functions are covered with tests in ingest/spatial_test.go,
// for ease of loading test data.

type MightIntersect struct {
	s2.Region
}

func (m MightIntersect) Compile(i FeatureIndex, w World) search.Iterator {
	return search.NewSpatialFromRegion(s2.Region(m)).Compile(i)
}

func (m MightIntersect) Matches(f Feature, w World) bool {
	return true
}

func (m MightIntersect) String() string {
	return search.NewSpatialFromRegion(s2.Region(m)).String()
}

func (m MightIntersect) ToProto() (*pb.QueryProto, error) {
	coverer := search.MakeCoverer()
	covering := coverer.Covering(m.Region)
	ids := make([]uint64, len(covering))
	for ii, id := range covering {
		ids[ii] = uint64(id)
	}
	return &pb.QueryProto{
		Query: &pb.QueryProto_MightIntersect{
			MightIntersect: &pb.S2CellIDsProto{
				S2CellIDs: ids,
			},
		},
	}, nil
}

func (m MightIntersect) Equal(other Query) bool {
	if mm, ok := other.(MightIntersect); ok {
		return reflect.DeepEqual(m.Region, mm.Region)
	}
	if mm, ok := other.(*MightIntersect); ok {
		return reflect.DeepEqual(m.Region, mm.Region)
	}
	return false
}

type IntersectsCells struct {
	Cells []s2.Cell
}

func (i IntersectsCells) String() string {
	tokens := make([]string, len(i.Cells))
	for ii, c := range i.Cells {
		tokens[ii] = c.ID().ToToken()
	}
	return fmt.Sprintf("(intersects-cells %s)", strings.Join(tokens, ","))
}

func (i IntersectsCells) Compile(index FeatureIndex, w World) search.Iterator {
	union := make(s2.CellUnion, len(i.Cells))
	for ii, cell := range i.Cells {
		union[ii] = cell.ID()
	}
	union.Normalize()
	if index, ok := index.(FeatureIndex); ok {
		return &intersectsCells{cells: i.Cells, index: index, iterator: search.NewSpatialFromRegion(&union).Compile(index)}
	}
	return search.NewEmptyIterator()
}

func (i IntersectsCells) Matches(feature Feature, w World) bool {
	return cellsIntersectFeature(i.Cells, feature)
}

func (i IntersectsCells) ToProto() (*pb.QueryProto, error) {
	ids := make([]uint64, len(i.Cells))
	for ii, cell := range i.Cells {
		ids[ii] = uint64(cell.ID())
	}
	return &pb.QueryProto{
		Query: &pb.QueryProto_IntersectsCells{
			IntersectsCells: &pb.S2CellIDsProto{
				S2CellIDs: ids,
			},
		},
	}, nil
}

func (i IntersectsCells) Equal(other Query) bool {
	var ii IntersectsCells
	switch iii := other.(type) {
	case IntersectsCells:
		ii = iii
	case *IntersectsCells:
		ii = *iii
	}
	if ii.Cells != nil {
		if len(i.Cells) != len(ii.Cells) {
			return false
		}
		for j := range i.Cells {
			if i.Cells[j].ID() != ii.Cells[j].ID() {
				return false
			}
		}
		return true
	}
	return false
}

func cellsIntersectFeature(cells []s2.Cell, feature Feature) bool {
	if f, ok := feature.(Geometry); ok {
		switch f.GeometryType() {
		case GeometryTypePoint:
			for _, c := range cells {
				if c.ContainsPoint(f.Point()) {
					return true
				}
			}
		case GeometryTypePath:
			polyline := f.Polyline()
			for _, c := range cells {
				if polyline.IntersectsCell(c) {
					return true
				}
			}
		case GeometryTypeArea:
			for j := 0; j < feature.(AreaFeature).Len(); j++ {
				polygon := feature.(AreaFeature).Polygon(j)
				for _, c := range cells {
					if polygon.IntersectsCell(c) {
						return true
					}
				}
			}
		}
	}
	return false
}

type intersectsCells struct {
	cells    []s2.Cell
	index    FeatureIndex
	iterator search.Iterator
}

func (i *intersectsCells) Next() bool {
	for {
		ok := i.iterator.Next()
		if !ok {
			return false
		}
		if cellsIntersectFeature(i.cells, i.index.Feature(i.Value())) {
			return true
		}
	}
}

func (i *intersectsCells) Advance(key search.Key) bool {
	ok := i.iterator.Advance(key)
	for ok && !cellsIntersectFeature(i.cells, i.index.Feature(i.Value())) {
		ok = i.iterator.Next()
	}
	return ok
}

func (i *intersectsCells) Value() search.Value {
	return i.iterator.Value()
}

func (i *intersectsCells) EstimateLength() int {
	return i.iterator.EstimateLength()
}

func NewIntersectsCellUnion(union s2.CellUnion) Query {
	cells := make([]s2.Cell, len(union))
	for i, cell := range union {
		cells[i] = s2.CellFromCellID(cell)
	}
	return IntersectsCells{Cells: cells}
}

func NewIntersectsCell(cell s2.Cell) Query {
	return IntersectsCells{Cells: []s2.Cell{cell}}
}

func NewIntersectsCellID(cell s2.CellID) Query {
	return NewIntersectsCell(s2.CellFromCellID(cell))
}

type IntersectsCap struct {
	cap      s2.Cap
	interior s2.CellUnion
	exterior s2.CellUnion
}

func NewIntersectsCap(cap s2.Cap) *IntersectsCap {
	coverer := s2.RegionCoverer{MaxLevel: 22, MaxCells: 4}
	return &IntersectsCap{
		cap:      cap,
		interior: coverer.InteriorCovering(cap),
		exterior: coverer.Covering(cap),
	}
}

func (i *IntersectsCap) String() string {
	ll := s2.LatLngFromPoint(i.cap.Center())
	return fmt.Sprintf("(intersecting-cap %f,%f %.2f)", ll.Lat.Degrees(), ll.Lng.Degrees(), AngleToMeters(i.cap.Radius()))
}

func (i *IntersectsCap) ToProto() (*pb.QueryProto, error) {
	return &pb.QueryProto{
		Query: &pb.QueryProto_IntersectsCap{
			IntersectsCap: &pb.CapProto{
				Center:       NewPointProtoFromS2Point(i.cap.Center()),
				RadiusMeters: AngleToMeters(i.cap.Radius()),
			},
		},
	}, nil
}

func (i *IntersectsCap) Compile(index FeatureIndex, w World) search.Iterator {
	if index, ok := index.(FeatureIndex); ok {
		return &intersectsCap{cap: i, index: index, iterator: search.NewSpatialFromRegion(i.cap).Compile(index)}
	}
	return search.NewEmptyIterator()
}

func (i *IntersectsCap) Matches(feature Feature, w World) bool {
	if f, ok := feature.(Geometry); ok {
		switch f.GeometryType() {
		case GeometryTypePoint:
			if i.cap.ContainsPoint(f.Point()) {
				return true
			}
		case GeometryTypePath:
			projection, _ := f.Polyline().Project(i.cap.Center())
			if i.cap.ContainsPoint(projection) {
				return true
			}
		case GeometryTypeArea:
			for j := 0; j < feature.(AreaFeature).Len(); j++ {
				if i.IntersectsPolygon(feature.(AreaFeature).Polygon(j)) {
					return true
				}
			}
		}
	}
	return false
}

type intersectsCapYAML struct {
	Center PointExpression
	Radius float64
}

func (i IntersectsCap) MarshalYAML() (interface{}, error) {
	return &intersectsCapYAML{
		Center: PointExpression(s2.LatLngFromPoint(i.cap.Center())),
		Radius: AngleToMeters(i.cap.Radius()),
	}, nil
}

func (i *IntersectsCap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var y intersectsCapYAML
	err := unmarshal(&y)
	if err == nil {
		cap := s2.CapFromCenterAngle(s2.PointFromLatLng(s2.LatLng(y.Center)), MetersToAngle(y.Radius))
		*i = *NewIntersectsCap(cap)
	}
	return err
}

func (i IntersectsCap) Equal(other Query) bool {
	if ii, ok := other.(*IntersectsCap); ok {
		return i.cap.Center() == ii.cap.Center() && i.cap.Radius() == ii.cap.Radius()
	}
	return false
}

// If the polygon has less than this number of vetices, it's faster to
// skip the index and test the edges of the polygon directly. Derived
// empirically via benchmarks alongside the unit tests.
const indexUseFasterAboveVetexCount = 16

func (i *IntersectsCap) IntersectsPolygon(p *s2.Polygon) bool {
	if p.Loop(0).NumVertices() > indexUseFasterAboveVetexCount {
		for _, id := range i.interior {
			if p.IntersectsCell(s2.CellFromCellID(id)) {
				return true
			}
		}
		outside := true
		for _, id := range i.exterior {
			if p.IntersectsCell(s2.CellFromCellID(id)) {
				outside = false
				break
			}
		}
		if outside {
			return false
		}
	}
	return CapIntersectsPolygon(i.cap, p)
}

type intersectsCap struct {
	cap      *IntersectsCap
	index    FeatureIndex
	iterator search.Iterator
}

func (i *intersectsCap) Next() bool {
	for {
		ok := i.iterator.Next()
		if !ok {
			return false
		}
		if i.cap.Matches(i.index.Feature(i.Value()), nil) {
			return true
		}
	}
}

func (i *intersectsCap) Advance(key search.Key) bool {
	ok := i.iterator.Advance(key)
	for ok && !i.cap.Matches(i.index.Feature(i.Value()), nil) {
		ok = i.iterator.Next()
	}
	return ok
}

func (i *intersectsCap) Value() search.Value {
	return i.iterator.Value()
}

func (i *intersectsCap) EstimateLength() int {
	return i.iterator.EstimateLength()
}

func CapIntersectsPolygon(c s2.Cap, p *s2.Polygon) bool {
	inside := 0
	for i := 0; i < p.NumLoops(); i++ {
		loop := p.Loop(i)
		onLeft := true
		for j := 0; j < loop.NumEdges(); j++ {
			edge := loop.Edge(j)
			point := s2.Project(c.Center(), edge.V0, edge.V1)
			if c.ContainsPoint(point) {
				return true
			}
			onLeft = onLeft && s2.Sign(c.Center(), edge.V0, edge.V1)
		}
		if onLeft {
			inside++
		}
	}
	return inside%2 == 1
}

type IntersectsFeature struct {
	ID FeatureID
}

func (i IntersectsFeature) Matches(f Feature, w World) bool {
	return i.ID == f.FeatureID() || i.toGeometryQuery(w).Matches(f, w)
}

func (i IntersectsFeature) Compile(index FeatureIndex, w World) search.Iterator {
	return i.toGeometryQuery(w).Compile(index, w)
}

func (i IntersectsFeature) String() string {
	return fmt.Sprintf("(intersects-feature %s)", i.ID)
}

func (i IntersectsFeature) ToProto() (*pb.QueryProto, error) {
	return &pb.QueryProto{
		Query: &pb.QueryProto_IntersectsFeature{
			IntersectsFeature: NewProtoFromFeatureID(i.ID),
		},
	}, nil
}

func (i IntersectsFeature) toGeometryQuery(w World) Query {
	if f := w.FindFeatureByID(i.ID); f != nil {
		if f, ok := f.(Geometry); ok {
			switch f.GeometryType() {
			case GeometryTypePoint:
				return IntersectsPoint{f.Point()}
			case GeometryTypePath:
				return IntersectsPolyline{f.Polyline()}
			case GeometryTypeArea:
				return IntersectsMultiPolygon{MultiPolygon: f.(AreaFeature).MultiPolygon()}
			}
		}
	}
	return Empty{}
}

func (i IntersectsFeature) Equal(other Query) bool {
	if ii, ok := other.(IntersectsFeature); ok {
		return i.ID == ii.ID
	}
	return false
}

type IntersectsPoint struct {
	Point s2.Point
}

func (i IntersectsPoint) String() string {
	ll := s2.LatLngFromPoint(i.Point)
	return fmt.Sprintf("(intersects-point %f %f)", ll.Lat.Degrees(), ll.Lng.Degrees())
}

func (i IntersectsPoint) Compile(index FeatureIndex, w World) search.Iterator {
	if index, ok := index.(FeatureIndex); ok {
		return &intersectsPoint{point: i.Point, index: index, iterator: search.NewSpatialFromRegion(i.Point).Compile(index)}
	}
	return search.NewEmptyIterator()
}

func (i IntersectsPoint) Matches(f Feature, w World) bool {
	return pointIntersectsFeature(i.Point, f)
}

func (i IntersectsPoint) ToProto() (*pb.QueryProto, error) {
	return &pb.QueryProto{
		Query: &pb.QueryProto_IntersectsPoint{
			IntersectsPoint: NewPointProtoFromS2Point(i.Point),
		},
	}, nil
}

func (i IntersectsPoint) Equal(other Query) bool {
	if ii, ok := other.(*IntersectsPoint); ok {
		return i.Point == ii.Point
	}
	return false
}

type intersectsPoint struct {
	point    s2.Point
	index    FeatureIndex
	iterator search.Iterator
}

func (i *intersectsPoint) Next() bool {
	for {
		ok := i.iterator.Next()
		if !ok {
			return false
		}
		if pointIntersectsFeature(i.point, i.index.Feature(i.Value())) {
			return true
		}
	}
}

func pointIntersectsFeature(point s2.Point, feature Feature) bool {
	if f, ok := feature.(Geometry); ok {
		switch f.GeometryType() {
		case GeometryTypePoint:
			return f.Point() == point
		case GeometryTypePath:
			projection, _ := f.Polyline().Project(point)
			// TODO: Define the tolerance with more rigour
			return projection.Distance(point) < MetersToAngle(0.001)
		case GeometryTypeArea:
			for i := 0; i < feature.(AreaFeature).Len(); i++ {
				if feature.(AreaFeature).Polygon(i).ContainsPoint(point) {
					return true
				}
			}
		}
	}
	return false
}

func (i *intersectsPoint) Advance(key search.Key) bool {
	ok := i.iterator.Advance(key)
	for ok && !pointIntersectsFeature(i.point, i.index.Feature(i.Value())) {
		ok = i.iterator.Next()
	}
	return ok
}

func (i *intersectsPoint) Value() search.Value {
	return i.iterator.Value()
}

func (i *intersectsPoint) EstimateLength() int {
	return i.iterator.EstimateLength()
}

type IntersectsPolyline struct {
	Polyline *s2.Polyline
}

func (i IntersectsPolyline) String() string {
	return "(intersects-polyline)"
}

func (i IntersectsPolyline) Compile(index FeatureIndex, w World) search.Iterator {
	if index, ok := index.(FeatureIndex); ok {
		return &intersectsPolyline{polyline: i.Polyline, index: index, iterator: search.NewSpatialFromRegion(i.Polyline).Compile(index)}
	}
	return search.NewEmptyIterator()
}

func (i IntersectsPolyline) Matches(f Feature, w World) bool {
	return polylineIntersectsFeature(i.Polyline, f)
}

func (i IntersectsPolyline) ToProto() (*pb.QueryProto, error) {
	return &pb.QueryProto{
		Query: &pb.QueryProto_IntersectsPolyline{
			IntersectsPolyline: NewPolylineProto(i.Polyline),
		},
	}, nil
}

func (i IntersectsPolyline) Equal(other Query) bool {
	if ii, ok := other.(*IntersectsPolyline); ok {
		return geometry.PolylineEqual(i.Polyline, ii.Polyline)
	}
	return false
}

type intersectsPolyline struct {
	polyline *s2.Polyline
	index    FeatureIndex
	iterator search.Iterator
}

func (i *intersectsPolyline) Next() bool {
	for {
		ok := i.iterator.Next()
		if !ok {
			return false
		}
		if polylineIntersectsFeature(i.polyline, i.index.Feature(i.Value())) {
			return true
		}
	}
}

func polylineIntersectsPolygon(polyline *s2.Polyline, polygon *s2.Polygon) bool {
	// TODO: This implementation is incorrect, as a path can pass though a polygon
	// without a vertex being inside it. Implement this properly based on, eg, the
	// alogorithms in geo/clip.go
	for i := 0; i < len(*polyline); i++ {
		if polygon.ContainsPoint((*polyline)[i]) {
			return true
		}
	}
	return false
}

func polylineIntersectsFeature(polyline *s2.Polyline, feature Feature) bool {
	if f, ok := feature.(Geometry); ok {
		switch f.GeometryType() {
		case GeometryTypePoint:
			projection, _ := polyline.Project(f.Point())
			// TODO: Define the tolerance with more rigour
			return projection.Distance(f.Point()) < MetersToAngle(0.001)
		case GeometryTypePath:
			return f.Polyline().Intersects(polyline)
		case GeometryTypeArea:
			for i := 0; i < feature.(AreaFeature).Len(); i++ {
				if polylineIntersectsPolygon(polyline, feature.(AreaFeature).Polygon(i)) {
					return true
				}
			}
		}
	}
	return false
}

func (i *intersectsPolyline) Advance(key search.Key) bool {
	ok := i.iterator.Advance(key)
	for ok && !polylineIntersectsFeature(i.polyline, i.index.Feature(i.Value())) {
		ok = i.iterator.Next()
	}
	return ok
}

func (i *intersectsPolyline) Value() search.Value {
	return i.iterator.Value()
}

func (i *intersectsPolyline) EstimateLength() int {
	return i.iterator.EstimateLength()
}

type IntersectsMultiPolygon struct {
	MultiPolygon geometry.MultiPolygon
}

func (i IntersectsMultiPolygon) String() string {
	return "(intersects-multipolygon)"
}

func (i IntersectsMultiPolygon) Compile(index FeatureIndex, w World) search.Iterator {
	if index, ok := index.(FeatureIndex); ok {
		coverer := search.MakeCoverer()
		var covering s2.CellUnion
		for _, polygon := range i.MultiPolygon {
			covering = s2.CellUnionFromUnion(covering, coverer.Covering(polygon))
		}
		return &intersectsMultiPolygon{polygons: i.MultiPolygon, index: index, iterator: search.NewSpatialFromCellUnion(covering).Compile(index)}
	}
	return search.NewEmptyIterator()
}

func (i IntersectsMultiPolygon) Matches(f Feature, w World) bool {
	return multiPolygonIntersectsFeature(i.MultiPolygon, f)
}

func (i IntersectsMultiPolygon) ToProto() (*pb.QueryProto, error) {
	return &pb.QueryProto{
		Query: &pb.QueryProto_IntersectsMultiPolygon{
			IntersectsMultiPolygon: NewMultiPolygonProto(i.MultiPolygon),
		},
	}, nil
}

func (i IntersectsMultiPolygon) Equal(other Query) bool {
	if ii, ok := other.(*IntersectsMultiPolygon); ok {
		return geometry.MultiPolygonEqual(i.MultiPolygon, ii.MultiPolygon)
	}
	return false
}

type intersectsMultiPolygon struct {
	polygons geometry.MultiPolygon
	index    FeatureIndex
	iterator search.Iterator
}

func (i *intersectsMultiPolygon) Next() bool {
	for {
		ok := i.iterator.Next()
		if !ok {
			return false
		}
		if multiPolygonIntersectsFeature(i.polygons, i.index.Feature(i.Value())) {
			return true
		}
	}
}

func multiPolygonIntersectsFeature(polygons geometry.MultiPolygon, feature Feature) bool {
	if f, ok := feature.(Geometry); ok {
		switch f.GeometryType() {
		case GeometryTypePoint:
			for _, polygon := range polygons {
				return polygon.ContainsPoint(f.Point())
			}
		case GeometryTypePath:
			polyline := f.Polyline()
			for _, polygon := range polygons {
				if polylineIntersectsPolygon(polyline, polygon) {
					return true
				}
			}
		case GeometryTypeArea:
			ps := make([]*s2.Polygon, feature.(AreaFeature).Len())
			for i := 0; i < feature.(AreaFeature).Len(); i++ {
				ps[i] = feature.(AreaFeature).Polygon(i)
			}
			for _, a := range ps {
				for _, b := range polygons {
					if a.Intersects(b) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (i *intersectsMultiPolygon) Advance(key search.Key) bool {
	ok := i.iterator.Advance(key)
	for ok && !multiPolygonIntersectsFeature(i.polygons, i.index.Feature(i.Value())) {
		ok = i.iterator.Next()
	}
	return ok
}

func (i *intersectsMultiPolygon) Value() search.Value {
	return i.iterator.Value()
}

func (i *intersectsMultiPolygon) EstimateLength() int {
	return i.iterator.EstimateLength()
}

type searchFeatureIterator struct {
	i     search.Iterator
	index FeatureIndex
}

func NewSearchFeatureIterator(i search.Iterator, index FeatureIndex) *searchFeatureIterator {
	return &searchFeatureIterator{i: i, index: index}
}

func (f *searchFeatureIterator) Next() bool {
	return f.i.Next()
}

func (f *searchFeatureIterator) Feature() Feature {
	return f.index.Feature(f.i.Value())
}

func (f *searchFeatureIterator) FeatureID() FeatureID {
	return f.index.ID(f.i.Value())
}

func NewQueryFromProto(p *pb.QueryProto) (Query, error) {
	switch q := p.Query.(type) {
	case *pb.QueryProto_All:
		return All{}, nil
	case *pb.QueryProto_Keyed:
		return Keyed{q.Keyed}, nil
	case *pb.QueryProto_Tagged:
		return Tagged{Key: q.Tagged.Key, Value: NewStringExpression(q.Tagged.Value)}, nil
	case *pb.QueryProto_IntersectsCap:
		ll := PointProtoToS2LatLng(q.IntersectsCap.Center)
		cap := s2.CapFromCenterAngle(s2.PointFromLatLng(ll), MetersToAngle(q.IntersectsCap.RadiusMeters))
		return NewIntersectsCap(cap), nil
	case *pb.QueryProto_IntersectsFeature:
		return IntersectsFeature{ID: NewFeatureIDFromProto(q.IntersectsFeature)}, nil
	case *pb.QueryProto_IntersectsPoint:
		return IntersectsPoint{PointProtoToS2Point(q.IntersectsPoint)}, nil
	case *pb.QueryProto_IntersectsPolyline:
		return IntersectsPolyline{PolylineProtoToS2Polyline(q.IntersectsPolyline)}, nil
	case *pb.QueryProto_IntersectsMultiPolygon:
		return IntersectsMultiPolygon{MultiPolygon: MultiPolygonProtoToS2MultiPolygon(q.IntersectsMultiPolygon)}, nil
	case *pb.QueryProto_Typed:
		if q.Typed.Query != nil {
			child, err := NewQueryFromProto(q.Typed.Query)
			if err != nil {
				return Empty{}, err
			}
			return Typed{Type: NewFeatureTypeFromProto(q.Typed.Type), Query: child}, nil
		}
	case *pb.QueryProto_Intersection:
		intersection := make(Intersection, len(q.Intersection.Queries))
		for i, next := range q.Intersection.Queries {
			if child, err := NewQueryFromProto(next); err == nil {
				intersection[i] = child
			} else {
				return Empty{}, err
			}
		}
		return intersection, nil
	case *pb.QueryProto_Union:
		union := make(Union, len(q.Union.Queries))
		for i, next := range q.Union.Queries {
			if child, err := NewQueryFromProto(next); err == nil {
				union[i] = child
			} else {
				return Empty{}, err
			}
		}
		return union, nil
	}
	return Empty{}, fmt.Errorf("Can't handle query %v", p)
}
