package transit

import (
	"log"
	"sort"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/graph"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

const (
	NamespaceTFLBus b6.Namespace = "diagonal.works/ns/legacy/tfl/bus"

	NaptanAtcoNamespace = "NaptanAtco"
)

type Features struct {
	Within10Meters         bool
	Within20Meters         bool
	UseableByVehicle       bool
	PreferredByVehicle     bool
	HeadingTowardsNextStop bool
	NameMatchesTransitData bool
}

func (f *Features) Score() int {
	score := 0
	if f.Within10Meters {
		score += 1
	}
	if f.Within20Meters {
		score += 1
	}
	if f.UseableByVehicle {
		score += 1
	}
	if f.PreferredByVehicle {
		score += 1
	}
	if f.HeadingTowardsNextStop {
		score += 1
	}
	if f.NameMatchesTransitData {
		score += 1
	}
	return score
}

type projection struct {
	Path     b6.PhysicalFeature
	Point    s2.Point
	Feature  b6.PhysicalFeature
	Distance s1.Angle
	Features Features
}

type ByScoreThenDistance []projection

func (p ByScoreThenDistance) Len() int      { return len(p) }
func (p ByScoreThenDistance) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p ByScoreThenDistance) Less(i, j int) bool {
	if p[i].Features.Score() != p[j].Features.Score() {
		return p[i].Features.Score() > p[j].Features.Score()
	}
	return p[i].Distance < p[j].Distance
}

func project(point s2.Point, path b6.PhysicalFeature, w b6.World) (s2.Point, b6.PhysicalFeature) {
	polyline := path.Polyline()
	projection, vertex := polyline.Project(point)
	if vertex < path.GeometryLen() {
		next := w.FindFeatureByID(path.Reference(vertex).Source()).(b6.PhysicalFeature)
		previous := w.FindFeatureByID(path.Reference(vertex - 1).Source()).(b6.PhysicalFeature)
		if projection.Distance(next.Point()) < projection.Distance(previous.Point()) {
			return projection, next
		} else {
			return projection, previous
		}
	}
	return projection, w.FindFeatureByID(path.Reference(vertex - 1).Source()).(b6.PhysicalFeature)
}

func stitchWays(stops []TripStop, projections [][]projection, w b6.World) []b6.FeatureID {
	if len(stops) != len(projections) {
		panic("Expected projections for all stops")
	}

	pathIDs := make([]b6.FeatureID, 0, len(stops))
	for i := range stops {
		if i+1 < len(stops) {
			if len(projections[i]) > 0 && len(projections[i+1]) > 0 {
				from, to := projections[i][0].Feature, projections[i+1][0].Feature
				segments := graph.ComputeShortestPath(from.FeatureID(), to.FeatureID(), PathSearchMaxDistanceMeters, graph.BusWeights{}, w)
				for _, segment := range segments {
					pathIDs = append(pathIDs, segment.Feature.FeatureID())
				}
			}
		} else {
			if len(projections[i]) > 0 {
				pathIDs = append(pathIDs, projections[i][0].Path.FeatureID())
			}
		}
	}
	return pathIDs
}

const StopSearchRadiusMeters = 30
const PathSearchMaxDistanceMeters = 1000

func lookupNaptanStreet(stop *Stop, w b6.World) (string, bool) {
	var atco string
	for _, id := range stop.AlternateIDs {
		if id.Namespace == NaptanAtcoNamespace {
			atco = id.ID
			break
		}
	}
	if atco == "" {
		return "", false
	}
	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(stop.Location), b6.MetersToAngle(StopSearchRadiusMeters))
	q := b6.Intersection{b6.Keyed{"#highway"}, b6.MightIntersect{cap}}
	points := w.FindFeatures(q)
	for points.Next() {
		point := points.Feature()
		if code := point.Get("naptan:AtcoCode"); code.Value.String() == atco {
			street := point.Get("naptan:Street")
			return street.Value.String(), street.IsValid()
		}
	}
	return "", false
}

func isPathNamedTheSameAsNaptanStreet(path b6.PhysicalFeature, stop *Stop, w b6.World) (matches bool, ok bool) {
	street, ok := lookupNaptanStreet(stop, w)
	if !ok {
		return false, false
	}

	if name := path.Get("name"); name.IsValid() {
		// TODO: Handle street abbreviations
		if strings.ToLower(name.Value.String()) == strings.ToLower(street) {
			return true, true
		}
		return false, true
	}
	return false, false
}

func isPathHeadingTowardsPoint(path b6.PhysicalFeature, point s2.Point) bool {
	if oneway := path.Get("oneway"); !oneway.IsValid() || oneway.Value.String() != "yes" {
		return true
	}
	return point.Distance(path.PointAt(path.GeometryLen()-1)) < point.Distance(path.PointAt(0))
}

func Project(stop *Stop, network *Network, w b6.World) []projection {
	cap := s2.CapFromCenterAngle(s2.PointFromLatLng(stop.Location), b6.MetersToAngle(StopSearchRadiusMeters))
	q := b6.Intersection{b6.Union{b6.Keyed{"#highway"}, b6.Keyed{"#railway"}}, b6.MightIntersect{cap}}
	paths := w.FindFeatures(q)
	projections := make([]projection, 0, 8)
	nextStop := MostCommonNextStop(network, stop)
	for paths.Next() {
		path, ok := paths.Feature().(b6.PhysicalFeature)
		if !ok {
			continue
		}
		p := projection{Path: path}
		p.Point, p.Feature = project(s2.PointFromLatLng(stop.Location), path, w)
		if p.Feature == nil {
			continue
		}
		p.Distance = s2.PointFromLatLng(stop.Location).Distance(p.Point)
		if p.Distance < b6.MetersToAngle(15) {
			p.Features.Within10Meters = true
		}
		if p.Distance < b6.MetersToAngle(25) {
			p.Features.Within20Meters = true
		}
		if graph.IsPathUsableByBus(path) {
			p.Features.UseableByVehicle = true
		}
		if graph.IsPathPreferredByBus(path) {
			p.Features.PreferredByVehicle = true
		}
		if nextStop != nil {
			p.Features.HeadingTowardsNextStop = isPathHeadingTowardsPoint(path, s2.PointFromLatLng(nextStop.Location))
		} else {
			p.Features.HeadingTowardsNextStop = true
		}
		if matches, ok := isPathNamedTheSameAsNaptanStreet(path, stop, w); ok && matches {
			p.Features.NameMatchesTransitData = true
		}
		projections = append(projections, p)
	}
	sort.Sort(ByScoreThenDistance(projections))
	return projections
}

func ConflateTrip(trip *Trip, w b6.World, network *Network) ([]b6.FeatureID, error) {
	allProjections := make([][]projection, 0, len(trip.Stops))
	for _, stop := range trip.Stops {
		projections := Project(stop.Stop, network, w)
		allProjections = append(allProjections, projections)
	}
	return stitchWays(trip.Stops, allProjections, w), nil
}

func Conflate(networks []*Network, w b6.World, ns b6.Namespace, output string, cores int) error {
	// TODO: this can trivially be parallelised.
	id := uint64(0)
	builder := ingest.NewBasicWorldBuilder(&ingest.BuildOptions{})
	for _, network := range networks {
		for _, trip := range network.Trips {
			if paths, err := ConflateTrip(trip, w, network); err == nil {
				if len(paths) > 0 {
					log.Printf("%s", trip.Route.Name)
					relation := &ingest.RelationFeature{
						RelationID: b6.MakeRelationID(ns, id),
						Members:    make([]b6.RelationMember, len(paths)),
						Tags:       make(b6.Tags, 0, 2),
					}
					for i, pathID := range paths {
						relation.Members[i] = b6.RelationMember{
							ID: pathID.FeatureID(),
						}
					}
					relation.AddTag(b6.Tag{Key: "#type", Value: b6.NewStringExpression("route")})
					relation.AddTag(b6.Tag{Key: "#route", Value: b6.NewStringExpression("bus")})
					relation.AddTag(b6.Tag{Key: "ref", Value: b6.NewStringExpression(trip.Route.Name)})
					// TODO: Add name and other tags in the format from https://wiki.openstreetmap.org/wiki/Tag:route%3Dbus
					relation.AddTag(b6.Tag{Key: "source", Value: b6.NewStringExpression("diagonal")})
					builder.AddFeature(relation)
					// TODO: Allocate IDs based on a hash of an input ID?
					id++
				}
			} else {
				log.Printf("%s: %s: %s", trip.Route.Name, trip.ID, err)
			}
		}
	}
	world, err := builder.Finish(&ingest.BuildOptions{Cores: cores})
	if err != nil {
		return err
	}
	source := ingest.WorldFeatureSource{World: world}
	options := compact.Options{
		OutputFilename:          output,
		Goroutines:              cores,
		PointsScratchOutputType: compact.OutputTypeMemory,
	}
	return compact.Build(source, &options)
}
