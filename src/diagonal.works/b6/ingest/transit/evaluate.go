package transit

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"diagonal.works/b6"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/osm"
)

const (
	RelationIDSouthbound214 osm.RelationID = 8194082
	TripIDSouthbound214     TripID         = "214:1"

	RelationIDNorthbound88 osm.RelationID = 5463939
	TripIDNorthbound88     TripID         = "88:2"
)

type PathPair struct {
	A b6.FeatureID
	B b6.FeatureID
}

func evaluate(tripID TripID, relationID b6.RelationID, w b6.World, network *Network) ([]b6.FeatureID, error) {
	log.Printf("Evaluating %s", tripID)
	relation := b6.FindRelationByID(relationID, w)
	if relation == nil {
		return nil, fmt.Errorf("Failed to find OSM relation %s for trip %s", relationID, tripID)
	}

	trip, ok := network.Trips[tripID]
	if !ok {
		return nil, fmt.Errorf("Failed to find trip %q", tripID)
	}
	log.Printf("  Stops: %d", len(trip.Stops))

	reference := make(map[PathPair]bool)
	for i := 0; i < relation.Len()-1; i++ {
		this := relation.Member(i)
		next := relation.Member(i + 1)
		if this.ID.Type == b6.FeatureTypePath && next.ID.Type == b6.FeatureTypePath {
			reference[PathPair{A: this.ID, B: next.ID}] = true
		}
	}

	pathIDs, err := ConflateTrip(trip, w, network)
	if err != nil {
		return nil, err
	}

	matched := 0
	for i := 0; i < len(pathIDs)-1; i++ {
		pair := PathPair{pathIDs[i], pathIDs[i+1]}
		if reference[pair] {
			matched++
		}
	}
	log.Printf("  Reference: %d", len(reference))
	log.Printf("  Conflated: %d", len(pathIDs))
	log.Printf("  Matched: %d", matched)

	return pathIDs, nil
}

func renderPath(pathID b6.FeatureID, w b6.World) []*geojson.Feature {
	var features []*geojson.Feature
	if path := w.FindFeatureByID(pathID); path != nil {
		if p, ok := path.(b6.PhysicalFeature); ok {
			features = []*geojson.Feature{geojson.NewFeatureFromS2Polyline(*p.Polyline())}
		}
	}
	return features
}

func render(trip *Trip, conflated []b6.FeatureID, reference b6.RelationID, filename string, w b6.World) error {
	collection := geojson.NewFeatureCollection()
	for _, pathID := range conflated {
		for _, feature := range renderPath(pathID, w) {
			feature.Properties["colour"] = "#ff0000"
			feature.Properties["align"] = "left"
			collection.AddFeature(feature)
		}
	}

	for _, stop := range trip.Stops {
		feature := geojson.NewFeatureFromS2LatLng(stop.Stop.Location)
		feature.Properties["colour"] = "#ff0000"
		collection.AddFeature(feature)
	}

	if relation := b6.FindRelationByID(reference, w); relation != nil {
		for i := 0; i < relation.Len(); i++ {
			member := relation.Member(i)
			if member.ID.Type == b6.FeatureTypePath {
				for _, feature := range renderPath(member.ID, w) {
					feature.Properties["colour"] = "#00ff00"
					feature.Properties["align"] = "right"
					collection.AddFeature(feature)
				}
			}
		}
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	output, err := json.MarshalIndent(collection, "", "   ")
	f.Write(output)
	f.Close()
	return nil
}

func Evaluate(networks []*Network, w b6.World, cores int) error {
	trips := []struct {
		tripID     TripID
		relationID b6.RelationID
	}{
		{TripIDSouthbound214, b6.MakeRelationID(b6.NamespaceOSMRelation, uint64(RelationIDSouthbound214))},
		{TripIDNorthbound88, b6.MakeRelationID(b6.NamespaceOSMRelation, uint64(RelationIDNorthbound88))},
	}

	for i, test := range trips {
		log.Printf("Evaluating %s", test.tripID)
		for _, network := range networks {
			if _, ok := network.Trips[test.tripID]; !ok {
				continue
			}
			if pathIDs, err := evaluate(test.tripID, test.relationID, w, network); err == nil {
				if i == 0 {
					if err := render(network.Trips[test.tripID], pathIDs, test.relationID, "output.geojson", w); err != nil {
						log.Fatal(err)
					}
				}
			} else {
				return err
			}
		}
	}
	return nil
}
