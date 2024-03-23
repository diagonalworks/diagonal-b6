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
	if histogram == nil || !b6.CanAdaptCollection[b6.FeatureID, int](histogram) {
		return &Tile{
			Layers: []*Layer{{Name: "histogram"}},
		}, nil
	}

	features := w.FindFeatures(b6.MightIntersect{Region: tile.RectBound()})
	rendered := make([]*Feature, 0, 4)
	tags := make([]b6.Tag, 0, 4)
	for features.Next() {
		if value, ok := histogram.FindValue(features.FeatureID()); ok {
			if bucket, ok := value.(int); ok {
				f := features.Feature()
				tags = tags[0:0]
				for _, rule := range r.rules {
					if t := f.Get(rule.Tag.Key); t.IsValid() && t.Value == rule.Tag.Value {
						tags = append(tags, b6.Tag{Key: rule.Tag.Key[1:], Value: t.Value})
						break
					}
				}
				tags = append(tags, b6.Tag{Key: "bucket", Value: b6.String(strconv.Itoa(bucket))})
				rendered = FillFeaturesFromFeature(features.Feature(), tags, rendered, &RenderRule{Label: true})
			}
		}
	}
	return &Tile{
		Layers: []*Layer{{Name: "histogram", Features: rendered}},
	}, nil
}
