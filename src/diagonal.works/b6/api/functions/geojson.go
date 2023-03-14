package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/ingest"

	"github.com/apache/beam/sdks/go/pkg/beam/io/filesystem"
	"github.com/golang/geo/s2"
)

func toGeoJSON(renderable b6.Renderable, c *api.Context) (geojson.GeoJSON, error) {
	if renderable != nil {
		return renderable.ToGeoJSON(), nil
	}
	return geojson.NewFeatureCollection(), nil
}

func toGeoJSONCollection(renderables api.AnyRenderableCollection, c *api.Context) (geojson.GeoJSON, error) {
	collection := geojson.NewFeatureCollection()
	var err error
	if renderables != nil {
		i := renderables.Begin()
		for {
			var ok bool
			ok, err = i.Next()
			if !ok || err != nil {
				break
			}
			if renderable, ok := i.Value().(b6.Renderable); ok {
				rendered := renderable.ToGeoJSON()
				switch r := rendered.(type) {
				case *geojson.Feature:
					collection.AddFeature(r)
				case *geojson.FeatureCollection:
					for _, f := range r.Features {
						collection.AddFeature(f)
					}
				case *geojson.Geometry:
					collection.AddFeature(geojson.NewFeatureWithGeometry(*r))
				}
			}
		}
	}
	return collection, err
}

type sequentialIDFactory struct {
	next uint64
}

func (s *sequentialIDFactory) AllocateForPoint(t b6.FeatureType, p s2.Point) b6.FeatureID {
	value := s.next
	s.next++
	return b6.FeatureID{Type: t, Namespace: b6.NamespacePrivate, Value: value}
}

func parseGeoJSON(s string, c *api.Context) (geojson.GeoJSON, error) {
	var collection geojson.FeatureCollection
	if err := json.Unmarshal([]byte(s), &collection); err != nil {
		return nil, err
	}
	return &collection, nil
}

func importGeoJSON(g geojson.GeoJSON, namespace string, c *api.Context) (ingest.Change, error) {
	add := &ingest.AddFeatures{
		IDsToReplace: map[b6.Namespace]b6.Namespace{
			b6.NamespacePrivate: b6.Namespace(namespace),
		},
	}
	add.FillFromGeoJSON(g)
	return add, nil
}

func importGeoJSONFile(filename string, namespace string, c *api.Context) (ingest.Change, error) {
	fs, err := filesystem.New(context.Background(), filename) // TODO: Pass the actual context
	if err != nil {
		return nil, err
	}

	f, err := fs.OpenRead(context.Background(), filename) // TODO: Pass the actual context
	if err != nil {
		return nil, err
	}

	var collection geojson.FeatureCollection
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&collection); err != nil {
		return nil, err
	}
	f.Close()

	add := &ingest.AddFeatures{
		IDsToReplace: map[b6.Namespace]b6.Namespace{
			b6.NamespacePrivate: b6.Namespace(namespace),
		},
	}
	add.FillFromGeoJSON(&collection)
	return add, nil
}

func geojsonAreas(g geojson.GeoJSON, c *api.Context) (api.StringAreaCollection, error) {
	polygons := g.ToS2Polygons()
	areas := &api.ArrayAreaCollection{
		Keys:   make([]string, len(polygons)),
		Values: make([]b6.Area, len(polygons)),
	}
	for i, p := range polygons {
		areas.Keys[i] = fmt.Sprintf("%d", i)
		for _, l := range p.Loops() {
			if l.Area() > 2.0*math.Pi {
				l.Invert()
			}
		}
		areas.Values[i] = b6.AreaFromS2Polygon(p)
	}
	return areas, nil
}

func applyToPoint(f func(b6.Point, *api.Context) (b6.Geometry, error), context *api.Context) func(b6.Geometry, *api.Context) (b6.Geometry, error) {
	return func(g b6.Geometry, context *api.Context) (b6.Geometry, error) {
		if point, ok := g.(b6.Point); ok {
			return f(point, context)
		}
		return g, nil
	}
}

func applyToPath(f func(b6.Path, *api.Context) (b6.Geometry, error), context *api.Context) func(b6.Geometry, *api.Context) (b6.Geometry, error) {
	return func(g b6.Geometry, context *api.Context) (b6.Geometry, error) {
		if path, ok := g.(b6.Path); ok {
			return f(path, context)
		}
		return g, nil
	}
}

func applyToArea(f func(b6.Area, *api.Context) (b6.Geometry, error), context *api.Context) func(b6.Geometry, *api.Context) (b6.Geometry, error) {
	return func(g b6.Geometry, context *api.Context) (b6.Geometry, error) {
		if area, ok := g.(b6.Area); ok {
			return f(area, context)
		}
		return g, nil
	}
}

func mapGeometry(g b6.Geometry, f func(b6.Geometry, *api.Context) (b6.Geometry, error), c *api.Context) (geojson.Coordinates, error) {
	var err error
	g, err = f(g, c)
	if err != nil {
		return nil, err
	}

	switch g := g.(type) {
	case b6.Point:
		return geojson.FromS2Point(g.Point()), nil
	case b6.Path:
		return geojson.FromPolyline(g.Polyline()), nil
	case b6.Area:
		if g.Len() == 1 {
			return geojson.FromPolygon(g.Polygon(0)), nil
		} else {
			polygons := make([]*s2.Polygon, g.Len())
			for i := range polygons {
				polygons[i] = g.Polygon(i)
			}
			return geojson.FromPolygons(polygons), nil
		}
	}
	return nil, fmt.Errorf("Can't map geometry of type %T", g)
}

func MapGeometries(g geojson.GeoJSON, f func(b6.Geometry, *api.Context) (b6.Geometry, error), c *api.Context) (geojson.GeoJSON, error) {
	m := func(cs geojson.Coordinates) (geojson.Coordinates, error) {
		switch cs := cs.(type) {
		case geojson.Point:
			return mapGeometry(b6.PointFromS2Point(cs.ToS2Point()), f, c)
		case geojson.LineString:
			return mapGeometry(b6.PathFromS2Points(cs.ToS2Polyline()), f, c)
		case geojson.Polygon:
			return mapGeometry(b6.AreaFromS2Polygon(cs.ToS2Polygon()), f, c)
		case geojson.MultiPolygon:
			return mapGeometry(b6.AreaFromS2Polygons(cs.ToS2Polygons()), f, c)
		}
		return cs, nil
	}
	return g.MapGeometries(m)
}
