package ingest

import (
	"testing"

	"github.com/golang/geo/s2"
)

func TestAncestorCellIDTokens(t *testing.T) {
	granarySquare := s2.CellIDFromToken("48761b3dc")
	kingsCross := s2.CellIDFromToken("48761b3c4")
	union := s2.CellUnion{granarySquare, kingsCross}

	ancestorTokens := CellIDAncestorTokens(union)
	expectedLength := 15
	if len(ancestorTokens) != expectedLength {
		t.Errorf("Expected %d tokens, found %d", expectedLength, len(ancestorTokens))
	}

	found := false
	expectedToken := "a2:484"
	for _, token := range ancestorTokens {
		if token == expectedToken {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find token %q", expectedToken)
	}
}
