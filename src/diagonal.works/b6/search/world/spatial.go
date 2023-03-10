package world

import (
	"fmt"

	"diagonal.works/b6"
	"diagonal.works/b6/geometry"
	"diagonal.works/b6/search"

	"github.com/golang/geo/s2"
)

type intersectsCapQuery struct {
	cap s2.Cap
}

func (i intersectsCapQuery) String() string {
	ll := s2.LatLngFromPoint(i.cap.Center())
	return fmt.Sprintf("(intersects-cap %f %f %.2f)", ll.Lat.Degrees(), ll.Lng.Degrees(), b6.AngleToMeters(i.cap.Radius()))
}

func (i intersectsCapQuery) Compile(index search.Index) search.Iterator {
	if index, ok := index.(b6.FeatureIndex); ok {
		// TODO: Implement a more exact way of calculating polygon/cap intersection
		capLoop := s2.RegularLoop(i.cap.Center(), i.cap.Radius(), 1024)
		capPolygon := s2.PolygonFromLoops([]*s2.Loop{capLoop})
		return &intersectsCap{cap: i.cap, capPolygon: capPolygon, index: index, iterator: search.NewSpatialFromRegion(i.cap).Compile(index)}
	}
	return search.NewEmptyIterator()
}

func (i intersectsCapQuery) Matches(feature b6.Feature) bool {
	polygon := s2.PolygonFromLoops([]*s2.Loop{s2.RegularLoop(i.cap.Center(), i.cap.Radius(), 1024)})
	return capIntersectsFeature(i.cap, polygon, feature)
}

func NewIntersectsCap(cap s2.Cap) b6.FeatureQuery {
	return &intersectsCapQuery{cap: cap}
}

type intersectsCap struct {
	cap        s2.Cap
	capPolygon *s2.Polygon
	index      b6.FeatureIndex
	iterator   search.Iterator
}

func (i *intersectsCap) Next() bool {
	for {
		ok := i.iterator.Next()
		if !ok {
			return false
		}
		if capIntersectsFeature(i.cap, i.capPolygon, i.index.Feature(i.Value())) {
			return true
		}
	}
}

func capIntersectsFeature(cap s2.Cap, capPolygon *s2.Polygon, feature b6.Feature) bool {
	switch f := feature.(type) {
	case b6.PointFeature:
		if cap.ContainsPoint(f.Point()) {
			return true
		}
	case b6.PathFeature:
		projection, _ := f.Polyline().Project(cap.Center())
		if cap.ContainsPoint(projection) {
			return true
		}
	case b6.AreaFeature:
		for j := 0; j < f.Len(); j++ {
			if f.Polygon(j).Intersects(capPolygon) {
				return true
			}
		}
	}
	return false
}

func (i *intersectsCap) Advance(key search.Key) bool {
	ok := i.iterator.Advance(key)
	for ok && !capIntersectsFeature(i.cap, i.capPolygon, i.index.Feature(i.Value())) {
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

func (i *intersectsCap) ToQuery() search.Query {
	return &intersectsCapQuery{cap: i.cap}
}

func NewIntersectsFeature(f b6.PhysicalFeature) b6.FeatureQuery {
	switch f := f.(type) {
	case b6.Point:
		return intersectsPointQuery{point: f.Point()}
	case b6.Path:
		return intersectsPolylineQuery{polyline: f.Polyline()}
	case b6.Area:
		return NewIntersectsArea(f)
	}
	return Empty{}
}

func NewIntersectsArea(a b6.Area) b6.FeatureQuery {
	polygons := make(geometry.MultiPolygon, a.Len())
	for i := 0; i < a.Len(); i++ {
		polygons[i] = a.Polygon(i)
	}
	return intersectsMultiPolygonQuery{polygons: polygons}
}

type intersectsPointQuery struct {
	point s2.Point
}

func (i intersectsPointQuery) String() string {
	ll := s2.LatLngFromPoint(i.point)
	return fmt.Sprintf("(intersects-point %f %f)", ll.Lat.Degrees(), ll.Lng.Degrees())
}

func (i intersectsPointQuery) Compile(index search.Index) search.Iterator {
	if index, ok := index.(b6.FeatureIndex); ok {
		return &intersectsPoint{point: i.point, index: index, iterator: search.NewSpatialFromRegion(i.point).Compile(index)}
	}
	return search.NewEmptyIterator()
}

func (i intersectsPointQuery) Matches(f b6.Feature) bool {
	return pointIntersectsFeature(i.point, f)
}

type intersectsPoint struct {
	point    s2.Point
	index    b6.FeatureIndex
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

func pointIntersectsFeature(point s2.Point, feature b6.Feature) bool {
	switch f := feature.(type) {
	case b6.PointFeature:
		return f.Point() == point
	case b6.PathFeature:
		projection, _ := f.Polyline().Project(point)
		// TODO: Define the tolerance with more rigour
		return projection.Distance(point) < b6.MetersToAngle(0.001)
	case b6.AreaFeature:
		for i := 0; i < f.Len(); i++ {
			if f.Polygon(i).ContainsPoint(point) {
				return true
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

func (i *intersectsPoint) ToQuery() search.Query {
	return &intersectsPointQuery{point: i.point}
}

func NewIntersectsPoint(p s2.Point) b6.FeatureQuery {
	return &intersectsPointQuery{point: p}
}

type intersectsPolylineQuery struct {
	polyline *s2.Polyline
}

func (i intersectsPolylineQuery) String() string {
	return "(intersects-polyline)"
}

func (i intersectsPolylineQuery) Compile(index search.Index) search.Iterator {
	if index, ok := index.(b6.FeatureIndex); ok {
		return &intersectsPolyline{polyline: i.polyline, index: index, iterator: search.NewSpatialFromRegion(i.polyline).Compile(index)}
	}
	return search.NewEmptyIterator()
}

func (i intersectsPolylineQuery) Matches(f b6.Feature) bool {
	return polylineIntersectsFeature(i.polyline, f)
}

type intersectsPolyline struct {
	polyline *s2.Polyline
	index    b6.FeatureIndex
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

func polylineIntersectsFeature(polyline *s2.Polyline, feature b6.Feature) bool {
	switch f := feature.(type) {
	case b6.PointFeature:
		projection, _ := polyline.Project(f.Point())
		// TODO: Define the tolerance with more rigour
		return projection.Distance(f.Point()) < b6.MetersToAngle(0.001)
	case b6.PathFeature:
		return f.Polyline().Intersects(polyline)
	case b6.AreaFeature:
		for i := 0; i < f.Len(); i++ {
			if polylineIntersectsPolygon(polyline, f.Polygon(i)) {
				return true
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

func (i *intersectsPolyline) ToQuery() search.Query {
	return &intersectsPolylineQuery{polyline: i.polyline}
}

type intersectsMultiPolygonQuery struct {
	polygons geometry.MultiPolygon
}

func (i intersectsMultiPolygonQuery) String() string {
	return "(intersects-multipolygon)"
}

func (i intersectsMultiPolygonQuery) Compile(index search.Index) search.Iterator {
	if index, ok := index.(b6.FeatureIndex); ok {
		coverer := search.MakeCoverer()
		var covering s2.CellUnion
		for _, polygon := range i.polygons {
			covering = s2.CellUnionFromUnion(covering, coverer.Covering(polygon))
		}
		return &intersectsMultiPolygon{polygons: i.polygons, index: index, iterator: search.NewSpatialFromCellUnion(covering).Compile(index)}
	}
	return search.NewEmptyIterator()
}

func (i intersectsMultiPolygonQuery) Matches(f b6.Feature) bool {
	return multiPolygonIntersectsFeature(i.polygons, f)
}

type intersectsMultiPolygon struct {
	polygons geometry.MultiPolygon
	index    b6.FeatureIndex
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

func multiPolygonIntersectsFeature(polygons geometry.MultiPolygon, feature b6.Feature) bool {
	switch f := feature.(type) {
	case b6.PointFeature:
		for _, polygon := range polygons {
			if polygon.ContainsPoint(f.Point()) {
				return true
			}
		}
	case b6.PathFeature:
		polyline := f.Polyline()
		for _, polygon := range polygons {
			if polylineIntersectsPolygon(polyline, polygon) {
				return true
			}
		}
	case b6.AreaFeature:
		ps := make([]*s2.Polygon, f.Len())
		for i := 0; i < f.Len(); i++ {
			ps[i] = f.Polygon(i)
		}
		for _, a := range ps {
			for _, b := range polygons {
				if a.Intersects(b) {
					return true
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

func (i *intersectsMultiPolygon) ToQuery() search.Query {
	return &intersectsMultiPolygonQuery{polygons: i.polygons}
}

func NewIntersectsPolygon(p *s2.Polygon) b6.FeatureQuery {
	return &intersectsMultiPolygonQuery{polygons: []*s2.Polygon{p}}
}

func NewIntersectsMultiPolygon(p geometry.MultiPolygon) b6.FeatureQuery {
	return &intersectsMultiPolygonQuery{polygons: p}
}
