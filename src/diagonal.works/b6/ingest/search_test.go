package ingest

import (
	"testing"

	"diagonal.works/b6"
	"github.com/golang/geo/s2"
)

func TestMatches(t *testing.T) {
	id := b6.FeatureID{b6.FeatureTypePoint, "diagonal.works/test", 0}
	f := &GenericFeature{ID: id, Tags: []b6.Tag{{Key: b6.PointTag, Value: b6.LatLng(s2.LatLngFromDegrees(51.5366567, -0.1263944))}}}
	f.AddTag(b6.Tag{Key: "name", Value: b6.StringExpression("Vermuteria")})
	f.AddTag(b6.Tag{Key: "#amenity", Value: b6.StringExpression("cafe")})

	cases := []struct {
		q        b6.Query
		expected bool
	}{
		{b6.Keyed{Key: "#amenity"}, true},
		{b6.Tagged{Key: "#amenity", Value: b6.StringExpression("cafe")}, true},
		{b6.Tagged{Key: "#amenity", Value: b6.StringExpression("restaurant")}, false},
		{b6.Union{b6.Tagged{Key: "#amenity", Value: b6.StringExpression("cafe")}}, true},
		{b6.Union{b6.Tagged{Key: "#amenity", Value: b6.StringExpression("restaurant")}}, false},
		{b6.Intersection{b6.Tagged{Key: "#amenity", Value: b6.StringExpression("cafe")}}, true},
		{b6.Intersection{b6.Tagged{Key: "#amenity", Value: b6.StringExpression("restaurant")}}, false},
		{b6.Union{b6.Tagged{Key: "#amenity", Value: b6.StringExpression("cafe")}, b6.Tagged{Key: "#amenity", Value: b6.StringExpression("restaurant")}}, true},
		{b6.Intersection{b6.Tagged{Key: "#amenity", Value: b6.StringExpression("cafe")}, b6.Tagged{Key: "#amenity", Value: b6.StringExpression("restaurant")}}, false},
	}

	for _, c := range cases {
		if matches := c.q.Matches(WrapFeature(f, nil), nil); matches != c.expected {
			t.Errorf("Unexpected matching for %s: expected %v, found %v", c.q, c.expected, matches)
		}
	}
}
