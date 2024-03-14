package functions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/ingest"

	"github.com/apache/beam/sdks/go/pkg/beam/io/filesystem"
	"github.com/golang/geo/s2"
)

func toGeoJSON(c *api.Context, renderable b6.Geometry) (geojson.GeoJSON, error) {
	if renderable != nil {
		return renderable.ToGeoJSON(), nil
	}
	return geojson.NewFeatureCollection(), nil
}

func toGeoJSONCollection(c *api.Context, renderables b6.Collection[interface{}, b6.Geometry]) (geojson.GeoJSON, error) {
	collection := geojson.NewFeatureCollection()
	var err error
	i := renderables.Begin()
	for {
		var ok bool
		ok, err = i.Next()
		if !ok || err != nil {
			break
		}
		rendered := i.Value().ToGeoJSON()
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

// Return the geojson represented by the given string.
func parseGeoJSON(c *api.Context, s string) (geojson.GeoJSON, error) {
	g, err := geojson.Unmarshal([]byte(s))
	if err != nil {
		return nil, fmt.Errorf("failed to parse geojson: %s", err)
	}
	return g, nil
}

// Return the geojson contained in the given file.
// As the file is read by the b6 server process, the filename it relative
// to the filesystems it sees. Reading from files on cloud storage is
// supported.
func parseGeoJSONFile(c *api.Context, filename string) (geojson.GeoJSON, error) {
	if !c.FileIOAllowed {
		return nil, fmt.Errorf("File IO is not allowed")
	}

	fs, err := filesystem.New(c.Context, filename)
	if err != nil {
		return nil, err
	}

	f, err := fs.OpenRead(c.Context, filename)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	f.Close()

	g, err := geojson.Unmarshal(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse geojson: %s", err)
	}
	return g, nil
}

// Add features from the given geojson to the world.
// IDs are formed from the given namespace, and the index of the feature
// within the geojson collection (or 0, if a single feature is used).
func importGeoJSON(c *api.Context, features geojson.GeoJSON, namespace string) (ingest.Change, error) {
	add := &ingest.AddFeatures{}
	add.FillFromGeoJSON(features)
	return add, nil
}

// Add features from the given geojson file to the world.
// IDs are formed from the given namespace, and the index of the feature
// within the geojson collection (or 0, if a single feature is used).
// As the file is read by the b6 server process, the filename it relative
// to the filesystems it sees. Reading from files on cloud storage is
// supported.
func importGeoJSONFile(c *api.Context, filename string, namespace string) (ingest.Change, error) {
	if !c.FileIOAllowed {
		return nil, fmt.Errorf("File IO is not allowed")
	}

	fs, err := filesystem.New(c.Context, filename)
	if err != nil {
		return nil, err
	}

	f, err := fs.OpenRead(c.Context, filename)
	if err != nil {
		return nil, err
	}

	var collection geojson.FeatureCollection
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&collection); err != nil {
		return nil, err
	}
	f.Close()

	add := &ingest.AddFeatures{}
	add.FillFromGeoJSON(&collection)
	return add, nil
}

// Return the areas present in the given geojson.
func geojsonAreas(c *api.Context, g geojson.GeoJSON) (b6.Collection[int, b6.Area], error) {
	polygons := g.ToS2Polygons()
	collection := b6.ArrayValuesCollection[b6.Area]{}
	for _, p := range polygons {
		if err := p.Validate(); err == nil {
			if p.NumLoops() > 0 && p.Loop(0).Area() > 2.0*math.Pi {
				p.Invert()
			}
			collection = append(collection, b6.AreaFromS2Polygon(p))
		}
	}
	return collection.Collection(), nil
}

// Wrap the given function such that it will only be called when passed a point.
func applyToPoint(context *api.Context, f func(*api.Context, b6.Geometry) (b6.Geometry, error)) func(*api.Context, b6.Geometry) (b6.Geometry, error) {
	return func(context *api.Context, g b6.Geometry) (b6.Geometry, error) {
		if g.GeometryType() == b6.GeometryTypePoint {
			return f(context, g)
		}
		return g, nil
	}
}

// Wrap the given function such that it will only be called when passed a path.
func applyToPath(context *api.Context, f func(*api.Context, b6.Path) (b6.Geometry, error)) func(*api.Context, b6.Geometry) (b6.Geometry, error) {
	return func(context *api.Context, g b6.Geometry) (b6.Geometry, error) {
		if path, ok := g.(b6.Path); ok {
			return f(context, path)
		}
		return g, nil
	}
}

// Wrap the given function such that it will only be called when passed an area.
func applyToArea(context *api.Context, f func(*api.Context, b6.Area) (b6.Geometry, error)) func(*api.Context, b6.Geometry) (b6.Geometry, error) {
	return func(context *api.Context, g b6.Geometry) (b6.Geometry, error) {
		if area, ok := g.(b6.Area); ok {
			return f(context, area)
		}
		return g, nil
	}
}

func mapGeometry(g b6.Geometry, f func(*api.Context, b6.Geometry) (b6.Geometry, error), c *api.Context) (geojson.Coordinates, error) {
	var err error
	g, err = f(c, g)
	if err != nil {
		return nil, err
	}

	switch g := g.(type) {
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
	default:
		return geojson.FromS2LatLng(g.Location()), nil
	}
}

// Return a geojson representing the result of applying the given function to each geometry in the given geojson.
func mapGeometries(c *api.Context, g geojson.GeoJSON, f func(*api.Context, b6.Geometry) (b6.Geometry, error)) (geojson.GeoJSON, error) {
	m := func(cs geojson.Coordinates) (geojson.Coordinates, error) {
		switch cs := cs.(type) {
		case geojson.Point:
			return mapGeometry(b6.GeometryFromLatLng(s2.LatLngFromPoint(cs.ToS2Point())), f, c)
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
