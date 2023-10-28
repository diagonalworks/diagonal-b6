package ui

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
	"diagonal.works/b6/ingest"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/renderer"
	"github.com/golang/geo/s2"
	"google.golang.org/protobuf/encoding/protojson"
)

type Options struct {
	StaticPath        string
	JavaScriptPath    string
	Renderer          UIRenderer
	Cores             int
	World             ingest.MutableWorld
	APIOptions        api.Options
	InstrumentHandler func(handler http.Handler, name string) http.Handler
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

	var uiRenderer UIRenderer
	if options.Renderer != nil {
		uiRenderer = options.Renderer
	} else {
		uiRenderer = NewDefaultUIRenderer(options.World)
	}
	startup := http.Handler(&StartupHandler{World: options.World, Renderer: uiRenderer})
	if options.InstrumentHandler != nil {
		startup = options.InstrumentHandler(startup, "startup")
	}
	root.Handle("/startup", startup)
	ui := http.Handler(NewUIHandler(uiRenderer, options.World, options.APIOptions))
	if options.InstrumentHandler != nil {
		ui = options.InstrumentHandler(ui, "ui")
	}
	root.Handle("/ui", ui)

	return nil
}

func NewDefaultUIRenderer(w b6.World) UIRenderer {
	return &DefaultUIRenderer{
		World:           w,
		FunctionSymbols: functions.Functions(),
		RenderRules:     renderer.BasemapRenderRules,
	}
}

func RegisterTiles(root *http.ServeMux, options *Options) {
	base := http.Handler(&renderer.TileHandler{Renderer: &renderer.BasemapRenderer{RenderRules: renderer.BasemapRenderRules, World: options.World}})
	if options.InstrumentHandler != nil {
		base = options.InstrumentHandler(base, "tiles_base")
	}
	root.Handle("/tiles/base/", base)
	query := http.Handler(&renderer.TileHandler{Renderer: renderer.NewQueryRenderer(options.World, options.Cores)})
	if options.InstrumentHandler != nil {
		query = options.InstrumentHandler(query, "tiles_query")
	}
	root.Handle("/tiles/query/", query)
}

type UIRenderer interface {
	Render(response *UIResponseJSON, value interface{}, context b6.RelationFeature, locked bool) error
}

type FeatureIDProtoJSON pb.FeatureIDProto

func (b *FeatureIDProtoJSON) MarshalJSON() ([]byte, error) {
	return protojson.Marshal((*pb.FeatureIDProto)(b))
}

func (b *FeatureIDProtoJSON) UnmarshalJSON(buffer []byte) error {
	return protojson.Unmarshal(buffer, (*pb.FeatureIDProto)(b))
}

type LatLngJSON struct {
	LatE7 int `json:"latE7"`
	LngE7 int `json:"lngE7"`
}

type StartupResponseJSON struct {
	Version       string              `json:"version,omitempty"`
	Docked        []*UIResponseJSON   `json:"docked,omitempty"`
	OpenDockIndex *int                `json:"openDockIndex,omitempty"`
	MapCenter     *LatLngJSON         `json:"mapCenter,omitempty"`
	MapZoom       int                 `json:"mapZoom,omitempty"`
	Context       *FeatureIDProtoJSON `json:"context,omitempty"`
	Expression    string              `json:"expression,omitempty"`
}

const DefaultMapZoom = 16

type StartupHandler struct {
	World    b6.World
	Renderer UIRenderer
}

func (s *StartupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := &StartupResponseJSON{
		Version: b6.BackendVersion,
	}

	if r := r.URL.Query().Get("r"); len(r) > 0 {
		context := b6.FeatureIDFromString(r[1:])
		if context.IsValid() {
			s.fillStartupResponseFromRootFeature(response, context)
			response.Context = (*FeatureIDProtoJSON)(b6.NewProtoFromFeatureID(context))
		}
	}

	if ll := r.URL.Query().Get("ll"); len(ll) > 0 {
		if parts := strings.Split(ll, ","); len(parts) == 2 {
			if lat, err := strconv.ParseFloat(parts[0], 64); err == nil {
				if lng, err := strconv.ParseFloat(parts[1], 64); err == nil {
					response.MapCenter = &LatLngJSON{
						LatE7: int(lat * 1e7),
						LngE7: int(lng * 1e7),
					}
				}
			}
		}
	}

	if z := r.URL.Query().Get("z"); len(z) > 0 {
		if zi, err := strconv.ParseInt(z, 10, 64); err == nil {
			response.MapZoom = int(zi)
		}
	}

	if d := r.URL.Query().Get("d"); len(d) > 0 {
		if di, err := strconv.ParseInt(d, 10, 64); err == nil {
			i := int(di)
			response.OpenDockIndex = &i
		}
	}

	if e := r.URL.Query().Get("e"); len(e) > 0 {
		response.Expression = e
	}

	output, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(output))
}

func (s *StartupHandler) fillStartupResponseFromRootFeature(response *StartupResponseJSON, id b6.FeatureID) {
	if f := s.World.FindFeatureByID(id); f != nil {
		if relation, ok := f.(b6.RelationFeature); ok {
			for i := 0; i < relation.Len(); i++ {
				if member := s.World.FindFeatureByID(relation.Member(i).ID); member != nil {
					if relation.Member(i).Role == "docked" {
						uiResponse := NewUIResponseJSON()
						if err := s.Renderer.Render(uiResponse, member, relation, true); err == nil {
							response.Docked = append(response.Docked, uiResponse)
						}
					} else if relation.Member(i).Role == "centroid" {
						if p, ok := member.(b6.PhysicalFeature); ok {
							ll := s2.LatLngFromPoint(b6.Centroid(p))
							response.MapCenter = &LatLngJSON{
								LatE7: int(ll.Lat.E7()),
								LngE7: int(ll.Lng.E7()),
							}
							response.MapZoom = DefaultMapZoom
						}
					}
				}
			}
		}
	}
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
