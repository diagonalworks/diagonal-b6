package renderer

import (
	"context"
	"fmt"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
)

type QueryRenderer struct {
	world b6.World
	cores int
	fs    api.FunctionSymbols
	fw    api.FunctionWrappers
}

const QueryRendererMaxFeaturesPerTile = 1000

func NewQueryRenderer(w b6.World, cores int) *QueryRenderer {
	return &QueryRenderer{
		world: w,
		cores: cores,
		fs:    functions.Functions(),
		fw:    functions.Wrappers(),
	}
}

func (r *QueryRenderer) Render(tile b6.Tile, args *TileArgs) (*Tile, error) {
	context := api.Context{
		World:            r.world,
		FunctionSymbols:  r.fs,
		FunctionWrappers: r.fw,
		Cores:            r.cores,
		Context:          context.Background(),
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
	var tags [1]b6.Tag
	tags[0].Key = "v"
	n := 0
	for features.Next() {
		if fv != nil {
			if v, err := fv(features.Feature()); err == nil {
				switch v := v.(type) {
				case int:
					tags[0].Value = fmt.Sprintf("%d", v)
				case string:
					tags[0].Value = v
				case fmt.Stringer:
					tags[0].Value = v.String()
				default:
					tags[0].Value = ""
				}
				rendered = FillFeaturesFromFeature(features.Feature(), tags[0:], rendered)
			} else {
				rendered = FillFeaturesFromFeature(features.Feature(), tags[0:0], rendered)
			}
		} else {
			rendered = FillFeaturesFromFeature(features.Feature(), tags[0:0], rendered)
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
