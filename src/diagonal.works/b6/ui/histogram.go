package ui

import (
	"diagonal.works/b6"
	"diagonal.works/b6/api"
	pb "diagonal.works/b6/proto"
)

func fillResponseFromHistogramFeature(response *UIResponseJSON, c b6.CollectionFeature, w b6.World) error {
	p := (*pb.UIResponseProto)(response.Proto)
	labels := api.HistogramBucketLabels(c)
	counts := api.HistogramBucketCounts(c)

	if len(labels) != len(counts) {
		p.Stack.Substacks = fillSubstacksFromFeature(p.Stack.Substacks, c, w)
		highlightInResponse(p, c.FeatureID())
		return nil
	}

	total := 0
	for _, count := range counts {
		total += count
	}

	substack := &pb.SubstackProto{}
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
	p.Stack.Substacks = append(p.Stack.Substacks, substack)

	p.Layers = append(p.Layers, &pb.MapLayerProto{
		Path:   "histogram",
		Q:      c.FeatureID().String(),
		Before: pb.MapLayerPosition_MapLayerPositionEnd,
	})
	return nil
}
