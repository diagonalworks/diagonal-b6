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
	{b6.Tag{Key: "#place", Value: b6.NewStringExpression("uprn")}, []Label{{"Property", "Properties"}}},
	{b6.Tag{Key: "#boundary", Value: b6.NewStringExpression("datazone")}, []Label{{"Data zone", "Data zones"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("house")}, []Label{{"House", "Houses"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("detached")}, []Label{{"Detached house", "Detached houses"}, {"House", "Houses"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("semidetached_house")}, []Label{{"Pub", "Pubs"}, {"House", "Houses"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("residential")}, []Label{{"Residential", "Residences"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("garage")}, []Label{{"Garage", "Garages"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("church")}, []Label{{"Church", "Churches"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("school")}, []Label{{"School", "Schools"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("apartments")}, []Label{{"Apartments", "Apartments"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("commercial")}, []Label{{"Commercial", "Commercial"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("retail")}, []Label{{"Retail", "Retail"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("roof")}, []Label{{"Roof", "Roofs"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("terrace")}, []Label{{"Terrace", "Terraces"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("static_caravan")}, []Label{{"Static caravan", "Static caravans"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("farm")}, []Label{{"Farm", "Farms"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("industrial")}, []Label{{"Industrial", "Industrial"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("garage")}, []Label{{"Garage", "Garages"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("garages")}, []Label{{"Garages", "Garages"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("stable")}, []Label{{"Stable", "Stables"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("shed")}, []Label{{"Shed", "Sheds"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("ruins")}, []Label{{"Ruins", "Ruins"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("greenhouse")}, []Label{{"Greenhouse", "Greenhouses"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("hotel")}, []Label{{"Hotel", "Hotels"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("university")}, []Label{{"University", "Universities"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("yes")}, []Label{{"Building", "Buildings"}}},
	{b6.Tag{Key: "#building", Value: b6.NewStringExpression("other")}, []Label{{"Other", "Other"}}}, // Only for analysis results
	{b6.Tag{Key: "#building"}, []Label{{"Building", "Buildings"}}},
	// The most frequent amenity tag values from: https://taginfo.openstreetmap.org/keys/amenity#values
	{b6.Tag{Key: "#amenity", Value: b6.NewStringExpression("parking")}, []Label{{"Parking", "Parking"}}},
	{b6.Tag{Key: "#amenity", Value: b6.NewStringExpression("bench")}, []Label{{"Bench", "Benches"}}},
	{b6.Tag{Key: "#amenity", Value: b6.NewStringExpression("parking_space")}, []Label{{"Parking space", "Parking spaces"}}},
	{b6.Tag{Key: "#amenity", Value: b6.NewStringExpression("place_of_worship")}, []Label{{"Place of worship", "Places of worship"}}},
	{b6.Tag{Key: "#amenity", Value: b6.NewStringExpression("restaurant")}, []Label{{"Restaurant", "Restaurants"}}},
	{b6.Tag{Key: "#amenity", Value: b6.NewStringExpression("school")}, []Label{{"School", "Schools"}}},
	{b6.Tag{Key: "#amenity", Value: b6.NewStringExpression("pub")}, []Label{{"Pub", "Pubs"}}},
	// The most frequent landuse tag values from: https://taginfo.openstreetmap.org/keys/shop#values
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("convenience")}, []Label{{"Convenience shop", "Convenience shops"}, {"Food shop", "Food shops"}}},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("supermarket")}, []Label{{"Supermarket", "Supermarkets"}, {"Food shop", "Food shops"}}},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("clothes")}, []Label{{"Clothes shop", "Clothes shops"}}},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("hairdresser")}, []Label{{"Hairdresser", "Hairdressers"}}},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("car_repair")}, []Label{{"Car repair", "Car repair"}}},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("bakery")}, []Label{{"Bakery", "Bakeries"}}},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("beauty")}, []Label{{"Beauty salon", "Beauty salons"}}},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("mobile_phone")}, []Label{{"Mobile phone shop", "Mobile phone shops"}}},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("butcher")}, []Label{{"Butcher", "Butchers"}}},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("furniture")}, []Label{{"Furniture shop", "Furniture shops"}}},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("alcohol")}, []Label{{"Off-license", "Off-licenses"}}},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("yes")}, []Label{{"Shop", "Shops"}}},
	{b6.Tag{Key: "#shop", Value: b6.NewStringExpression("")}, []Label{{"Shop", "Shops"}}},
	// The most frequent landuse tag values from: https://taginfo.openstreetmap.org/keys/landuse#values
	{b6.Tag{Key: "#landuse", Value: b6.NewStringExpression("farmland")}, []Label{{"Farmland", "Farmland"}}},
	{b6.Tag{Key: "#landuse", Value: b6.NewStringExpression("grass")}, []Label{{"Grass", "Grass"}}},
	{b6.Tag{Key: "#landuse", Value: b6.NewStringExpression("forest")}, []Label{{"Forest", "Forests"}}},
	{b6.Tag{Key: "#landuse", Value: b6.NewStringExpression("meadow")}, []Label{{"Meadow", "Meadows"}}},
	{b6.Tag{Key: "#landuse", Value: b6.NewStringExpression("orchard")}, []Label{{"Orchard", "Orchards"}}},
	{b6.Tag{Key: "#landuse", Value: b6.NewStringExpression("industrial")}, []Label{{"Industrial area", "Industrial areas"}}},
	{b6.Tag{Key: "#landuse", Value: b6.NewStringExpression("vineyard")}, []Label{{"Vineyard", "Vineyards"}}},
	{b6.Tag{Key: "#landuse", Value: b6.NewStringExpression("cemetery")}, []Label{{"Cemetery", "Cemeteries"}}},
	{b6.Tag{Key: "#landuse", Value: b6.NewStringExpression("allotments")}, []Label{{"Allotments", "Allotments"}}},
	{b6.Tag{Key: "#natural", Value: b6.NewStringExpression("heath")}, []Label{{"Heath", "Heaths"}}},
	// The most frequent leisure tag values from: https://taginfo.openstreetmap.org/keys/landuse#values
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("pitch")}, []Label{{"Pitch", "Pitches"}}},
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("swimming_pool")}, []Label{{"Swimming pool", "Swimming pools"}}},
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("park")}, []Label{{"Park", "Parks"}}},
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("garden")}, []Label{{"Garden", "Gardens"}}},
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("playground")}, []Label{{"Playground", "Playgrounds"}}},
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("sports_centre")}, []Label{{"Sports centre", "Sports centres"}}},
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("nature_reserve")}, []Label{{"Nature reserve", "Nature reserves"}}},
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("track")}, []Label{{"Track", "Tracks"}}},
	{b6.Tag{Key: "#leisure", Value: b6.NewStringExpression("allotments")}, []Label{{"Allotments", "Allotments"}}},
}

func featureLabel(f b6.Feature) string {
	if name := f.Get("name"); name.IsValid() {
		return name.Value.String()
	} else if code := f.Get("code"); code.IsValid() {
		return code.Value.String()
	} else if ref := f.Get("ref"); ref.IsValid() {
		return ref.Value.String()
	} else {
		switch f.FeatureID().Namespace {
		case b6.NamespaceGBCodePoint:
			if postcode, ok := b6.PostcodeFromPointID(f.FeatureID()); ok {
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
		if tt := f.Get(t.Tag.Key); tt.IsValid() && (!t.Tag.IsValid() || t.Tag.Value == tt.Value) {
			return t.Labels[0]
		}
	}
	switch f.FeatureID().Type {
	case b6.FeatureTypePoint:
		return Label{Singular: "Point", Plural: "Points"}
	case b6.FeatureTypePath:
		return Label{Singular: "Path", Plural: "Paths"}
	case b6.FeatureTypeArea:
		return Label{Singular: "Area", Plural: "Areas"}
	case b6.FeatureTypeRelation:
		return Label{Singular: "Relation", Plural: "Relations"}
	case b6.FeatureTypeCollection:
		return Label{Singular: "Collection", Plural: "Collections"}
	case b6.FeatureTypeExpression:
		return Label{Singular: "Expression", Plural: "Expressions"}
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
