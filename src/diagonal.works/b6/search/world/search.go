package world

import (
	"diagonal.works/b6"
	"diagonal.works/b6/search"
)

type Empty struct {
	search.Empty
}

func (_ Empty) Matches(b6.Feature) bool {
	return false
}
