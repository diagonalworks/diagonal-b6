package renderer

import (
	"sort"
	"strconv"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
)

type Renderer interface {
	Render(tile b6.Tile, args *TileArgs) (*Tile, error)
}

type byLayerThenID []b6.Feature

func (b byLayerThenID) Len() int      { return len(b) }
func (b byLayerThenID) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byLayerThenID) Less(i, j int) bool {
	li, lj := layerNumber(b[i]), layerNumber(b[j])
	if li == lj {
		return b[i].FeatureID().Less(b[j].FeatureID())
	}
	return li < lj
}

func layerNumber(f b6.Feature) int {
	if layer := f.Get("layer"); layer.IsValid() {
		if i, ok := layer.IntValue(); ok {
			return i
		}
	}
	return 0
}

type BasemapLayer int

const (
	BasemapLayerBoundary BasemapLayer = iota
	BasemapLayerContour
	BasemapLayerWater
	BasemapLayerRoad
	BasemapLayerLandUse
	BasemapLayerBuilding
	BasemapLayerPoint
	BasemapLayerLabel
	BasemapLayerInvalid

	BasemapLayerBegin = BasemapLayerBoundary
	BasemapLayerEnd   = BasemapLayerInvalid
)

type BasemapLayers [BasemapLayerEnd]*Layer

func NewLayers() *BasemapLayers {
	var l BasemapLayers
	l[BasemapLayerBoundary] = NewLayer("boundary")
	l[BasemapLayerContour] = NewLayer("contour")
	l[BasemapLayerWater] = NewLayer("water")
	l[BasemapLayerRoad] = NewLayer("road")
	l[BasemapLayerLandUse] = NewLayer("landuse")
	l[BasemapLayerBuilding] = NewLayer("building")
	l[BasemapLayerPoint] = NewLayer("point")
	l[BasemapLayerLabel] = NewLayer("label")
	return &l
}

type RenderRule struct {
	Tag     b6.Tag
	MinZoom uint
	MaxZoom uint
	Layer   BasemapLayer
}

func (r *RenderRule) ToQuery(zoom uint) (b6.Query, bool) {
	if (r.MinZoom > 0 && zoom < r.MinZoom) || (r.MaxZoom > 0 && zoom > r.MaxZoom) {
		return nil, false
	}
	if r.Tag.Value != "" {
		return b6.Tagged(r.Tag), true
	} else {
		return b6.Keyed{Key: r.Tag.Key}, true
	}
}

type RenderRules []RenderRule

func (rs RenderRules) ToQuery(zoom uint) b6.Query {
	union := make(b6.Union, 0, len(rs))
	for _, r := range rs {
		if q, ok := r.ToQuery(zoom); ok {
			union = append(union, q)
		}
	}
	return union
}

func (rs RenderRules) IsRendered(tag b6.Tag) bool {
	for _, r := range rs {
		if r.Tag.Key == tag.Key {
			if r.Tag.Value == "" || r.Tag.Value == tag.Value {
				return true
			}
		}
	}
	return false
}

var BasemapRenderRules = RenderRules{
	{Tag: b6.Tag{Key: "#building"}, MinZoom: 14, Layer: BasemapLayerBuilding},
	{Tag: b6.Tag{Key: "#highway", Value: "cycleway"}, MinZoom: 16, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: "footway"}, MinZoom: 16, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: "motorway"}, MinZoom: 12, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: "path"}, MinZoom: 16, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: "pedestrian"}, MinZoom: 16, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: "primary"}, MinZoom: 12, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: "residential"}, MinZoom: 14, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: "secondary"}, MinZoom: 16, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: "service"}, MinZoom: 14, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: "street"}, MinZoom: 14, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: "tertiary"}, MinZoom: 14, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: "trunk"}, MinZoom: 12, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: "unclassified"}, MinZoom: 14, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#landuse", Value: "cemetary"}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#landuse", Value: "forest"}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#landuse", Value: "grass"}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#landuse", Value: "heath"}, MinZoom: 16, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#landuse", Value: "meadow"}, MinZoom: 16, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#landuse", Value: "park"}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#landuse", Value: "pitch"}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#landuse", Value: "vacant"}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#leisure", Value: "park"}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#leisure", Value: "pitch"}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#leisure", Value: "playground"}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#leisure", Value: "garden"}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#leisure", Value: "nature_reserve"}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#natural", Value: "heath"}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#outline", Value: "contour"}, MinZoom: 14, Layer: BasemapLayerContour},
	{Tag: b6.Tag{Key: "#water"}, MinZoom: 12, Layer: BasemapLayerWater},
	{Tag: b6.Tag{Key: "#waterway"}, MinZoom: 14, Layer: BasemapLayerWater},
	{Tag: b6.Tag{Key: "#place", Value: "city"}, MaxZoom: 11, Layer: BasemapLayerLabel},
}

type BasemapRenderer struct {
	RenderRules RenderRules
	World       b6.World
}

func (b *BasemapRenderer) Render(tile b6.Tile, args *TileArgs) (*Tile, error) {
	features := b.findFeatures(tile)
	layers := NewLayers()
	fs := make([]*Feature, 0, 2)
	for _, feature := range features {
		fs = b.renderFeature(feature, layers, fs[0:0])
	}
	return &Tile{Layers: (*layers)[0:]}, nil
}

func (b *BasemapRenderer) findFeatures(tile b6.Tile) []b6.Feature {
	bounds := tile.RectBound()
	regionQuery := b6.MightIntersect{Region: bounds}
	q := b6.Intersection{b.RenderRules.ToQuery(tile.Z), regionQuery}
	features := b6.AllFeatures(b.World.FindFeatures(q))
	sort.Sort(byLayerThenID(features))
	return features
}

func (b *BasemapRenderer) renderFeature(f b6.Feature, layers *BasemapLayers, fs []*Feature) []*Feature {
	var tags [1]b6.Tag
	for _, rule := range b.RenderRules {
		if t := f.Get(rule.Tag.Key); t.IsValid() && (rule.Tag.Value == "" || t.Value == rule.Tag.Value) {
			tags[0] = b6.Tag{Key: rule.Tag.Key[1:], Value: t.Value}
			fs = FillFeaturesFromFeature(f, tags[0:], fs)
			layers[rule.Layer].AddFeatures(fs)
			break
		}
	}
	return fs
}

func FillFeaturesFromFeature(f b6.Feature, tags []b6.Tag, tfs []*Feature) []*Feature {
	switch f := f.(type) {
	case b6.PointFeature:
		tfs = fillFeaturesFromPoint(f, tags, tfs)
	case b6.PathFeature:
		tfs = fillFeaturesFromPath(f, tags, tfs)
	case b6.AreaFeature:
		tfs = fillFeaturesFromArea(f, tags, tfs)
	}
	return tfs
}

func fillFeaturesFromPoint(point b6.PointFeature, tags []b6.Tag, fs []*Feature) []*Feature {
	f := NewFeature(NewPoint(point.Point()))
	f.ID = api.TileFeatureID(point.FeatureID())
	for _, t := range tags {
		f.Tags[t.Key] = t.Value
	}
	fillTagsFromTags(f, point)
	return append(fs, f)
}

func fillFeaturesFromPath(path b6.PathFeature, tags []b6.Tag, fs []*Feature) []*Feature {
	f := NewFeature(NewLineString(path.Polyline()))
	f.ID = api.TileFeatureID(path.FeatureID())
	for _, t := range tags {
		f.Tags[t.Key] = t.Value
	}
	fillTagsFromTags(f, path)
	return append(fs, f)
}

func fillFeaturesFromArea(area b6.AreaFeature, tags []b6.Tag, fs []*Feature) []*Feature {
	for i := 0; i < area.Len(); i++ {
		f := NewFeature(NewPolygon(area.Polygon(i)))
		f.ID = api.TileFeatureIDForPolygon(area.FeatureID(), i)
		for _, t := range tags {
			f.Tags[t.Key] = t.Value
		}
		fillTagsFromTags(f, area)
		fs = append(fs, f)
	}
	return fs
}

func fillTagsFromTags(tf *Feature, f b6.Feature) {
	fillNameFromFeature(tf, f)
	fillColourFromFeature(tf, f)
	fillIDFromFeature(tf, f)
}

func fillNameFromFeature(tf *Feature, f b6.Feature) {
	if name := f.Get("addr:housename"); name.IsValid() {
		tf.Tags["name"] = name.Value
	} else if name := f.Get("name"); name.IsValid() {
		tf.Tags["name"] = name.Value
	}
}

func fillIDFromFeature(tf *Feature, f b6.Feature) {
	// We split the namespace from the ID to reduce tile size, as namespaces repeat
	// within a tile, and attribute values are compressed in the encoded tile via a
	// string table.
	tf.Tags["id"] = strconv.FormatUint(f.FeatureID().Value, 16)
	tf.Tags["ns"] = f.FeatureID().Namespace.String()
}

// Blue to red gradient from simona@diagonal.works
var gradient = Gradient{
	{Value: 0.0, Colour: ColourFromHexString("#d3d6fd")},
	{Value: 0.30, Colour: ColourFromHexString("#fca364")},
	{Value: 0.60, Colour: ColourFromHexString("#f88a4f")},
	{Value: 1.00, Colour: ColourFromHexString("#f96c53")},
}

// DiagonalColour is the tag key used to explicitly set the colour of features
const DiagonalColour = "diagonal:colour"

func colourFromTagValue(v string) string {
	if len(v) == 7 && v[0] == '#' {
		return ColourFromHexString(v).ToHexString() // Roundtrip to sanitise
	} else if point, err := strconv.ParseFloat(v, 64); err == nil {
		return gradient.Interpolate(point).ToHexString()
	}
	return ""
}

func fillColourFromFeature(tf *Feature, f b6.Taggable) {
	if colour := f.Get(DiagonalColour); colour.IsValid() {
		if converted := colourFromTagValue(colour.Value); converted != "" {
			tf.Tags["colour"] = converted
		}
	}
}
