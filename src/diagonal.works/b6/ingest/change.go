package ingest

import (
	"fmt"
	"math"

	"diagonal.works/b6"
	"diagonal.works/b6/geojson"
	"github.com/golang/geo/s2"
)

type Change interface {
	Apply(w MutableWorld) (b6.Collection[b6.FeatureID, b6.FeatureID], error)
}

type AddFeatures []Feature

func (a *AddFeatures) String() string {
	return fmt.Sprintf("added %d features", len(*a))
}

func (a *AddFeatures) Apply(w MutableWorld) (b6.Collection[b6.FeatureID, b6.FeatureID], error) {
	features := make(map[b6.FeatureID]b6.FeatureID)
	for _, feature := range *a {
		if err := w.AddFeature(feature); err != nil {
			return b6.ArrayCollection[b6.FeatureID, b6.FeatureID]{}.Collection(), err
		}

		features[feature.FeatureID()] = feature.FeatureID()
	}

	c := b6.ArrayCollection[b6.FeatureID, b6.FeatureID]{
		Keys:   make([]b6.FeatureID, 0, len(features)),
		Values: make([]b6.FeatureID, 0, len(features)),
	}
	for k, v := range features {
		c.Keys = append(c.Keys, k)
		c.Values = append(c.Values, v)
	}

	return c.Collection(), nil
}

func (a *AddFeatures) FillFromGeoJSON(g geojson.GeoJSON, namespace b6.Namespace) {
	switch g := g.(type) {
	case *geojson.FeatureCollection:
		for i, f := range g.Features {
			a.fillFromFeature(f, namespace, uint64(i))
		}
	case *geojson.Feature:
		a.fillFromFeature(g, namespace, 0)
	}
}

func (a *AddFeatures) fillFromFeature(f *geojson.Feature, namespace b6.Namespace, id uint64) {
	var feature Feature

	switch geometry := f.Geometry.Coordinates.(type) {
	case geojson.Point:
		feature = &GenericFeature{ID: b6.FeatureID{b6.FeatureTypePoint, namespace, id}, Tags: []b6.Tag{{Key: b6.PointTag, Value: b6.NewPointExpressionFromLatLng(s2.LatLngFromDegrees(geometry.Lat, geometry.Lng))}}}
	case geojson.LineString:
		path := &GenericFeature{ID: b6.FeatureID{b6.FeatureTypePath, namespace, id}}
		for j, point := range geometry {
			path.ModifyOrAddTagAt(b6.Tag{b6.PathTag, b6.NewPointExpressionFromLatLng(point.ToS2LatLng())}, j)
		}
		feature = path
	case geojson.Polygon:
		area := NewAreaFeature(1)
		area.AreaID = b6.MakeAreaID(namespace, id)
		loops := make([]*s2.Loop, len(geometry))
		for j, loop := range geometry {
			points := make([]s2.Point, len(loop))
			for k, point := range loop {
				points[k] = point.ToS2Point()
			}
			loops[j] = s2.LoopFromPoints(points)
			if loops[j].Area() > 2.0*math.Pi {
				loops[j].Invert()
			}
		}
		area.SetPolygon(0, s2.PolygonFromLoops(loops))
		feature = area
	case geojson.MultiPolygon:
		area := NewAreaFeature(len(geometry))
		area.AreaID = b6.MakeAreaID(namespace, id)
		for j, polygon := range geometry {
			loops := make([]*s2.Loop, len(polygon))
			for k, loop := range polygon {
				points := make([]s2.Point, len(loop))
				for l, point := range loop {
					points[l] = point.ToS2Point()
				}
				loops[k] = s2.LoopFromPoints(points)
				if loops[k].Area() > 2.0*math.Pi {
					loops[k].Invert()
				}
			}
			area.SetPolygon(j, s2.PolygonFromLoops(loops))
		}
		feature = area
	}

	if feature != nil {
		*a = append(*a, feature)

		for key, value := range f.Properties {
			feature.AddTag(b6.Tag{Key: key, Value: b6.NewStringExpression(value)})
		}
	}
}

type AddTag struct {
	ID  b6.FeatureID
	Tag b6.Tag
}

type AddTags []AddTag

func (a AddTags) String() string {
	return fmt.Sprintf("set tags: %d", len(a))
}

func (a AddTags) Apply(w MutableWorld) (b6.Collection[b6.FeatureID, b6.FeatureID], error) {
	modified := b6.ArrayCollection[b6.FeatureID, b6.FeatureID]{}
	for _, t := range a {
		if err := w.AddTag(t.ID, t.Tag); err != nil {
			return modified.Collection(), err
		}
		modified.Keys = append(modified.Keys, t.ID)
		modified.Values = append(modified.Values, t.ID)
	}
	return modified.Collection(), nil
}

type RemoveTag struct {
	ID  b6.FeatureID
	Key string
}

type RemoveTags []RemoveTag

func (r RemoveTags) String() string {
	return fmt.Sprintf("remove tags: %d", len(r))
}

func (r RemoveTags) Apply(w MutableWorld) (b6.Collection[b6.FeatureID, b6.FeatureID], error) {
	modified := b6.ArrayCollection[b6.FeatureID, b6.FeatureID]{}
	for _, t := range r {
		if err := w.RemoveTag(t.ID, t.Key); err != nil {
			return modified.Collection(), err
		}
		modified.Keys = append(modified.Keys, t.ID)
		modified.Values = append(modified.Values, t.ID)
	}
	return modified.Collection(), nil
}

type MergedChange []Change

func (m MergedChange) Apply(w MutableWorld) (b6.Collection[b6.FeatureID, b6.FeatureID], error) {
	// To ensure the world is unmodified following failure, we first
	// apply the changes to a fresh overlay, and only change the
	// underlying world if there's no error
	applied := b6.ArrayCollection[b6.FeatureID, b6.FeatureID]{}
	canary := NewMutableOverlayWorld(w)
	for _, c := range m {
		if _, err := c.Apply(canary); err != nil {
			return applied.Collection(), err
		}
	}
	for _, c := range m {
		a, err := c.Apply(w)
		if err != nil {
			return applied.Collection(), fmt.Errorf("change partially applied: %s", err)
		}
		if applied.Keys, err = a.AllKeys(applied.Keys); err != nil {
			panic(fmt.Sprintf("broken keys: %s", err))
		}
		if applied.Values, err = a.AllKeys(applied.Values); err != nil {
			panic(fmt.Sprintf("broken values: %s", err))
		}
	}
	return applied.Collection(), nil
}
