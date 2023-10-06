package ingest

import (
	"fmt"
	"math"

	"diagonal.works/b6"
	"diagonal.works/b6/geojson"
	"github.com/golang/geo/s2"
)

type AppliedChange map[b6.FeatureID]b6.FeatureID

type Change interface {
	Apply(w MutableWorld) (AppliedChange, error)
}

type AddFeatures struct {
	Points       []*PointFeature
	Paths        []*PathFeature
	Areas        []*AreaFeature
	Relations    []*RelationFeature
	Collections  []*CollectionFeature
	IDsToReplace map[b6.Namespace]b6.Namespace
}

func (a *AddFeatures) String() string {
	return fmt.Sprintf("added: points: %d paths: %d areas: %d relations: %d", len(a.Points), len(a.Paths), len(a.Areas), len(a.Relations))
}

func (a *AddFeatures) Apply(w MutableWorld) (AppliedChange, error) {
	newIDs := make(AppliedChange)
	for _, point := range a.Points {
		if ns, ok := a.IDsToReplace[point.PointID.Namespace]; ok {
			allocated := allocateID(WrapFeature(point, w), ns, w)
			newIDs[point.PointID.FeatureID()] = allocated
			point.PointID = allocated.ToPointID()
		}
		if err := w.AddPoint(point); err != nil {
			return nil, err
		}
	}

	for _, path := range a.Paths {
		if len(a.IDsToReplace) > 0 {
			for i := 0; i < path.Len(); i++ {
				if id, ok := path.PointID(i); ok {
					if allocated, ok := newIDs[id.FeatureID()]; ok {
						path.SetPointID(i, allocated.ToPointID())
					}
				}
			}
			if ns, ok := a.IDsToReplace[path.PathID.Namespace]; ok {
				allocated := allocateID(WrapFeature(path, w), ns, w)
				newIDs[path.PathID.FeatureID()] = allocated
				path.PathID = allocated.ToPathID()
			}
		}
		if err := w.AddPath(path); err != nil {
			return nil, err
		}
	}

	for _, area := range a.Areas {
		if len(a.IDsToReplace) > 0 {
			for i := 0; i < area.Len(); i++ {
				if ids, ok := area.PathIDs(i); ok {
					for j, id := range ids {
						if allocated, ok := newIDs[id.FeatureID()]; ok {
							area.SetPathID(i, j, allocated.ToPathID())
						}
					}
				}
			}
			if ns, ok := a.IDsToReplace[area.AreaID.Namespace]; ok {
				allocated := allocateID(WrapFeature(area, w), ns, w)
				newIDs[area.AreaID.FeatureID()] = allocated
				area.AreaID = allocated.ToAreaID()
			}
		}
		if err := w.AddArea(area); err != nil {
			return nil, err
		}
	}

	for _, relation := range a.Relations {
		if len(a.IDsToReplace) > 0 {
			for i := range relation.Members {
				if allocated, ok := newIDs[relation.Members[i].ID]; ok {
					relation.Members[i].ID = allocated
				}
			}
		}
		if _, ok := a.IDsToReplace[relation.RelationID.Namespace]; ok {
			return nil, fmt.Errorf("Can't allocate new IDs for relations: %s", relation.RelationID)
		}
		if err := w.AddRelation(relation); err != nil {
			return nil, err
		}
	}

	for _, collection := range a.Collections {
		// ID replacement not supported for collections.
		// TODO delete replacement for all features, no longer necessary.
		if err := w.AddCollection(collection); err != nil {
			return nil, err
		}
	}

	return newIDs, nil
}

func (a *AddFeatures) FillFromGeoJSON(g geojson.GeoJSON) {
	switch g := g.(type) {
	case *geojson.FeatureCollection:
		for i, f := range g.Features {
			a.fillFromFeature(f, uint64(i))
		}
	case *geojson.Feature:
		a.fillFromFeature(g, 0)
	}
}

func (a *AddFeatures) fillFromFeature(f *geojson.Feature, id uint64) {
	var feature Feature
	switch geometry := f.Geometry.Coordinates.(type) {
	case geojson.Point:
		point := NewPointFeature(b6.MakePointID(b6.NamespacePrivate, id), geometry.ToS2LatLng())
		a.Points = append(a.Points, point)
		feature = point
	case geojson.LineString:
		path := NewPathFeature(len(geometry))
		path.PathID = b6.MakePathID(b6.NamespacePrivate, id)
		for j, point := range geometry {
			path.SetLatLng(j, point.ToS2LatLng())
		}
		a.Paths = append(a.Paths, path)
		feature = path
	case geojson.Polygon:
		area := NewAreaFeature(1)
		area.AreaID = b6.MakeAreaID(b6.NamespacePrivate, id)
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
		a.Areas = append(a.Areas, area)
		feature = area
	case geojson.MultiPolygon:
		area := NewAreaFeature(len(geometry))
		area.AreaID = b6.MakeAreaID(b6.NamespacePrivate, id)
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
		a.Areas = append(a.Areas, area)
		feature = area
	}
	if feature != nil {
		for key, value := range f.Properties {
			feature.AddTag(b6.Tag{Key: key, Value: value})
		}
	}
}

func allocateID(f b6.Feature, ns b6.Namespace, byID b6.FeaturesByID) b6.FeatureID {
	var value s2.CellID
	switch f := f.(type) {
	case b6.PointFeature:
		value = s2.CellIDFromLatLng(s2.LatLngFromPoint(f.Point()))
	case b6.PathFeature:
		value = s2.CellIDFromLatLng(s2.LatLngFromPoint(f.Polyline().Centroid()))
	case b6.AreaFeature:
		centroids := make([]s2.Point, f.Len())
		for i := 0; i < f.Len(); i++ {
			centroids[i] = f.Polygon(i).Loop(0).Centroid()
		}
		loop := s2.LoopFromPoints(centroids)
		if loop.Area() > 2.0*math.Pi {
			loop.Invert()
		}
		value = s2.CellIDFromLatLng(s2.LatLngFromPoint(loop.Centroid()))
	default:
		panic(fmt.Sprintf("Can't allocate ID for %s", f.FeatureID().Type))
	}
	for {
		id := b6.FeatureID{Type: f.FeatureID().Type, Namespace: ns, Value: uint64(value)}
		if !byID.HasFeatureWithID(id) {
			return id
		}
		value = value.Advance(1)
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

func (a AddTags) Apply(w MutableWorld) (AppliedChange, error) {
	modified := make(AppliedChange)
	for _, t := range a {
		if err := w.AddTag(t.ID, t.Tag); err != nil {
			return nil, err
		}
		modified[t.ID] = t.ID
	}
	return modified, nil
}

type RemoveTag struct {
	ID  b6.FeatureID
	Key string
}

type RemoveTags []RemoveTag

func (r RemoveTags) String() string {
	return fmt.Sprintf("remove tags: %d", len(r))
}

func (r RemoveTags) Apply(w MutableWorld) (AppliedChange, error) {
	modified := make(AppliedChange)
	for _, t := range r {
		if err := w.RemoveTag(t.ID, t.Key); err != nil {
			return nil, err
		}
		modified[t.ID] = t.ID
	}
	return modified, nil
}

type MergedChange []Change

func (m MergedChange) Apply(w MutableWorld) (AppliedChange, error) {
	// To ensure the world is unmodified following failure, we first
	// apply the changes to a fresh overlay, and only change the
	// underlying world if there's no error
	canary := NewMutableOverlayWorld(w)
	for _, c := range m {
		if _, err := c.Apply(canary); err != nil {
			return nil, err
		}
	}
	applied := make(AppliedChange)
	for _, c := range m {
		a, err := c.Apply(w)
		if err != nil {
			return applied, fmt.Errorf("change partially applied: %s", err)
		}
		for before, after := range a {
			applied[before] = after
		}
	}
	return applied, nil
}
