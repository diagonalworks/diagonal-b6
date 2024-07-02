package renderer

import (
	"diagonal.works/b6"
)

var iconsForTag = []struct {
	Tag  b6.Tag
	Icon string
}{
	{b6.Tag{Key: "#highway", Value: b6.NewStringExpression("bus_stop")}, "bus"},
	{b6.Tag{Key: "#gtfs", Value: b6.NewStringExpression("stop")}, "bus"},
	{b6.Tag{Key: "#railway", Value: b6.NewStringExpression("tram_stop")}, "rail-metro"},
	{b6.Tag{Key: "#railway", Value: b6.NewStringExpression("train_station")}, "rail"},
	// The most frequent amenity tag values from: https://taginfo.openstreetmap.org/keys/amenity#values
	{b6.Tag{Key: "#amenity", Value: b6.NewStringExpression("parking")}, "parking"},
	{b6.Tag{Key: "#amenity", Value: b6.NewStringExpression("parking_space")}, "parking"},
	{b6.Tag{Key: "#amenity", Value: b6.NewStringExpression("place_of_worship")}, "religious-christian"}, // TODO: breakdown
	{b6.Tag{Key: "#amenity", Value: b6.NewStringExpression("restaurant")}, "restaurant"},
	{b6.Tag{Key: "#amenity", Value: b6.NewStringExpression("school")}, "school"},
	{b6.Tag{Key: "#amenity", Value: b6.NewStringExpression("pub")}, "beer"},
	{b6.Tag{Key: "#amenity", Value: b6.NewStringExpression("doctors")}, "hospital"},
	// The most frequent landuse tag values from: https://taginfo.openstreetmap.org/keys/shop#values
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("convenience")}, "grocery"},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("supermarket")}, "grocery"},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("clothes")}, "clothing-store"},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("hairdresser")}, "hairdresser"},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("car_repair")}, "car-repair"},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("bakery")}, "bakery"},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("alcohol")}, "alcohol-shop"},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("yes")}, "shop"},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("")}, "shop"},
	// The most frequent building tag values from: https://taginfo.openstreetmap.org/keys/building#values
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("house")}, "home"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("detached")}, "home"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("semidetached_house")}, "home"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("residential")}, "home"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("garage")}, "car"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("church")}, "religious-christian"}, // TODO: breakdown
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("school")}, "school"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("apartments")}, "home"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("commercial")}, "shop"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("retail")}, "shop"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("farm")}, "farm"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("industrial")}, "industry"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("garage")}, "car"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("garages")}, "car"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("hotel")}, "loding"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("stable")}, "horse-riding"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("train_station")}, "rail"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("yes")}, "building"},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("other")}, "building"}, // Only for analysis results
	{b6.Tag{Key: "#building"}, "building"},
	// The most frequent landuse tag values from: https://taginfo.openstreetmap.org/keys/landuse#values
	{b6.Tag{Key: "#landuse", Value: b6.NewStringExpression("farmland")}, "farm"},
	{b6.Tag{Key: "#landuse", Value: b6.NewStringExpression("forest")}, "park"},  // As the icon is a tree
	{b6.Tag{Key: "#landuse", Value: b6.NewStringExpression("orchard")}, "park"}, // As the icon is a tree
	{b6.Tag{Key: "#landuse", Value: b6.NewStringExpression("industrial")}, "industry"},
	{b6.Tag{Key: "#landuse", Value: b6.NewStringExpression("vineyard")}, "alcohol-shop"}, // As the icon is a wine bottle & glass
	// The most frequent leisure tag values from: https://taginfo.openstreetmap.org/keys/landuse#values
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("pitch")}, "soccer"},
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("swimming_pool")}, "swimming"},
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("park")}, "park"},
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("garden")}, "garden"},
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("playground")}, "playground"},
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("sports_centre")}, "soccer"},
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("nature_reserve")}, "park"},
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("allotments")}, "garden"},
}

func IconForTag(t b6.Tag) (string, bool) {
	for _, i := range iconsForTag {
		if i.Tag.Key == t.Key && (i.Tag.Value.String() == "" || i.Tag.Value.String() == t.Value.String()) {
			return i.Icon, true
		}
	}
	return "dot", false
}

func IconForFeature(f b6.Feature) (string, bool) {
	for _, i := range iconsForTag {
		if t := f.Get(i.Tag.Key); t.IsValid() && (i.Tag.Value.String() == "" || i.Tag.Value.String() == t.Value.String()) {
			return i.Icon, true
		}
	}
	return "dot", false
}
