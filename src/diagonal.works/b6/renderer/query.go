package renderer

import (
	"context"
	"fmt"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
)

type QueryRenderer struct {
	rules RenderRules
	world b6.World
	cores int
	fs    api.FunctionSymbols
	a     api.Adaptors
}

// QueryRenderRules is used to fill in an tile feature attributre
// describing the type of the feature, similar to BasemapRenderer
var QueryRenderRules = RenderRules{
	{Tag: b6.Tag{Key: "#amenity"}},
	{Tag: b6.Tag{Key: "#boundary"}},
	{Tag: b6.Tag{Key: "#highway"}},
	{Tag: b6.Tag{Key: "#landuse"}},
	{Tag: b6.Tag{Key: "#natural"}},
	{Tag: b6.Tag{Key: "#place"}},
	{Tag: b6.Tag{Key: "#railway"}},
	{Tag: b6.Tag{Key: "#water"}},
	{Tag: b6.Tag{Key: "#waterway"}},
}

const QueryRendererMaxFeaturesPerTile = 10000

func NewQueryRenderer(w b6.World, cores int) *QueryRenderer {
	return &QueryRenderer{
		rules: QueryRenderRules,
		world: w,
		cores: cores,
		fs:    functions.Functions(),
		a:     functions.Adaptors(),
	}
}

func (r *QueryRenderer) Render(tile b6.Tile, args *TileArgs) (*Tile, error) {
	context := api.Context{
		World:           r.world,
		FunctionSymbols: r.fs,
		Adaptors:        r.a,
		Cores:           r.cores,
		Context:         context.Background(),
	}
	v, err := api.EvaluateString(args.Q, &context)
	if err != nil {
		return nil, err
	}
	q, ok := v.(b6.Query)
	if !ok {
		return nil, fmt.Errorf("Expected a Query, found %T", v)
	}

	var fv func(interface{}) (interface{}, error)
	if args.V != "" {
		v, err := api.EvaluateString(args.V, &context)
		if err != nil {
			return nil, err
		}
		fv = v.(func(interface{}) (interface{}, error))
	}

	bounds := tile.RectBound()
	features := r.world.FindFeatures(b6.Intersection{q, b6.MightIntersect{Region: bounds}})
	rendered := make([]*Feature, 0, 4)
	n := 0
	tags := make([]b6.Tag, 0, 4)
	for features.Next() {
		f := features.Feature()
		tags = tags[0:0]
		for _, rule := range r.rules {
			if t := f.Get(rule.Tag.Key); t.IsValid() && (rule.Tag.Value == "" || t.Value == rule.Tag.Value) {
				tags = append(tags, b6.Tag{Key: rule.Tag.Key[1:], Value: t.Value})
				break
			}
		}
		if fv != nil {
			if v, err := fv(f); err == nil {
				switch v := v.(type) {
				case int:
					tags = append(tags, b6.Tag{Key: "v", Value: fmt.Sprintf("%d", v)})
				case string:
					tags = append(tags, b6.Tag{Key: "v", Value: v})
				case fmt.Stringer:
					tags = append(tags, b6.Tag{Key: "v", Value: v.String()})
				}
				rendered = FillFeaturesFromFeature(features.Feature(), tags, rendered)
			} else {
				rendered = FillFeaturesFromFeature(features.Feature(), tags, rendered)
			}
		} else {
			rendered = FillFeaturesFromFeature(features.Feature(), tags, rendered)
		}
		n++
		if n > QueryRendererMaxFeaturesPerTile {
			break
		}
	}
	return &Tile{
		Layers: []*Layer{{Name: "query", Features: rendered}},
	}, nil
}
