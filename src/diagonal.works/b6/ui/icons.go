package ui

import (
	"diagonal.works/b6"
)

var iconsForTag = []struct {
	Tag  b6.Tag
	Icon string
}{
	{b6.Tag{Key: "#highway", Value: "bus_stop"}, "bus"},
	{b6.Tag{Key: "#railway", Value: "tram_stop"}, "rail-metro"},
	{b6.Tag{Key: "#railway", Value: "train_station"}, "rail"},
	// The most frequent amenity tag values from: https://taginfo.openstreetmap.org/keys/amenity#values
	{b6.Tag{Key: "#amenity", Value: "parking"}, "parking"},
	{b6.Tag{Key: "#amenity", Value: "parking_space"}, "parking"},
	{b6.Tag{Key: "#amenity", Value: "place_of_worship"}, "religious-christian"}, // TODO: breakdown
	{b6.Tag{Key: "#amenity", Value: "restaurant"}, "restaurant"},
	{b6.Tag{Key: "#amenity", Value: "school"}, "school"},
	{b6.Tag{Key: "#amenity", Value: "pub"}, "beer"},
	{b6.Tag{Key: "#amenity", Value: "doctors"}, "hospital"},
	// The most frequent landuse tag values from: https://taginfo.openstreetmap.org/keys/shop#values
	{b6.Tag{Key: "#shop", Value: "convenience"}, "grocery"},
	{b6.Tag{Key: "#shop", Value: "supermarket"}, "grocery"},
	{b6.Tag{Key: "#shop", Value: "clothes"}, "clothing-store"},
	{b6.Tag{Key: "#shop", Value: "hairdresser"}, "hairdresser"},
	{b6.Tag{Key: "#shop", Value: "car_repair"}, "car-repair"},
	{b6.Tag{Key: "#shop", Value: "bakery"}, "bakery"},
	{b6.Tag{Key: "#shop", Value: "alcohol"}, "alcohol-shop"},
	{b6.Tag{Key: "#shop", Value: "yes"}, "shop"},
	{b6.Tag{Key: "#shop", Value: ""}, "shop"},
	// The most frequent building tag values from: https://taginfo.openstreetmap.org/keys/building#values
	{b6.Tag{Key: "#building", Value: "house"}, "home"},
	{b6.Tag{Key: "#building", Value: "detached"}, "home"},
	{b6.Tag{Key: "#building", Value: "semidetached_house"}, "home"},
	{b6.Tag{Key: "#building", Value: "residential"}, "home"},
	{b6.Tag{Key: "#building", Value: "garage"}, "car"},
	{b6.Tag{Key: "#building", Value: "church"}, "religious-christian"}, // TODO: breakdown
	{b6.Tag{Key: "#building", Value: "school"}, "school"},
	{b6.Tag{Key: "#building", Value: "apartments"}, "home"},
	{b6.Tag{Key: "#building", Value: "commercial"}, "shop"},
	{b6.Tag{Key: "#building", Value: "retail"}, "shop"},
	{b6.Tag{Key: "#building", Value: "farm"}, "farm"},
	{b6.Tag{Key: "#building", Value: "industrial"}, "industry"},
	{b6.Tag{Key: "#building", Value: "garage"}, "car"},
	{b6.Tag{Key: "#building", Value: "garages"}, "car"},
	{b6.Tag{Key: "#building", Value: "hotel"}, "loding"},
	{b6.Tag{Key: "#building", Value: "stable"}, "horse-riding"},
	{b6.Tag{Key: "#building", Value: "train_station"}, "rail"},
	{b6.Tag{Key: "#building", Value: "yes"}, "building"},
	{b6.Tag{Key: "#building", Value: "other"}, "building"}, // Only for analysis results
	{b6.Tag{Key: "#building"}, "building"},
	// The most frequent landuse tag values from: https://taginfo.openstreetmap.org/keys/landuse#values
	{b6.Tag{Key: "#landuse", Value: "farmland"}, "farm"},
	{b6.Tag{Key: "#landuse", Value: "forest"}, "park"},  // As the icon is a tree
	{b6.Tag{Key: "#landuse", Value: "orchard"}, "park"}, // As the icon is a tree
	{b6.Tag{Key: "#landuse", Value: "industrial"}, "industry"},
	{b6.Tag{Key: "#landuse", Value: "vineyard"}, "alcohol-shop"}, // As the icon is a wine bottle & glass
	// The most frequent leisure tag values from: https://taginfo.openstreetmap.org/keys/landuse#values
	{b6.Tag{Key: "#leisure", Value: "pitch"}, "soccer"},
	{b6.Tag{Key: "#leisure", Value: "swimming_pool"}, "swimming"},
	{b6.Tag{Key: "#leisure", Value: "park"}, "park"},
	{b6.Tag{Key: "#leisure", Value: "garden"}, "garden"},
	{b6.Tag{Key: "#leisure", Value: "playground"}, "playground"},
	{b6.Tag{Key: "#leisure", Value: "sports_centre"}, "soccer"},
	{b6.Tag{Key: "#leisure", Value: "nature_reserve"}, "park"},
	{b6.Tag{Key: "#leisure", Value: "allotments"}, "garden"},
}

func IconForTag(t b6.Tag) string {
	for _, i := range iconsForTag {
		if i.Tag.Key == t.Key && (i.Tag.Value == t.Value || i.Tag.Value == "") {
			return i.Icon
		}
	}
	return "dot"
}

func IconForFeature(f b6.Feature) string {
	for _, i := range iconsForTag {
		if t := f.Get(i.Tag.Key); t.IsValid() && (i.Tag.Value == "" || i.Tag.Value == t.Value) {
			return i.Icon
		}
	}
	return "dot"
}
