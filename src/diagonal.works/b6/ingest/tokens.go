package ingest

import (
	"fmt"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/search"
	"github.com/golang/geo/s2"
)

const MaxIndexedCellLevel = 16

const S2CellIDTokenPrefix = "s2:"
const S2AncestorCellIDTokenPrefix = "a2:"
const TileTokenPrefix = "t:"

func CellIDAncestorTokens(covering s2.CellUnion) []string {
	tokens := make([]string, 0, len(covering)*2)
	cells := make(map[s2.CellID]struct{})
	for _, cell := range covering {
		cells[cell] = struct{}{}
	}
	for len(cells) > 0 {
		parents := make(map[s2.CellID]struct{})
		for id := range cells {
			if id.Level() != 0 {
				parents[id.Parent(id.Level()-1)] = struct{}{}
			}
		}
		for id := range parents {
			tokens = append(tokens, AncestorCellIDToToken(id))
		}
		cells = parents
	}
	return tokens
}

func CellIDToToken(cell s2.CellID) string {
	return fmt.Sprintf("%s%s", S2CellIDTokenPrefix, cell.ToToken())
}

func TokenToCellID(token string) (s2.CellID, bool) {
	if strings.HasPrefix(token, S2CellIDTokenPrefix) {
		return s2.CellIDFromToken(token[len(S2CellIDTokenPrefix):]), true
	}
	return s2.CellID(0), false
}

func AncestorCellIDToToken(cell s2.CellID) string {
	return fmt.Sprintf("%s%s", S2AncestorCellIDTokenPrefix, cell.ToToken())
}

func TokensForFeature(feature b6.PhysicalFeature) []string {
	if feature.FeatureID().Type == b6.FeatureTypePoint && len(feature.AllTags()) == 0 {
		return []string{}
	}

	tokens := make([]string, 0, 64) // Best guess
	tokens = append(tokens, search.AllToken)

	cells := feature.Covering(s2.RegionCoverer{MaxLevel: MaxIndexedCellLevel, MaxCells: 5})
	for i, cell := range cells {
		if cells[i].Level() == 0 {
			continue
		}
		tokens = append(tokens, CellIDToToken(cell))
	}
	tokens = append(tokens, CellIDAncestorTokens(cells)...)

	for _, tag := range feature.AllTags() {
		if token, ok := TokenForTag(tag); ok {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func TokenForTag(tag b6.Tag) (string, bool) {
	if strings.HasPrefix(tag.Key, "#") {
		return fmt.Sprintf("%s=%s", tag.Key[1:], tag.Value), true
	} else if strings.HasPrefix(tag.Key, "@") {
		return tag.Key[1:], true
	}
	return "", false
}

func QueryForKeyValue(key string, value string) (search.Query, error) {
	if strings.HasPrefix(key, "#") {
		return search.All{Token: fmt.Sprintf("%s=%s", key[1:], value)}, nil
	}
	return search.Empty{}, fmt.Errorf("Can't search for values for tag %q: not indexed", key)
}

func QueryForAllValues(key string) (search.Query, error) {
	if strings.HasPrefix(key, "#") {
		return search.TokenPrefix{Prefix: fmt.Sprintf("%s=", key[1:])}, nil
	} else if strings.HasPrefix(key, "@") {
		return search.All{Token: key[1:]}, nil
	}
	return search.Empty{}, fmt.Errorf("Can't search for all values for tag %q: not indexed", key)
}
