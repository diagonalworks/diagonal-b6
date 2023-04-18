package renderer

import (
	"github.com/golang/geo/s2"
)

type Geometry interface {
	isGeometry()
}

type Point s2.Point

func NewPoint(p s2.Point) *Point {
	pp := Point(p)
	return &pp
}

func (p *Point) ToS2Point() s2.Point {
	return *((*s2.Point)(p))
}

func (p *Point) isGeometry() {}

type LineString s2.Polyline

func NewLineString(l *s2.Polyline) *LineString {
	return (*LineString)(l)
}

func (l *LineString) isGeometry() {}

func (t *LineString) ToS2Polyline() *s2.Polyline {
	return (*s2.Polyline)(t)
}

type Polygon s2.Polygon

func NewPolygon(p *s2.Polygon) *Polygon {
	return (*Polygon)(p)
}

func (p *Polygon) ToS2Polygon() *s2.Polygon {
	return (*s2.Polygon)(p)
}

func (p *Polygon) isGeometry() {}

type Feature struct {
	Geometry Geometry
	ID       uint64
	Tags     map[string]string
}

func NewFeature(g Geometry) *Feature {
	return &Feature{Geometry: g, Tags: make(map[string]string)}
}

type Layer struct {
	Name     string
	Features []*Feature
}

func NewLayer(name string) *Layer {
	return &Layer{Name: name, Features: make([]*Feature, 0, 4)}
}

func (l *Layer) AddFeature(f *Feature) {
	l.Features = append(l.Features, f)
}

func (l *Layer) AddFeatures(fs []*Feature) {
	l.Features = append(l.Features, fs...)
}

type Tile struct {
	Layers []*Layer
}

func (t *Tile) FindLayer(name string) *Layer {
	for _, layer := range t.Layers {
		if layer.Name == name {
			return layer
		}
	}
	return nil
}
