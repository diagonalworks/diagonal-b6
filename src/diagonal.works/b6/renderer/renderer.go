package renderer

import (
	"fmt"
	"sort"
	"strconv"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"github.com/golang/geo/s2"
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
		if i, err := strconv.Atoi(layer.Value.String()); err == nil {
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

func (b BasemapLayer) String() string {
	switch b {
	case BasemapLayerBoundary:
		return "boundary"
	case BasemapLayerContour:
		return "contour"
	case BasemapLayerWater:
		return "water"
	case BasemapLayerRoad:
		return "road"
	case BasemapLayerLandUse:
		return "landuse"
	case BasemapLayerBuilding:
		return "building"
	case BasemapLayerPoint:
		return "point"
	case BasemapLayerLabel:
		return "label"
	}
	return "invalid"
}

func (b BasemapLayer) MarshalYAML() (interface{}, error) {
	return b.String(), nil
}

func (b *BasemapLayer) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	for l := BasemapLayerBegin; l <= BasemapLayerEnd; l++ {
		if l.String() == s {
			*b = l
			return nil
		}
	}
	return fmt.Errorf("bad layer value: %s", s)
}

type BasemapLayers [BasemapLayerEnd]*Layer

func NewLayers() *BasemapLayers {
	var ls BasemapLayers
	for l := BasemapLayerBegin; l < BasemapLayerEnd; l++ {
		ls[l] = NewLayer(l.String())
	}
	return &ls
}

type RenderRule struct {
	Tag     b6.Tag
	MinZoom uint `yaml:"min,omitempty"`
	MaxZoom uint `yaml:"max,omitempty"`
	Layer   BasemapLayer
	Label   bool `yaml:",omitempty"`
	Icon    bool `yaml:",omitempty"`
}

func (r *RenderRule) ToQuery(zoom uint) (b6.Query, bool) {
	if (r.MinZoom > 0 && zoom < r.MinZoom) || (r.MaxZoom > 0 && zoom > r.MaxZoom) {
		return nil, false
	}
	if r.Tag.IsValid() && r.Tag.Value.String() != "" {
		return b6.Tagged(r.Tag), true
	} else {
		return b6.Keyed{Key: r.Tag.Key}, true
	}
}

func (r *RenderRule) Matches(f b6.Taggable) (b6.Value, bool) {
	t := f.Get(r.Tag.Key)
	if !t.IsValid() {
		return nil, false
	}
	matches := r.Tag.Value == nil || r.Tag.Value.String() == "" || r.Tag.Value.String() == t.Value.String()
	return t.Value, matches
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
			if r.Tag.Value == nil || r.Tag.Value.String() == "" || r.Tag.Value.String() == tag.Value.String() {
				return true
			}
		}
	}
	return false
}

var BasemapRenderRules = RenderRules{
	{Tag: b6.Tag{Key: "#building", Value: b6.String("train_station")}, MinZoom: 12, Layer: BasemapLayerBuilding},
	{Tag: b6.Tag{Key: "#building"}, MinZoom: 14, Layer: BasemapLayerBuilding},
	{Tag: b6.Tag{Key: "#highway", Value: b6.String("cycleway")}, MinZoom: 16, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: b6.String("footway")}, MinZoom: 16, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: b6.String("motorway")}, MinZoom: 12, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: b6.String("path")}, MinZoom: 16, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: b6.String("pedestrian")}, MinZoom: 16, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: b6.String("primary")}, MinZoom: 12, Layer: BasemapLayerRoad, Label: true},
	{Tag: b6.Tag{Key: "#highway", Value: b6.String("residential")}, MinZoom: 14, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: b6.String("secondary")}, MinZoom: 16, Layer: BasemapLayerRoad, Label: true},
	{Tag: b6.Tag{Key: "#highway", Value: b6.String("service")}, MinZoom: 14, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: b6.String("street")}, MinZoom: 14, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#highway", Value: b6.String("tertiary")}, MinZoom: 14, Layer: BasemapLayerRoad, Label: true},
	{Tag: b6.Tag{Key: "#highway", Value: b6.String("trunk")}, MinZoom: 12, Layer: BasemapLayerRoad, Label: true},
	{Tag: b6.Tag{Key: "#highway", Value: b6.String("unclassified")}, MinZoom: 14, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#landuse", Value: b6.String("cemetary")}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#landuse", Value: b6.String("forest")}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#landuse", Value: b6.String("grass")}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#landuse", Value: b6.String("heath")}, MinZoom: 16, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#landuse", Value: b6.String("meadow")}, MinZoom: 16, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#landuse", Value: b6.String("park")}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#landuse", Value: b6.String("pitch")}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#landuse", Value: b6.String("vacant")}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#leisure", Value: b6.String("park")}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#leisure", Value: b6.String("pitch")}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#leisure", Value: b6.String("playground")}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#leisure", Value: b6.String("garden")}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#leisure", Value: b6.String("nature_reserve")}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#natural", Value: b6.String("coastline")}, MinZoom: 12, Layer: BasemapLayerBoundary},
	{Tag: b6.Tag{Key: "#natural", Value: b6.String("heath")}, MinZoom: 14, Layer: BasemapLayerLandUse},
	{Tag: b6.Tag{Key: "#outline", Value: b6.String("contour")}, MinZoom: 14, Layer: BasemapLayerContour},
	{Tag: b6.Tag{Key: "#railway", Value: b6.String("rail")}, MinZoom: 12, Layer: BasemapLayerRoad},
	{Tag: b6.Tag{Key: "#water"}, MinZoom: 12, Layer: BasemapLayerWater},
	{Tag: b6.Tag{Key: "#waterway"}, MinZoom: 14, Layer: BasemapLayerWater},
	{Tag: b6.Tag{Key: "#place", Value: b6.String("city")}, MaxZoom: 11, Layer: BasemapLayerLabel},
}

type BasemapRenderer struct {
	RenderRules RenderRules
	Worlds      ingest.Worlds
}

func (b *BasemapRenderer) Render(tile b6.Tile, args *TileArgs) (*Tile, error) {
	features := b.findFeatures(args.R, tile)
	layers := NewLayers()
	fs := make([]*Feature, 0, 2)
	for _, feature := range features {
		fs = b.renderFeature(args.R, feature, layers, fs[0:0])
	}
	return &Tile{Layers: (*layers)[0:]}, nil
}

func (b *BasemapRenderer) findFeatures(root b6.FeatureID, tile b6.Tile) []b6.Feature { // TODO: rename root to smth more sensible / like world
	bounds := tile.RectBound()
	regionQuery := b6.MightIntersect{Region: bounds}
	q := b6.Intersection{b.RenderRules.ToQuery(tile.Z), regionQuery}
	features := b6.AllFeatures(b.Worlds.FindOrCreateWorld(root).FindFeatures(q))
	sort.Sort(byLayerThenID(features))
	return features
}

func (b *BasemapRenderer) renderFeature(root b6.FeatureID, f b6.Feature, layers *BasemapLayers, fs []*Feature) []*Feature {
	var tags [1]b6.Tag
	for _, rule := range b.RenderRules {
		if v, ok := rule.Matches(f); ok {
			tags[0] = b6.Tag{Key: rule.Tag.Key[1:], Value: v}
			fs = FillFeaturesFromFeature(f, tags[0:], fs, &rule, b.Worlds.FindOrCreateWorld(root))
			layers[rule.Layer].AddFeatures(fs)
			break
		}
	}
	return fs
}

func FillFeaturesFromFeature(f b6.Feature, tags []b6.Tag, tfs []*Feature, rule *RenderRule, w ingest.MutableWorld) []*Feature {
	if f, ok := f.(b6.PhysicalFeature); ok {
		switch f.GeometryType() {
		case b6.GeometryTypePoint:
			tfs = fillFeaturesFromPoint(f, tags, tfs, rule)
		case b6.GeometryTypePath:
			tfs = fillFeaturesFromPath(f, tags, tfs, rule)
		case b6.GeometryTypeArea:
			tfs = fillFeaturesFromArea(f.(b6.AreaFeature), tags, tfs, rule, w)
		}
	}
	return tfs
}

func fillFeaturesFromPoint(point b6.PhysicalFeature, tags []b6.Tag, fs []*Feature, rule *RenderRule) []*Feature {
	f := NewFeature(NewPoint(point.Point()))
	f.ID = api.TileFeatureID(point.FeatureID())
	for _, t := range tags {
		f.Tags[t.Key] = t.Value.String()
	}
	fillTagsFromTags(f, point, rule)
	fillTagsFromIcon(f, point, rule)
	return append(fs, f)
}

func fillFeaturesFromPath(path b6.PhysicalFeature, tags []b6.Tag, fs []*Feature, rule *RenderRule) []*Feature {
	f := NewFeature(NewLineString(path.Polyline()))
	f.ID = api.TileFeatureID(path.FeatureID())
	for _, t := range tags {
		f.Tags[t.Key] = t.Value.String()
	}
	fillTagsFromTags(f, path, rule)
	return append(fs, f)
}

func fillFeaturesFromArea(area b6.AreaFeature, tags []b6.Tag, fs []*Feature, rule *RenderRule, w ingest.MutableWorld) []*Feature {
	if highway := area.Get("#highway"); highway.IsValid() {
		if a := area.Get("area"); !a.IsValid() || a.Value.String() == "no" {
			for i := 0; i < area.Len(); i++ {
				for _, path := range area.Feature(i) {
					f := NewFeature(NewLineString(path.Polyline()))
					f.ID = api.TileFeatureID(path.FeatureID())
					fillTagsFromTags(f, area, rule)
					fs = append(fs, f)
				}
			}
			return fs
		}
	}

	for i := 0; i < area.Len(); i++ {
		f := NewFeature(NewPolygon(area.Polygon(i)))
		f.ID = api.TileFeatureIDForPolygon(area.FeatureID(), i)
		for _, t := range tags {
			f.Tags[t.Key] = t.Value.String()
		}
		fillTagsFromTags(f, area, rule)
		fs = append(fs, f)
	}
	if rule.Icon {
		if point, ok := findIconPoint(area, w); ok {
			f := NewFeature(NewPoint(point))
			f.ID = api.TileFeatureID(area.FeatureID())
			fillTagsFromIcon(f, area, rule)
			if _, ok := f.Tags["icon"]; ok {
				fs = append(fs, f)
			}
		}
	}
	return fs
}

func findIconPoint(area b6.AreaFeature, w ingest.MutableWorld) (s2.Point, bool) {
	for i := 0; i < area.Len(); i++ {
		paths := area.Feature(i)
		for _, path := range paths {
			for j, r := range path.References() {
				if point := w.FindFeatureByID(r.Source()); point != nil {
					if entrance := point.Get("entrance"); entrance.IsValid() {
						return path.PointAt(j), true
					}
				}
			}
		}
	}
	if area.Len() > 0 {
		if paths := area.Feature(0); len(paths) > 0 {
			if paths[0].GeometryLen() > 0 {
				return paths[0].PointAt(0), true
			}
		}
	}
	return s2.Point{}, false
}

func fillTagsFromTags(tf *Feature, f b6.Feature, rule *RenderRule) {
	if rule.Label {
		fillNameFromFeature(tf, f)
	}
	fillColourFromFeature(tf, f)
	fillIDFromFeature(tf, f)
}

func fillTagsFromIcon(tf *Feature, f b6.Feature, rule *RenderRule) {
	if rule.Icon {
		if icon, ok := IconForFeature(f); ok {
			tf.Tags["icon"] = icon
		}
	}
}

func fillNameFromFeature(tf *Feature, f b6.Feature) {
	if name := f.Get("addr:housename"); name.IsValid() {
		tf.Tags["name"] = name.Value.String()
	} else if name := f.Get("name"); name.IsValid() {
		tf.Tags["name"] = name.Value.String()
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
		if converted := colourFromTagValue(colour.Value.String()); converted != "" {
			tf.Tags["colour"] = converted
		}
	}
}
