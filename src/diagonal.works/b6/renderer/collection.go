package renderer

import (
	"fmt"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
)

type CollectionRenderer struct {
	rules  RenderRules
	worlds ingest.Worlds
}

func NewCollectionRenderer(rules RenderRules, worlds ingest.Worlds) *CollectionRenderer {
	return &CollectionRenderer{
		rules:  rules,
		worlds: worlds,
	}
}

func (r *CollectionRenderer) Render(tile b6.Tile, args *TileArgs) (*Tile, error) {
	w := r.worlds.FindOrCreateWorld(args.R)

	id := b6.FeatureIDFromString(args.Q)
	if id == b6.FeatureIDInvalid || id.Type != b6.FeatureTypeCollection {
		return nil, fmt.Errorf("expected a collection ID for arg q, found %s", args.Q)
	}
	collection := b6.FindCollectionByID(id.ToCollectionID(), w)
	rendered := make([]*Feature, 0, 4)
	if collection != nil {
		bounds := b6.NewIntersectsCap(tile.CapBound())
		features := b6.AdaptCollection[any, b6.Identifiable](collection)
		tags := make([]b6.Tag, 0, 4)
		i := features.Begin()
		for {
			ok, err := i.Next()
			if err != nil {
				return nil, err
			} else if !ok {
				break
			}
			if f := w.FindFeatureByID(i.Value().FeatureID()); f != nil {
				if bounds.Matches(f, w) {
					tags = r.rules.AddTags(f, tags[0:0])
					rendered = FillFeaturesFromFeature(f, tags, rendered, &RenderRule{Label: true}, w)
				}
			}
		}
	} else {
		return nil, fmt.Errorf("no feature with ID %s", id)
	}

	return &Tile{
		Layers: []*Layer{{Name: "collection", Features: rendered}},
	}, nil
}
