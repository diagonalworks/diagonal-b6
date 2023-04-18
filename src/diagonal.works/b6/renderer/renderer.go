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
	BasemapLayerWater BasemapLayer = iota
	BasemapLayerRoad
	BasemapLayerLandUse
	BasemapLayerContour
	BasemapLayerBuilding
	BasemapLayerRoute
	BasemapLayerLabel
	BasemapLayerInvalid

	BasemapLayerBegin = BasemapLayerWater
	BasemapLayerEnd   = BasemapLayerInvalid
)

type BasemapLayers [BasemapLayerEnd]*Layer

func NewLayers() *BasemapLayers {
	var l BasemapLayers
	l[BasemapLayerWater] = NewLayer("water")
	l[BasemapLayerRoad] = NewLayer("road")
	l[BasemapLayerLandUse] = NewLayer("landuse")
	l[BasemapLayerContour] = NewLayer("contour")
	l[BasemapLayerBuilding] = NewLayer("building")
	l[BasemapLayerRoute] = NewLayer("route")
	l[BasemapLayerLabel] = NewLayer("label")
	return &l
}

type BasemapRenderer struct {
	World b6.World
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
	regionQuery := b6.MightIntersect{bounds}
	var typeQuery b6.Query
	if tile.Z >= 16 {
		typeQuery = b6.Union{
			b6.Tagged{Key: "#highway", Value: "footway"},
			b6.Tagged{Key: "#highway", Value: "path"},
			b6.Tagged{Key: "#highway", Value: "pedestrian"},
			b6.Tagged{Key: "#highway", Value: "residential"},
			b6.Tagged{Key: "#highway", Value: "unclassified"},
			b6.Tagged{Key: "#highway", Value: "cycleway"},
			b6.Tagged{Key: "#highway", Value: "service"},
			b6.Tagged{Key: "#highway", Value: "street"},
			b6.Tagged{Key: "#highway", Value: "tertiary"},
			b6.Tagged{Key: "#highway", Value: "secondary"},
			b6.Tagged{Key: "#highway", Value: "primary"},
			b6.Tagged{Key: "#highway", Value: "trunk"},
			b6.Tagged{Key: "#highway", Value: "motorway"},
			b6.Tagged{Key: "#landuse", Value: "cemetary"},
			b6.Tagged{Key: "#landuse", Value: "forest"},
			b6.Tagged{Key: "#landuse", Value: "grass"},
			b6.Tagged{Key: "#landuse", Value: "meadow"},
			b6.Tagged{Key: "#landuse", Value: "vacant"},
			b6.Tagged{Key: "#leisure", Value: "park"},
			b6.Tagged{Key: "#leisure", Value: "pitch"},
			b6.Tagged{Key: "#natural", Value: "heath"},
			b6.Tagged{Key: "#outline", Value: "contour"},
			b6.Keyed{Key: "#building"},
			b6.Keyed{Key: "#water"},
			b6.Keyed{Key: "#waterway"},
		}
	} else if tile.Z >= 14 {
		typeQuery = b6.Union{
			b6.Tagged{Key: "#highway", Value: "residential"},
			b6.Tagged{Key: "#highway", Value: "unclassified"},
			b6.Tagged{Key: "#highway", Value: "service"},
			b6.Tagged{Key: "#highway", Value: "street"},
			b6.Tagged{Key: "#highway", Value: "tertiary"},
			b6.Tagged{Key: "#highway", Value: "secondary"},
			b6.Tagged{Key: "#highway", Value: "primary"},
			b6.Tagged{Key: "#highway", Value: "trunk"},
			b6.Tagged{Key: "#highway", Value: "motorway"},
			b6.Tagged{Key: "#landuse", Value: "cemetary"},
			b6.Tagged{Key: "#landuse", Value: "forest"},
			b6.Tagged{Key: "#landuse", Value: "grass"},
			b6.Tagged{Key: "#landuse", Value: "meadow"},
			b6.Tagged{Key: "#landuse", Value: "vacant"},
			b6.Tagged{Key: "#leisure", Value: "park"},
			b6.Tagged{Key: "#leisure", Value: "pitch"},
			b6.Tagged{Key: "#natural", Value: "heath"},
			b6.Tagged{Key: "#outline", Value: "contour"},
			b6.Keyed{Key: "#building"},
			b6.Keyed{Key: "#water"},
			b6.Keyed{Key: "#waterway"},
		}
	} else if tile.Z >= 12 {
		typeQuery = b6.Union{
			b6.Tagged{Key: "#highway", Value: "trunk"},
			b6.Tagged{Key: "#highway", Value: "primary"},
			b6.Tagged{Key: "#highway", Value: "motorway"},
			b6.Keyed{Key: "#water"},
		}
	} else {
		typeQuery = b6.Union{
			b6.Tagged{Key: "#place", Value: "city"},
		}
	}

	var q b6.Query
	if typeQuery != nil {
		q = b6.Intersection{regionQuery, typeQuery}
	} else {
		q = regionQuery
	}
	features := b6.AllFeatures(b.World.FindFeatures(q))
	sort.Sort(byLayerThenID(features))
	return features
}

// TODO: Include zoom level here, and automatically create the query above
var renderRules = []struct {
	Tag       b6.Tag
	Attribute string
	Layer     BasemapLayer
}{
	{Tag: b6.Tag{Key: "#highway"}, Attribute: "highway", Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#building"}, Attribute: "building", Layer: BasemapLayerBuilding},
	{Tag: b6.Tag{Key: "#landuse"}, Attribute: "landuse", Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#water"}, Attribute: "water", Layer: BasemapLayerWater},
	{Tag: b6.Tag{Key: "#waterway"}, Attribute: "waterway", Layer: BasemapLayerWater},
	{Tag: b6.Tag{Key: "#railway"}, Attribute: "railway", Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#natural"}, Attribute: "natural", Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#leisure"}, Attribute: "leisure", Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#place", Value: "city"}, Attribute: "leisure", Layer: BasemapLayerLabel},
	{Tag: b6.Tag{Key: "#outline", Value: "contour"}, Attribute: "outline", Layer: BasemapLayerContour},
}

func (b *BasemapRenderer) renderFeature(f b6.Feature, layers *BasemapLayers, fs []*Feature) []*Feature {
	var tags [1]b6.Tag
	for _, rule := range renderRules {
		if t := f.Get(rule.Tag.Key); t.IsValid() && (rule.Tag.Value == "" || t.Value == rule.Tag.Value) {
			tags[0] = b6.Tag{Key: rule.Attribute, Value: t.Value}
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
	for _, tf := range tfs {
		fillNameFromFeature(tf, f)
		fillColourFromFeature(tf, f)
		fillIDFromFeature(tf, f)
	}
	return tfs
}

func fillFeaturesFromPoint(point b6.PointFeature, tags []b6.Tag, fs []*Feature) []*Feature {
	f := NewFeature(NewPoint(point.Point()))
	f.ID = api.TileFeatureID(point.FeatureID())
	for _, t := range tags {
		f.Tags[t.Key] = t.Value
	}
	return append(fs, f)
}

func fillFeaturesFromPath(path b6.PathFeature, tags []b6.Tag, fs []*Feature) []*Feature {
	f := NewFeature(NewLineString(path.Polyline()))
	f.ID = api.TileFeatureID(path.FeatureID())
	for _, t := range tags {
		f.Tags[t.Key] = t.Value
	}
	return append(fs, f)
}

func fillFeaturesFromArea(area b6.AreaFeature, tags []b6.Tag, fs []*Feature) []*Feature {
	for i := 0; i < area.Len(); i++ {
		f := NewFeature(NewPolygon(area.Polygon(i)))
		f.ID = api.TileFeatureIDForPolygon(area.FeatureID(), i)
		for _, t := range tags {
			f.Tags[t.Key] = t.Value
		}
		fs = append(fs, f)
	}
	return fs
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
