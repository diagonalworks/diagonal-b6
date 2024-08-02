package transit

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang/geo/s2"
)

const (
	GTFSRoutesRouteColumn int = 0
)

type CSVReader struct {
	csv     *csv.Reader
	t       reflect.Type
	fields  []string
	indices []int
}

func NewReader(r io.Reader) (*CSVReader, error) {
	reader := &CSVReader{csv: csv.NewReader(r)}
	reader.csv.LazyQuotes = true
	reader.csv.FieldsPerRecord = -1
	var err error
	reader.fields, err = reader.csv.Read()
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func NewReaderFromFilename(filename string) (*CSVReader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return NewReader(file)
}

func normaliseField(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), "_", "")
}

func (r *CSVReader) Read(value interface{}) error {
	t := reflect.TypeOf(value).Elem()
	if r.t != t {
		keys := make(map[string]int)
		for i := 0; i < t.NumField(); i++ {
			keys[normaliseField(t.Field(i).Name)] = i
		}
		r.indices = make([]int, len(r.fields))
		for i, field := range r.fields {
			if j, ok := keys[normaliseField(field)]; ok {
				r.indices[i] = j
			} else {
				r.indices[i] = -1
			}
		}
		r.t = t
	}
	row, err := r.csv.Read()
	if err != nil {
		return err
	}
	v := reflect.ValueOf(value).Elem()
	for i := 0; i < len(row) && i < len(r.fields); i++ {
		if j := r.indices[i]; j >= 0 {
			v.Field(j).Set(reflect.ValueOf(row[i]))
		}
	}
	return nil
}

type GTFSLocationType int

const (
	GTFSLocationTypeStop         GTFSLocationType = 0
	GTFSLocationTypeStation                       = 1
	GTFSLocationTypeEntranceExit                  = 2
	GTFSLocationTypeGenericNode                   = 3
	GTFSLocationTypeBoardingArea                  = 4
)

type stopRow struct {
	StopID             string
	StopCode           string
	StopName           string
	StopDesc           string
	StopLat            string
	StopLon            string
	ZoneID             string
	StopURL            string
	LocationType       string
	ParentStation      string
	StopTimezone       string
	WheelchairBoarding string
	LevelID            string
	PlatformCode       string
}

func needsLocation(t GTFSLocationType) bool {
	return t == GTFSLocationTypeStop || t == GTFSLocationTypeStation || t == GTFSLocationTypeEntranceExit
}

func fillGTFSStops(directory string, network *Network) error {
	reader, err := NewReaderFromFilename(filepath.Join(directory, "stops.txt"))
	if err != nil {
		return err
	}
	network.Stops = make(map[StopID]*Stop)
	var row stopRow
	for {
		err := reader.Read(&row)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		locationType := GTFSLocationTypeStop
		if n, err := strconv.Atoi(row.LocationType); err == nil {
			locationType = GTFSLocationType(n)
		}
		stop := &Stop{ID: StopID(row.StopID), Name: row.StopName}
		var lat, lng float64
		if lat, err = strconv.ParseFloat(row.StopLat, 64); err != nil {
			if needsLocation(locationType) {
				network.SkippedStops++
				continue
			}
		}
		if lng, err = strconv.ParseFloat(row.StopLon, 64); err != nil {
			if needsLocation(locationType) {
				network.SkippedStops++
				continue
			}
		}
		stop.Location = s2.LatLngFromDegrees(lat, lng)
		network.Stops[stop.ID] = stop
	}
	return nil
}

type routeRow struct {
	RouteID           string
	AgencyID          string
	RouteShortName    string
	RouteLongName     string
	RouteDesc         string
	RouteType         string
	RouteURL          string
	RouteColor        string
	RouteTextColor    string
	RouteSortOrder    string
	ContinuousPickup  string
	ContinuousDropoff string
}

func fillGTFSRoutes(directory string, network *Network) error {
	reader, err := NewReaderFromFilename(filepath.Join(directory, "routes.txt"))
	if err != nil {
		return err
	}
	network.Routes = make(map[RouteID]*Route)
	var row routeRow
	for {
		err := reader.Read(&row)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		name := row.RouteLongName
		if name == "" {
			name = row.RouteShortName
		}
		route := &Route{ID: RouteID(row.RouteID), Name: name}
		network.Routes[route.ID] = route
	}
	return nil
}

type tripRow struct {
	RouteID              string
	ServiceID            string
	TripID               string
	TripHeadsign         string
	TripShortName        string
	DirectionID          string
	BlockID              string
	ShapeID              string
	WheelchairAccessible string
	BikesAllowed         string
}

func fillGTFSTrips(directory string, network *Network) error {
	reader, err := NewReaderFromFilename(filepath.Join(directory, "trips.txt"))
	if err != nil {
		return err
	}
	network.Trips = make(map[TripID]*Trip)
	var row tripRow
	for {
		err := reader.Read(&row)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		if route, ok := network.Routes[RouteID(row.RouteID)]; ok {
			trip := &Trip{ID: TripID(row.TripID), Route: route}
			network.Trips[trip.ID] = trip
		} else {
			network.SkippedTrips++
		}
	}
	return nil
}

type stopTimeRow struct {
	TripID             string
	ArrivalTime        string
	DepartureTime      string
	StopID             string
	StopSequence       string
	StopHeadsign       string
	PickupType         string
	DropOffType        string
	ContinuousPickup   string
	ContinuousDropoff  string
	ShapeDistTravelled string
	Timepoint          string
}

func fillGTFSTripStops(directory string, network *Network) error {
	reader, err := NewReaderFromFilename(filepath.Join(directory, "stop_times.txt"))
	if err != nil {
		return err
	}
	var row stopTimeRow
	for {
		err := reader.Read(&row)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		trip, tripOK := network.Trips[TripID(row.TripID)]
		stop, stopOK := network.Stops[StopID(row.StopID)]
		sequence, err := strconv.Atoi(row.StopSequence)
		if tripOK && stopOK && err == nil {
			if trip.Stops == nil {
				trip.Stops = make([]TripStop, 0, 1)
			}
			trip.Stops = append(trip.Stops, TripStop{Stop: stop, Sequence: sequence, ArrivalTime: row.ArrivalTime, DepartureTime: row.DepartureTime})
		} else {
			network.SkippedTripStops++
		}
	}
	return nil
}

func ReadGTFS(directory string) (*Network, error) {
	network := &Network{}
	if err := fillGTFSRoutes(directory, network); err != nil {
		return nil, err
	}
	if err := fillGTFSStops(directory, network); err != nil {
		return nil, err
	}
	if err := fillGTFSTrips(directory, network); err != nil {
		return nil, err
	}
	if err := fillGTFSTripStops(directory, network); err != nil {
		return nil, err
	}
	AddStopsToTrips(network)
	return network, nil
}
