package ingest

import (
	"testing"

	"diagonal.works/b6"
	"github.com/golang/geo/s2"
)

func TestMatches(t *testing.T) {
	id := b6.MakePointID("diagonal.works/test", 0)
	p := NewPointFeature(id, s2.LatLngFromDegrees(51.5366567, -0.1263944))
	p.Tags.AddTag(b6.Tag{Key: "name", Value: "Vermuteria"})
	p.Tags.AddTag(b6.Tag{Key: "#amenity", Value: "cafe"})

	cases := []struct {
		q        b6.Query
		expected bool
	}{
		{b6.Keyed{Key: "#amenity"}, true},
		{b6.Tagged{Key: "#amenity", Value: "cafe"}, true},
		{b6.Tagged{Key: "#amenity", Value: "restaurant"}, false},
		{b6.Union{b6.Tagged{Key: "#amenity", Value: "cafe"}}, true},
		{b6.Union{b6.Tagged{Key: "#amenity", Value: "restaurant"}}, false},
		{b6.Intersection{b6.Tagged{Key: "#amenity", Value: "cafe"}}, true},
		{b6.Intersection{b6.Tagged{Key: "#amenity", Value: "restaurant"}}, false},
		{b6.Union{b6.Tagged{Key: "#amenity", Value: "cafe"}, b6.Tagged{Key: "#amenity", Value: "restaurant"}}, true},
		{b6.Intersection{b6.Tagged{Key: "#amenity", Value: "cafe"}, b6.Tagged{Key: "#amenity", Value: "restaurant"}}, false},
	}

	for _, c := range cases {
		if matches := c.q.Matches(WrapFeature(p, nil), nil); matches != c.expected {
			t.Errorf("Unexpected matching for %s: expected %v, found %v", c.q, c.expected, matches)
		}
	}
}
