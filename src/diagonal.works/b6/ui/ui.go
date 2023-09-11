package ui

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/renderer"
)

type Options struct {
	StaticPath     string
	JavaScriptPath string
	Cores          int
	World          ingest.MutableWorld
}

func RegisterWebInterface(root *http.ServeMux, options *Options) error {
	root.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(options.StaticPath, "index.html"))
		} else {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	}))
	for _, filename := range []string{"b6.css", "palette.html"} {
		filename := filename
		root.Handle("/"+filename, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(options.StaticPath, filename))
		}))
	}
	root.Handle("/bundle.js", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(options.JavaScriptPath, "bundle.js"))
	}))
	root.Handle("/images/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := strings.LastIndex(r.URL.Path, "/")
		http.ServeFile(w, r, filepath.Join(options.StaticPath, "images", r.URL.Path[i+1:]))
	}))

	root.Handle("/bootstrap", http.HandlerFunc(serveBootstrap))
	root.Handle("/ui", NewUIHandler(options.World, options.Cores))

	return nil
}

func RegisterTiles(root *http.ServeMux, w b6.World, cores int) {
	base := &renderer.TileHandler{Renderer: &renderer.BasemapRenderer{RenderRules: renderer.BasemapRenderRules, World: w}}
	root.Handle("/tiles/base/", base)
	query := &renderer.TileHandler{Renderer: renderer.NewQueryRenderer(w, cores)}
	root.Handle("/tiles/query/", query)
}

type BootstrapResponseJSON struct {
	Version string
}

func serveBootstrap(w http.ResponseWriter, r *http.Request) {
	response := BootstrapResponseJSON{
		Version: b6.BackendVersion,
	}
	output, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(output))
}

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
	{b6.Tag{Key: "#building", Value: "semidetached_house"}, []Label{{"Semi-detached house", "Semi-detached houses"}}},
	{b6.Tag{Key: "#building", Value: "semidetached_house"}, []Label{{"Pub", "Pubs"}}},
	{b6.Tag{Key: "#building", Value: "residential"}, []Label{{"Residential", "Residences"}}},
	{b6.Tag{Key: "#building", Value: "garage"}, []Label{{"Garage", "Garages"}}},
	{b6.Tag{Key: "#building", Value: "church"}, []Label{{"Church", "Churches"}}},
	{b6.Tag{Key: "#building", Value: "detached"}, []Label{{"Detached house", "Detached houses"}}},
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
