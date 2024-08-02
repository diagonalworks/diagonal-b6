package transit

import (
	"sort"

	"github.com/golang/geo/s2"
)

type StopID string

const InvalidStopID StopID = ""

type AlternateID struct {
	Namespace string
	ID        string
}

type Stop struct {
	ID           StopID
	Location     s2.LatLng
	Name         string
	AlternateIDs []AlternateID
	Trips        []TripIndex
}

type RouteID string

type Route struct {
	ID   RouteID
	Name string
}

type TripID string

type TripStop struct {
	Stop          *Stop
	Sequence      int
	ArrivalTime   string
	DepartureTime string
}

type TripIndex struct {
	Trip  *Trip
	Index int
}

type BySequence []TripStop

func (b BySequence) Len() int           { return len(b) }
func (b BySequence) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b BySequence) Less(i, j int) bool { return b[i].Sequence < b[j].Sequence }

type Trip struct {
	ID    TripID
	Route *Route
	Stops []TripStop
}

type Network struct {
	Stops  map[StopID]*Stop
	Routes map[RouteID]*Route
	Trips  map[TripID]*Trip

	SkippedStops     int
	SkippedTrips     int
	SkippedTripStops int
}

func AddStopsToTrips(network *Network) {
	for _, trip := range network.Trips {
		sort.Sort(BySequence(trip.Stops))
		for i, tripStop := range trip.Stops {
			if tripStop.Stop.Trips == nil {
				tripStop.Stop.Trips = make([]TripIndex, 0, 1)
			}
			tripStop.Stop.Trips = append(tripStop.Stop.Trips, TripIndex{Trip: trip, Index: i})
		}
	}
}

// MostCommonNextStop returns the stop that follows the specified stop on the majority of trips.
func MostCommonNextStop(network *Network, stop *Stop) *Stop {
	counts := make(map[StopID]int)
	for _, trip := range stop.Trips {
		if trip.Index+1 < len(trip.Trip.Stops) {
			counts[trip.Trip.Stops[trip.Index+1].Stop.ID]++
		}
	}

	commonID := InvalidStopID
	commonCount := 0
	for id, count := range counts {
		if count > commonCount || (count == commonCount && id > commonID) {
			commonID = id
			commonCount = count
		}
	}
	if commonID != InvalidStopID {
		return network.Stops[commonID]
	}
	return nil
}
