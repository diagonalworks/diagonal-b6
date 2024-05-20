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
	{b6.Tag{Key: "#place", Value: b6.StringExpression("uprn")}, []Label{{"Property", "Properties"}}},
	{b6.Tag{Key: "#boundary", Value: b6.StringExpression("datazone")}, []Label{{"Data zone", "Data zones"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("house")}, []Label{{"House", "Houses"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("detached")}, []Label{{"Detached house", "Detached houses"}, {"House", "Houses"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("semidetached_house")}, []Label{{"Pub", "Pubs"}, {"House", "Houses"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("residential")}, []Label{{"Residential", "Residences"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("garage")}, []Label{{"Garage", "Garages"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("church")}, []Label{{"Church", "Churches"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("school")}, []Label{{"School", "Schools"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("apartments")}, []Label{{"Apartments", "Apartments"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("commercial")}, []Label{{"Commercial", "Commercial"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("retail")}, []Label{{"Retail", "Retail"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("roof")}, []Label{{"Roof", "Roofs"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("terrace")}, []Label{{"Terrace", "Terraces"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("static_caravan")}, []Label{{"Static caravan", "Static caravans"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("farm")}, []Label{{"Farm", "Farms"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("industrial")}, []Label{{"Industrial", "Industrial"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("garage")}, []Label{{"Garage", "Garages"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("garages")}, []Label{{"Garages", "Garages"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("stable")}, []Label{{"Stable", "Stables"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("shed")}, []Label{{"Shed", "Sheds"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("ruins")}, []Label{{"Ruins", "Ruins"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("greenhouse")}, []Label{{"Greenhouse", "Greenhouses"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("hotel")}, []Label{{"Hotel", "Hotels"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("university")}, []Label{{"University", "Universities"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("yes")}, []Label{{"Building", "Buildings"}}},
	{b6.Tag{Key: "#building", Value: b6.StringExpression("other")}, []Label{{"Other", "Other"}}}, // Only for analysis results
	{b6.Tag{Key: "#building"}, []Label{{"Building", "Buildings"}}},
	// The most frequent amenity tag values from: https://taginfo.openstreetmap.org/keys/amenity#values
	{b6.Tag{Key: "#amenity", Value: b6.StringExpression("parking")}, []Label{{"Parking", "Parking"}}},
	{b6.Tag{Key: "#amenity", Value: b6.StringExpression("bench")}, []Label{{"Bench", "Benches"}}},
	{b6.Tag{Key: "#amenity", Value: b6.StringExpression("parking_space")}, []Label{{"Parking space", "Parking spaces"}}},
	{b6.Tag{Key: "#amenity", Value: b6.StringExpression("place_of_worship")}, []Label{{"Place of worship", "Places of worship"}}},
	{b6.Tag{Key: "#amenity", Value: b6.StringExpression("restaurant")}, []Label{{"Restaurant", "Restaurants"}}},
	{b6.Tag{Key: "#amenity", Value: b6.StringExpression("school")}, []Label{{"School", "Schools"}}},
	{b6.Tag{Key: "#amenity", Value: b6.StringExpression("pub")}, []Label{{"Pub", "Pubs"}}},
	// The most frequent landuse tag values from: https://taginfo.openstreetmap.org/keys/shop#values
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("convenience")}, []Label{{"Convenience shop", "Convenience shops"}, {"Food shop", "Food shops"}}},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("supermarket")}, []Label{{"Supermarket", "Supermarkets"}, {"Food shop", "Food shops"}}},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("clothes")}, []Label{{"Clothes shop", "Clothes shops"}}},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("hairdresser")}, []Label{{"Hairdresser", "Hairdressers"}}},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("car_repair")}, []Label{{"Car repair", "Car repair"}}},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("bakery")}, []Label{{"Bakery", "Bakeries"}}},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("beauty")}, []Label{{"Beauty salon", "Beauty salons"}}},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("mobile_phone")}, []Label{{"Mobile phone shop", "Mobile phone shops"}}},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("butcher")}, []Label{{"Butcher", "Butchers"}}},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("furniture")}, []Label{{"Furniture shop", "Furniture shops"}}},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("alcohol")}, []Label{{"Off-license", "Off-licenses"}}},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("yes")}, []Label{{"Shop", "Shops"}}},
	{b6.Tag{Key: "#shop", Value: b6.StringExpression("")}, []Label{{"Shop", "Shops"}}},
	// The most frequent landuse tag values from: https://taginfo.openstreetmap.org/keys/landuse#values
	{b6.Tag{Key: "#landuse", Value: b6.StringExpression("farmland")}, []Label{{"Farmland", "Farmland"}}},
	{b6.Tag{Key: "#landuse", Value: b6.StringExpression("grass")}, []Label{{"Grass", "Grass"}}},
	{b6.Tag{Key: "#landuse", Value: b6.StringExpression("forest")}, []Label{{"Forest", "Forests"}}},
	{b6.Tag{Key: "#landuse", Value: b6.StringExpression("meadow")}, []Label{{"Meadow", "Meadows"}}},
	{b6.Tag{Key: "#landuse", Value: b6.StringExpression("orchard")}, []Label{{"Orchard", "Orchards"}}},
	{b6.Tag{Key: "#landuse", Value: b6.StringExpression("industrial")}, []Label{{"Industrial area", "Industrial areas"}}},
	{b6.Tag{Key: "#landuse", Value: b6.StringExpression("vineyard")}, []Label{{"Vineyard", "Vineyards"}}},
	{b6.Tag{Key: "#landuse", Value: b6.StringExpression("cemetery")}, []Label{{"Cemetery", "Cemeteries"}}},
	{b6.Tag{Key: "#landuse", Value: b6.StringExpression("allotments")}, []Label{{"Allotments", "Allotments"}}},
	{b6.Tag{Key: "#natural", Value: b6.StringExpression("heath")}, []Label{{"Heath", "Heaths"}}},
	// The most frequent leisure tag values from: https://taginfo.openstreetmap.org/keys/landuse#values
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("pitch")}, []Label{{"Pitch", "Pitches"}}},
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("swimming_pool")}, []Label{{"Swimming pool", "Swimming pools"}}},
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("park")}, []Label{{"Park", "Parks"}}},
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("garden")}, []Label{{"Garden", "Gardens"}}},
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("playground")}, []Label{{"Playground", "Playgrounds"}}},
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("sports_centre")}, []Label{{"Sports centre", "Sports centres"}}},
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("nature_reserve")}, []Label{{"Nature reserve", "Nature reserves"}}},
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("track")}, []Label{{"Track", "Tracks"}}},
	{b6.Tag{Key: "#leisure", Value: b6.StringExpression("allotments")}, []Label{{"Allotments", "Allotments"}}},
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
