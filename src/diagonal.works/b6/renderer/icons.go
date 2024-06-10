package renderer

import (
	"diagonal.works/b6"
)

var iconsForTag = []struct {
	Tag  b6.Tag
	Icon string
}{
	{b6.Tag{Key: "#highway", Value: b6.StringExpression("bus_stop")}, "bus"},
	{b6.Tag{Key: "#gtfs", Value: b6.StringExpression("stop")}, "bus"},
	{b6.Tag{Key: "#railway", Value: b6.StringExpression("tram_stop")}, "rail-metro"},
	{b6.Tag{Key: "#railway", Value: b6.StringExpression("train_station")}, "rail"},
	// The most frequent amenity tag values from: https://taginfo.openstreetmap.org/keys/amenity#values
	{b6.Tag{Key: "#amenity", Value: b6.StringExpression("parking")}, "parking"},
	{b6.Tag{Key: "#amenity", Value: b6.StringExpression("parking_space")}, "parking"},
	{b6.Tag{Key: "#amenity", Value: b6.StringExpression("place_of_worship")}, "religious-christian"}, // TODO: breakdown
	{b6.Tag{Key: "#amenity", Value: b6.StringExpression("restaurant")}, "restaurant"},
	{b6.Tag{Key: "#amenity", Value: b6.StringExpression("school")}, "school"},
	{b6.Tag{Key: "#amenity", Value: b6.StringExpression("pub")}, "beer"},
	{b6.Tag{Key: "#amenity", Value: b6.StringExpression("doctors")}, "hospital"},
	// The most frequent landuse tag values from: https://taginfo.openstreetmap.org/keys/shop#values
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("convenience")}, "grocery"},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("supermarket")}, "grocery"},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("clothes")}, "clothing-store"},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("hairdresser")}, "hairdresser"},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("car_repair")}, "car-repair"},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("bakery")}, "bakery"},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("alcohol")}, "alcohol-shop"},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("yes")}, "shop"},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("")}, "shop"},
	// The most frequent building tag values from: https://taginfo.openstreetmap.org/keys/building#values
	{b6.Tag{Key: "#building", Value: b6.StringExpression("house")}, "home"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("detached")}, "home"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("semidetached_house")}, "home"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("residential")}, "home"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("garage")}, "car"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("church")}, "religious-christian"}, // TODO: breakdown
	{b6.Tag{Key: "#building", Value: b6.StringExpression("school")}, "school"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("apartments")}, "home"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("commercial")}, "shop"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("retail")}, "shop"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("farm")}, "farm"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("industrial")}, "industry"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("garage")}, "car"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("garages")}, "car"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("hotel")}, "loding"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("stable")}, "horse-riding"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("train_station")}, "rail"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("yes")}, "building"},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("other")}, "building"}, // Only for analysis results
	{b6.Tag{Key: "#building"}, "building"},
	// The most frequent landuse tag values from: https://taginfo.openstreetmap.org/keys/landuse#values
	{b6.Tag{Key: "#landuse", Value: b6.StringExpression("farmland")}, "farm"},
	{b6.Tag{Key: "#landuse", Value: b6.StringExpression("forest")}, "park"},  // As the icon is a tree
	{b6.Tag{Key: "#landuse", Value: b6.StringExpression("orchard")}, "park"}, // As the icon is a tree
	{b6.Tag{Key: "#landuse", Value: b6.StringExpression("industrial")}, "industry"},
	{b6.Tag{Key: "#landuse", Value: b6.StringExpression("vineyard")}, "alcohol-shop"}, // As the icon is a wine bottle & glass
	// The most frequent leisure tag values from: https://taginfo.openstreetmap.org/keys/landuse#values
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("pitch")}, "soccer"},
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("swimming_pool")}, "swimming"},
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("park")}, "park"},
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("garden")}, "garden"},
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("playground")}, "playground"},
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("sports_centre")}, "soccer"},
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("nature_reserve")}, "park"},
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("allotments")}, "garden"},
}

func IconForTag(t b6.Tag) (string, bool) {
	for _, i := range iconsForTag {
		if i.Tag.Key == t.Key && (i.Tag.Value == nil || i.Tag.Value.String() == "" || i.Tag.Value.String() == t.Value.String()) {
			return i.Icon, true
		}
	}
	return "dot", false
}

func IconForFeature(f b6.Feature) (string, bool) {
	for _, i := range iconsForTag {
		if t := f.Get(i.Tag.Key); t.IsValid() && (i.Tag.Value == nil || i.Tag.Value.String() == "" || i.Tag.Value.String() == t.Value.String()) {
			return i.Icon, true
		}
	}
	return "dot", false
}
