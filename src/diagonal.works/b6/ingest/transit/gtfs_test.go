package transit

import (
	"testing"

	"diagonal.works/b6/test"
	"github.com/golang/geo/s2"
)

func TestParseGTFSWithManchesterData(t *testing.T) {
	network, err := ReadGTFS(test.Data("gtfs-manchester/"))
	if err != nil {
		t.Errorf("Expected no error, found %v", err)
	}

	if network.SkippedStops > 0 || network.SkippedTrips > 0 || network.SkippedTripStops > 0 {
		t.Errorf("Expected no skipped stops, trips or trip stops")
	}

	if len(network.Stops) < 100 {
		t.Errorf("Fewer than expected stops, found %d", len(network.Stops))
	}

	topLeft, bottomRight := s2.LatLngFromDegrees(53.665145, -2.723844), s2.LatLngFromDegrees(53.316270, -1.894376)
	manchester := s2.EmptyRect().AddPoint(topLeft).AddPoint(bottomRight)
	outsideMancheser := 0
	for _, stop := range network.Stops {
		if !manchester.ContainsLatLng(stop.Location) {
			outsideMancheser++
		}
	}
	if outsideMancheser > 0 {
		t.Errorf("Expected all stops to be within Manchester, found %d outside", outsideMancheser)
	}

	airport, ok := network.Trips[TripID("Trip010716")]
	if !ok {
		t.Errorf("Failed to find expected trip")
		return
	}

	expectedTripStops := 53
	if len(airport.Stops) != expectedTripStops {
		t.Errorf("Expected %d trip stops, found %d", expectedTripStops, len(airport.Stops))
	}

	westDidsbury := StopID("1800SB34381")
	if stop, ok := network.Stops[westDidsbury]; ok {
		if len(stop.Trips) < 200 {
			t.Errorf("Expected more than 200 trips for stop %q", westDidsbury)
		}
	} else {
		t.Errorf("Failed to find stop %q", westDidsbury)
	}
}
