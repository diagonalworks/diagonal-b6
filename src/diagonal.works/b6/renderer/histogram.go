package renderer

import (
	"fmt"
	"strconv"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
)

type HistogramRenderer struct {
	rules  RenderRules
	worlds ingest.Worlds
}

func NewHistogramRenderer(rules RenderRules, worlds ingest.Worlds) *HistogramRenderer {
	return &HistogramRenderer{
		rules:  rules,
		worlds: worlds,
	}
}

func (r *HistogramRenderer) Render(tile b6.Tile, args *TileArgs) (*Tile, error) {
	w := r.worlds.FindOrCreateWorld(args.R)

	id := b6.FeatureIDFromString(args.Q)
	if id == b6.FeatureIDInvalid || id.Type != b6.FeatureTypeCollection {
		return nil, fmt.Errorf("expected a collection ID for arg q, found %s", args.Q)
	}
	histogram := b6.FindCollectionByID(id.ToCollectionID(), w)

	var findBucket func(id b6.FeatureID, values []any) (int, bool)
	if histogram != nil {
		// TODO: Use types when collections have types
		if b6.CanAdaptCollection[b6.FeatureID, b6.FeatureID](histogram) {
			findBucket = func(id b6.FeatureID, values []any) (int, bool) {
				bucket := 0
				present := false
				for _, v := range histogram.FindValues(id, values[0:0]) {
					present = true
					if id, ok := v.(b6.FeatureID); ok && id != b6.FeatureIDInvalid {
						bucket++
					}
				}
				return bucket, present
			}
		} else if b6.CanAdaptCollection[b6.FeatureID, int](histogram) {
			findBucket = func(id b6.FeatureID, _ []any) (int, bool) {
				var bucket int
				var ok bool
				var v any
				if v, ok = histogram.FindValue(id); ok {
					bucket, ok = v.(int)
				}
				return bucket, ok
			}
		}
	}
	if findBucket == nil {
		return &Tile{
			Layers: []*Layer{{Name: "histogram"}},
		}, nil
	}

	features := w.FindFeatures(b6.MightIntersect{Region: tile.RectBound()})
	rendered := make([]*Feature, 0, 4)
	tags := make([]b6.Tag, 0, 4)
	values := make([]any, 0, 4)
	for features.Next() {
		if value, ok := findBucket(features.FeatureID(), values); ok {
			f := features.Feature()
			tags = tags[0:0]
			for _, rule := range r.rules {
				if t := f.Get(rule.Tag.Key); t.IsValid() && t.Value == rule.Tag.Value {
					tags = append(tags, b6.Tag{Key: rule.Tag.Key[1:], Value: t.Value})
					break
				}
			}
			tags = append(tags, b6.Tag{Key: "bucket", Value: b6.String(strconv.Itoa(value))})
			rendered = FillFeaturesFromFeature(features.Feature(), tags, rendered, &RenderRule{Label: true}, w)
		}
	}
	return &Tile{
		Layers: []*Layer{{Name: "histogram", Features: rendered}},
	}, nil
}
