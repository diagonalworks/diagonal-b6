package ingest

import (
	"diagonal.works/b6"
	"diagonal.works/b6/search"
	"github.com/golang/geo/s2"
)

func TokensForFeature(feature b6.Feature) []string {
	if feature.FeatureID().Type == b6.FeatureTypePoint && len(feature.AllTags()) == 0 {
		return []string{}
	}

	tokens := make([]string, 0, 64) // Best guess
	tokens = append(tokens, search.AllToken)

	if p, ok := feature.(b6.PhysicalFeature); ok {
		covering := p.Covering(s2.RegionCoverer{MaxLevel: search.MaxIndexedCellLevel, MaxCells: 5})
		tokens = search.TokensForCovering(covering, tokens)
	}

	for _, tag := range feature.AllTags() {
		if token, ok := b6.TokenForTag(tag); ok {
			tokens = append(tokens, token)
		}
	}
	return tokens
}
