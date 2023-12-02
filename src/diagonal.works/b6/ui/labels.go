package ui

import (
	"strings"

	"diagonal.works/b6"
)

type Label struct {
	Singular string
	Plural   string
}

var labelsForTag = []struct {
	Tag    b6.Tag
	Labels []Label
}{
	{b6.Tag{Key: "#place", Value: "uprn"}, []Label{{"Property", "Properties"}}},
	{b6.Tag{Key: "#boundary", Value: "datazone"}, []Label{{"Data zone", "Data zones"}}},
	{b6.Tag{Key: "#building", Value: "house"}, []Label{{"House", "Houses"}}},
	{b6.Tag{Key: "#building", Value: "detached"}, []Label{{"Detached house", "Detached houses"}, {"House", "Houses"}}},
	{b6.Tag{Key: "#building", Value: "semidetached_house"}, []Label{{"Pub", "Pubs"}, {"House", "Houses"}}},
	{b6.Tag{Key: "#building", Value: "residential"}, []Label{{"Residential", "Residences"}}},
	{b6.Tag{Key: "#building", Value: "garage"}, []Label{{"Garage", "Garages"}}},
	{b6.Tag{Key: "#building", Value: "church"}, []Label{{"Church", "Churches"}}},
	{b6.Tag{Key: "#building", Value: "school"}, []Label{{"School", "Schools"}}},
	{b6.Tag{Key: "#building", Value: "apartments"}, []Label{{"Apartments", "Apartments"}}},
	{b6.Tag{Key: "#building", Value: "commercial"}, []Label{{"Commercial", "Commercial"}}},
	{b6.Tag{Key: "#building", Value: "retail"}, []Label{{"Retail", "Retail"}}},
	{b6.Tag{Key: "#building", Value: "roof"}, []Label{{"Roof", "Roofs"}}},
	{b6.Tag{Key: "#building", Value: "terrace"}, []Label{{"Terrace", "Terraces"}}},
	{b6.Tag{Key: "#building", Value: "static_caravan"}, []Label{{"Static caravan", "Static caravans"}}},
	{b6.Tag{Key: "#building", Value: "farm"}, []Label{{"Farm", "Farms"}}},
	{b6.Tag{Key: "#building", Value: "industrial"}, []Label{{"Industrial", "Industrial"}}},
	{b6.Tag{Key: "#building", Value: "garage"}, []Label{{"Garage", "Garages"}}},
	{b6.Tag{Key: "#building", Value: "garages"}, []Label{{"Garages", "Garages"}}},
	{b6.Tag{Key: "#building", Value: "stable"}, []Label{{"Stable", "Stables"}}},
	{b6.Tag{Key: "#building", Value: "shed"}, []Label{{"Shed", "Sheds"}}},
	{b6.Tag{Key: "#building", Value: "ruins"}, []Label{{"Ruins", "Ruins"}}},
	{b6.Tag{Key: "#building", Value: "greenhouse"}, []Label{{"Greenhouse", "Greenhouses"}}},
	{b6.Tag{Key: "#building", Value: "hotel"}, []Label{{"Hotel", "Hotels"}}},
	{b6.Tag{Key: "#building", Value: "university"}, []Label{{"University", "Universities"}}},
	{b6.Tag{Key: "#building", Value: "yes"}, []Label{{"Building", "Buildings"}}},
	{b6.Tag{Key: "#building", Value: "other"}, []Label{{"Other", "Other"}}}, // Only for analysis results
	{b6.Tag{Key: "#building"}, []Label{{"Building", "Buildings"}}},
	// The most frequent amenity tag values from: https://taginfo.openstreetmap.org/keys/amenity#values
	{b6.Tag{Key: "#amenity", Value: "parking"}, []Label{{"Parking", "Parking"}}},
	{b6.Tag{Key: "#amenity", Value: "bench"}, []Label{{"Bench", "Benches"}}},
	{b6.Tag{Key: "#amenity", Value: "parking_space"}, []Label{{"Parking space", "Parking spaces"}}},
	{b6.Tag{Key: "#amenity", Value: "place_of_worship"}, []Label{{"Place of worship", "Places of worship"}}},
	{b6.Tag{Key: "#amenity", Value: "restaurant"}, []Label{{"Restaurant", "Restaurants"}}},
	{b6.Tag{Key: "#amenity", Value: "school"}, []Label{{"School", "Schools"}}},
	{b6.Tag{Key: "#amenity", Value: "pub"}, []Label{{"Pub", "Pubs"}}},
	// The most frequent landuse tag values from: https://taginfo.openstreetmap.org/keys/shop#values
	{b6.Tag{Key: "#shop", Value: "convenience"}, []Label{{"Convenience shop", "Convenience shops"}, {"Food shop", "Food shops"}}},
	{b6.Tag{Key: "#shop", Value: "supermarket"}, []Label{{"Supermarket", "Supermarkets"}, {"Food shop", "Food shops"}}},
	{b6.Tag{Key: "#shop", Value: "clothes"}, []Label{{"Clothes shop", "Clothes shops"}}},
	{b6.Tag{Key: "#shop", Value: "hairdresser"}, []Label{{"Hairdresser", "Hairdressers"}}},
	{b6.Tag{Key: "#shop", Value: "car_repair"}, []Label{{"Car repair", "Car repair"}}},
	{b6.Tag{Key: "#shop", Value: "bakery"}, []Label{{"Bakery", "Bakeries"}}},
	{b6.Tag{Key: "#shop", Value: "beauty"}, []Label{{"Beauty salon", "Beauty salons"}}},
	{b6.Tag{Key: "#shop", Value: "mobile_phone"}, []Label{{"Mobile phone shop", "Mobile phone shops"}}},
	{b6.Tag{Key: "#shop", Value: "butcher"}, []Label{{"Butcher", "Butchers"}}},
	{b6.Tag{Key: "#shop", Value: "furniture"}, []Label{{"Furniture shop", "Furniture shops"}}},
	{b6.Tag{Key: "#shop", Value: "alcohol"}, []Label{{"Off-license", "Off-licenses"}}},
	{b6.Tag{Key: "#shop", Value: "yes"}, []Label{{"Shop", "Shops"}}},
	{b6.Tag{Key: "#shop", Value: ""}, []Label{{"Shop", "Shops"}}},
	// The most frequent landuse tag values from: https://taginfo.openstreetmap.org/keys/landuse#values
	{b6.Tag{Key: "#landuse", Value: "farmland"}, []Label{{"Farmland", "Farmland"}}},
	{b6.Tag{Key: "#landuse", Value: "grass"}, []Label{{"Grass", "Grass"}}},
	{b6.Tag{Key: "#landuse", Value: "forest"}, []Label{{"Forest", "Forests"}}},
	{b6.Tag{Key: "#landuse", Value: "meadow"}, []Label{{"Meadow", "Meadows"}}},
	{b6.Tag{Key: "#landuse", Value: "orchard"}, []Label{{"Orchard", "Orchards"}}},
	{b6.Tag{Key: "#landuse", Value: "industrial"}, []Label{{"Industrial area", "Industrial areas"}}},
	{b6.Tag{Key: "#landuse", Value: "vineyard"}, []Label{{"Vineyard", "Vineyards"}}},
	{b6.Tag{Key: "#landuse", Value: "cemetery"}, []Label{{"Cemetery", "Cemeteries"}}},
	{b6.Tag{Key: "#landuse", Value: "allotments"}, []Label{{"Allotments", "Allotments"}}},
	{b6.Tag{Key: "#natural", Value: "heath"}, []Label{{"Heath", "Heaths"}}},
	// The most frequent leisure tag values from: https://taginfo.openstreetmap.org/keys/landuse#values
	{b6.Tag{Key: "#leisure", Value: "pitch"}, []Label{{"Pitch", "Pitches"}}},
	{b6.Tag{Key: "#leisure", Value: "swimming_pool"}, []Label{{"Swimming pool", "Swimming pools"}}},
	{b6.Tag{Key: "#leisure", Value: "park"}, []Label{{"Park", "Parks"}}},
	{b6.Tag{Key: "#leisure", Value: "garden"}, []Label{{"Garden", "Gardens"}}},
	{b6.Tag{Key: "#leisure", Value: "playground"}, []Label{{"Playground", "Playgrounds"}}},
	{b6.Tag{Key: "#leisure", Value: "sports_centre"}, []Label{{"Sports centre", "Sports centres"}}},
	{b6.Tag{Key: "#leisure", Value: "nature_reserve"}, []Label{{"Nature reserve", "Nature reserves"}}},
	{b6.Tag{Key: "#leisure", Value: "track"}, []Label{{"Track", "Tracks"}}},
	{b6.Tag{Key: "#leisure", Value: "allotments"}, []Label{{"Allotments", "Allotments"}}},
}

func featureLabel(f b6.Feature) string {
	if name := f.Get("name"); name.IsValid() {
		return name.Value
	} else if code := f.Get("code"); code.IsValid() {
		return code.Value
	} else if ref := f.Get("ref"); ref.IsValid() {
		return ref.Value
	} else {
		switch f.FeatureID().Namespace {
		case b6.NamespaceGBCodePoint:
			if postcode, ok := b6.PostcodeFromPointID(f.FeatureID().ToPointID()); ok {
				return postcode
			}
		case b6.NamespaceUKONSBoundaries:
			if code, _, ok := b6.UKONSCodeFromFeatureID(f.FeatureID()); ok {
				return code
			}
		}
	}
	return LabelForFeature(f).Singular
}

func LabelForTag(t b6.Tag) (Label, bool) {
	for _, l := range labelsForTag {
		if l.Tag.Key == t.Key && l.Tag.Value == t.Value {
			return l.Labels[0], true
		}
	}
	return Label{}, false
}

func LabelForFeature(f b6.Feature) Label {
	for _, t := range labelsForTag {
		if tt := f.Get(t.Tag.Key); tt.IsValid() && (t.Tag.Value == "" || t.Tag.Value == tt.Value) {
			return t.Labels[0]
		}
	}
	switch f.(type) {
	case b6.PointFeature:
		return Label{Singular: "Point", Plural: "Points"}
	case b6.PathFeature:
		return Label{Singular: "Path", Plural: "Paths"}
	case b6.AreaFeature:
		return Label{Singular: "Area", Plural: "Areas"}
	case b6.RelationFeature:
		return Label{Singular: "Relation", Plural: "Relations"}
	}
	return Label{Singular: "Feature", Plural: "Features"}
}

var labelsForNamespace = map[b6.Namespace]string{
	b6.NamespaceOSMNode:         "OSM",
	b6.NamespaceOSMWay:          "OSM",
	b6.NamespaceOSMRelation:     "OSM",
	b6.NamespaceUKONSBoundaries: "ONS",
	b6.NamespaceGBUPRN:          "UPRN",
	b6.NamespaceGBCodePoint:     "Codepoint",
	b6.NamespaceGTFS:            "GTFS",
}

var labelsForNamespacePrefixes = map[b6.Namespace]string{
	b6.NamespaceDiagonalEntrances:    "Entrance",
	b6.NamespaceDiagonalAccessPaths:  "Access",
	b6.NamespaceDiagonalAccessPoints: "Access",
}

func LabelForNamespace(ns b6.Namespace) string {
	if l, ok := labelsForNamespace[ns]; ok {
		return l
	}
	for prefix, l := range labelsForNamespacePrefixes {
		if strings.HasPrefix(string(ns), string(prefix)) {
			return l
		}
	}
	return ""
}
