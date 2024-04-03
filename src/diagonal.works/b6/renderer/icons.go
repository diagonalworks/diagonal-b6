package renderer

import (
	"diagonal.works/b6"
)

var iconsForTag = []struct {
	Tag  b6.Tag
	Icon string
}{
	{b6.Tag{Key: "#highway", Value: b6.String("bus_stop")}, "bus"},
	{b6.Tag{Key: "#gtfs", Value: b6.String("stop")}, "bus"},
	{b6.Tag{Key: "#railway", Value: b6.String("tram_stop")}, "rail-metro"},
	{b6.Tag{Key: "#railway", Value: b6.String("train_station")}, "rail"},
	// The most frequent amenity tag values from: https://taginfo.openstreetmap.org/keys/amenity#values
	{b6.Tag{Key: "#amenity", Value: b6.String("parking")}, "parking"},
	{b6.Tag{Key: "#amenity", Value: b6.String("parking_space")}, "parking"},
	{b6.Tag{Key: "#amenity", Value: b6.String("place_of_worship")}, "religious-christian"}, // TODO: breakdown
	{b6.Tag{Key: "#amenity", Value: b6.String("restaurant")}, "restaurant"},
	{b6.Tag{Key: "#amenity", Value: b6.String("school")}, "school"},
	{b6.Tag{Key: "#amenity", Value: b6.String("pub")}, "beer"},
	{b6.Tag{Key: "#amenity", Value: b6.String("doctors")}, "hospital"},
	// The most frequent landuse tag values from: https://taginfo.openstreetmap.org/keys/shop#values
	{b6.Tag{Key: "#shop", Value: b6.String("convenience")}, "grocery"},
	{b6.Tag{Key: "#shop", Value: b6.String("supermarket")}, "grocery"},
	{b6.Tag{Key: "#shop", Value: b6.String("clothes")}, "clothing-store"},
	{b6.Tag{Key: "#shop", Value: b6.String("hairdresser")}, "hairdresser"},
	{b6.Tag{Key: "#shop", Value: b6.String("car_repair")}, "car-repair"},
	{b6.Tag{Key: "#shop", Value: b6.String("bakery")}, "bakery"},
	{b6.Tag{Key: "#shop", Value: b6.String("alcohol")}, "alcohol-shop"},
	{b6.Tag{Key: "#shop", Value: b6.String("yes")}, "shop"},
	{b6.Tag{Key: "#shop", Value: b6.String("")}, "shop"},
	// The most frequent building tag values from: https://taginfo.openstreetmap.org/keys/building#values
	{b6.Tag{Key: "#building", Value: b6.String("house")}, "home"},
	{b6.Tag{Key: "#building", Value: b6.String("detached")}, "home"},
	{b6.Tag{Key: "#building", Value: b6.String("semidetached_house")}, "home"},
	{b6.Tag{Key: "#building", Value: b6.String("residential")}, "home"},
	{b6.Tag{Key: "#building", Value: b6.String("garage")}, "car"},
	{b6.Tag{Key: "#building", Value: b6.String("church")}, "religious-christian"}, // TODO: breakdown
	{b6.Tag{Key: "#building", Value: b6.String("school")}, "school"},
	{b6.Tag{Key: "#building", Value: b6.String("apartments")}, "home"},
	{b6.Tag{Key: "#building", Value: b6.String("commercial")}, "shop"},
	{b6.Tag{Key: "#building", Value: b6.String("retail")}, "shop"},
	{b6.Tag{Key: "#building", Value: b6.String("farm")}, "farm"},
	{b6.Tag{Key: "#building", Value: b6.String("industrial")}, "industry"},
	{b6.Tag{Key: "#building", Value: b6.String("garage")}, "car"},
	{b6.Tag{Key: "#building", Value: b6.String("garages")}, "car"},
	{b6.Tag{Key: "#building", Value: b6.String("hotel")}, "loding"},
	{b6.Tag{Key: "#building", Value: b6.String("stable")}, "horse-riding"},
	{b6.Tag{Key: "#building", Value: b6.String("train_station")}, "rail"},
	{b6.Tag{Key: "#building", Value: b6.String("yes")}, "building"},
	{b6.Tag{Key: "#building", Value: b6.String("other")}, "building"}, // Only for analysis results
	{b6.Tag{Key: "#building"}, "building"},
	// The most frequent landuse tag values from: https://taginfo.openstreetmap.org/keys/landuse#values
	{b6.Tag{Key: "#landuse", Value: b6.String("farmland")}, "farm"},
	{b6.Tag{Key: "#landuse", Value: b6.String("forest")}, "park"},  // As the icon is a tree
	{b6.Tag{Key: "#landuse", Value: b6.String("orchard")}, "park"}, // As the icon is a tree
	{b6.Tag{Key: "#landuse", Value: b6.String("industrial")}, "industry"},
	{b6.Tag{Key: "#landuse", Value: b6.String("vineyard")}, "alcohol-shop"}, // As the icon is a wine bottle & glass
	// The most frequent leisure tag values from: https://taginfo.openstreetmap.org/keys/landuse#values
	{b6.Tag{Key: "#leisure", Value: b6.String("pitch")}, "soccer"},
	{b6.Tag{Key: "#leisure", Value: b6.String("swimming_pool")}, "swimming"},
	{b6.Tag{Key: "#leisure", Value: b6.String("park")}, "park"},
	{b6.Tag{Key: "#leisure", Value: b6.String("garden")}, "garden"},
	{b6.Tag{Key: "#leisure", Value: b6.String("playground")}, "playground"},
	{b6.Tag{Key: "#leisure", Value: b6.String("sports_centre")}, "soccer"},
	{b6.Tag{Key: "#leisure", Value: b6.String("nature_reserve")}, "park"},
	{b6.Tag{Key: "#leisure", Value: b6.String("allotments")}, "garden"},
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
