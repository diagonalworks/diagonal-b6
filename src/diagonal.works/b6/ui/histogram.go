package ui

import (
	"diagonal.works/b6"
	"diagonal.works/b6/api"
	pb "diagonal.works/b6/proto"
)

func addLabel(response *UIResponseJSON, f b6.Feature) {
	if label := f.Get("b6:label"); label.IsValid() {
		substack := pb.SubstackProto{
			Lines: []*pb.LineProto{
				{
					Line: &pb.LineProto_Header{
						Header: &pb.HeaderLineProto{
							Title: &pb.AtomProto{
								Atom: &pb.AtomProto_Value{
									Value: label.Value.String(),
								},
							},
						},
					},
				},
			},
		}
		response.Proto.Stack.Substacks = append(response.Proto.Stack.Substacks, &substack)
	}
}

func fillResponseFromHistogramFeature(response *UIResponseJSON, c b6.CollectionFeature, w b6.World) error {
	p := (*pb.UIResponseProto)(response.Proto)
	counts, total, err := api.HistogramBucketCounts(c)
	labels := api.HistogramBucketLabels(c, len(counts))
	if err != nil {
		return err
	}

	addLabel(response, c)

	substack := &pb.SubstackProto{}
	if h := c.Get("b6:histogram"); h.Value.String() == "swatch" {
		for i, label := range labels {
			substack.Lines = append(substack.Lines, &pb.LineProto{
				Line: &pb.LineProto_Swatch{
					Swatch: &pb.SwatchLineProto{
						Label: AtomFromValue(label, w),
						Index: int32(i),
					},
				},
			})
		}
	} else {
		for i, label := range labels {
			substack.Lines = append(substack.Lines, &pb.LineProto{
				Line: &pb.LineProto_HistogramBar{
					HistogramBar: &pb.HistogramBarLineProto{
						Range: AtomFromValue(label, w),
						Value: int32(counts[i]),
						Index: int32(i),
						Total: int32(total),
					},
				},
			})
		}
	}
	p.Stack.Substacks = append(p.Stack.Substacks, substack)

	p.Layers = append(p.Layers, &pb.MapLayerProto{
		Path:   "histogram",
		Q:      c.FeatureID().String(),
		Before: pb.MapLayerPosition_MapLayerPositionEnd,
	})

	if b6.CanAdaptCollection[any, b6.FeatureID](c) {
		ids := b6.AdaptCollection[any, b6.FeatureID](c)
		err = fillResponseFromHistogramFeatures(response, ids, w)
	}

	return err
}

func fillResponseFromHistogramFeatures(response *UIResponseJSON, c b6.Collection[any, b6.FeatureID], w b6.World) error {
	i := c.Begin()
	values := make(map[b6.FeatureID]struct{})
	for {
		ok, err := i.Next()
		if err != nil {
			return err
		} else if !ok {
			break
		}
		values[i.Value()] = struct{}{}
	}
	fillResponseFromDestinations(response, values, w)
	return nil
}
