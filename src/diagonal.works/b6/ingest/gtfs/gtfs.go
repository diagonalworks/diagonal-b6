package gtfs

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"diagonal.works/b6"
	"diagonal.works/b6/graph"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/transit"
	"golang.org/x/sync/errgroup"
)

func isPeakTrafficTime(x time.Time) bool {
	morningStart, _ := time.Parse(time.TimeOnly, "08:00:00")
	morningEnd, _ := time.Parse(time.TimeOnly, "10:00:00")

	eveStart, _ := time.Parse(time.TimeOnly, "16:00:00")
	eveEnd, _ := time.Parse(time.TimeOnly, "18:00:00")

	// Checking the inverse; makes respecting inclusive [peak] time bounds a bit neater.
	return !(x.Before(morningStart) || x.After(eveEnd) || (x.After(morningEnd) && x.Before(eveStart)))
}

func sanitizeTime(s string) string {
	if len(s) > 2 { // Expects hh: format.
		if h, err := strconv.Atoi(strings.TrimPrefix(s[0:2], "0")); err == nil {
			leadingZero := ""
			if h%24 < 10 { // 24h is considered out of range.
				leadingZero = "0"
			}

			return leadingZero + fmt.Sprintf("%d", h%24) + s[2:]
		}
	}

	return s
}

func point(tripStop *transit.TripStop, operator string) ingest.Feature {
	h := fnv.New64()
	h.Write([]byte(string(tripStop.Stop.ID) + tripStop.Stop.Location.String()))
	id := h.Sum64()

	return &ingest.GenericFeature{
		ID: b6.FeatureID{Type: b6.FeatureTypePoint, Namespace: b6.Namespace(b6.NamespaceGTFS.String() + operator), Value: id},
		Tags: []b6.Tag{
			{Key: "#gtfs", Value: b6.NewStringExpression("stop")},
			{Key: b6.PointTag, Value: b6.NewPointExpressionFromLatLng(tripStop.Stop.Location)},
		},
	}
}

func travelTimes(from *transit.TripStop, to *transit.TripStop) ([]fraction, error) {
	departureTime, err := time.Parse(time.TimeOnly, sanitizeTime(from.DepartureTime))
	if err != nil {
		return nil, err
	}

	arrivalTime, err := time.Parse(time.TimeOnly, sanitizeTime(to.ArrivalTime))
	if err != nil {
		return nil, err
	}

	if arrivalTime.Before(departureTime) { // Midnight wrap.
		departureTime = departureTime.AddDate(0, 0, 0)
		arrivalTime = arrivalTime.AddDate(0, 0, 1)
	}

	travelTime :=
		fraction{
			numerator:   math.Round(arrivalTime.Sub(departureTime).Seconds()),
			denominator: 1,
		}

	// stopConnection expects peak / off-peak ordering of travel time.
	if isPeakTrafficTime(departureTime) || isPeakTrafficTime(arrivalTime) {
		return []fraction{travelTime, fraction{0, 0}}, nil
	} else {
		return []fraction{fraction{0, 0}, travelTime}, nil
	}
}

type fraction struct {
	numerator   float64
	denominator int32
}

type connection struct {
	from                     ingest.Feature
	to                       ingest.Feature
	averagePeakTravelTime    fraction
	averageOffPeakTravelTime fraction
	operator                 string
}

func stopConnection(from *transit.TripStop, to *transit.TripStop, operator string) (connection, error) {
	travelTimes, err := travelTimes(from, to)
	if err != nil {
		return connection{}, err
	}

	return connection{
		from: point(from, operator),
		to:   point(to, operator),
		// Relying on initialization for both .*TravelTimes in consolidateTravelTimes.
		averagePeakTravelTime:    travelTimes[0],
		averageOffPeakTravelTime: travelTimes[1],
		operator:                 operator,
	}, nil
}

func consolidateTravelTimes(m *[]connection) {
	group := make(map[[2]b6.FeatureID][]connection)
	for _, e := range *m {
		id := [2]b6.FeatureID{e.from.FeatureID(), e.to.FeatureID()}
		group[id] = append(group[id], e)
	}

	connections := make([]connection, 0)
	for _, duplicates := range group {
		var c connection
		for i, e := range duplicates {
			if i == 0 {
				c = e
				continue
			} // .from and .to don't vary, so setting only once.

			c.averagePeakTravelTime.numerator += e.averagePeakTravelTime.numerator
			c.averagePeakTravelTime.denominator += e.averagePeakTravelTime.denominator
			c.averageOffPeakTravelTime.numerator += e.averageOffPeakTravelTime.numerator
			c.averageOffPeakTravelTime.denominator += e.averageOffPeakTravelTime.denominator
		}

		connections = append(connections, c)
	}

	*m = connections
}

func tripConnections(n *transit.Network, operator string) ([]connection, error) {
	trips := n.Trips

	var connections []connection
	var err error

	for _, trip := range trips {
		sort.Sort(transit.BySequence(trip.Stops))
		for i := 0; i < len(trip.Stops)-1; i++ {
			c, e := stopConnection(&trip.Stops[i], &trip.Stops[i+1], operator)
			if e != nil {
				err = e
			}

			connections = append(connections, c)
		}
	}

	consolidateTravelTimes(&connections)

	return connections, err
}

func path(c *connection) *ingest.GenericFeature {
	path := &ingest.GenericFeature{
		Tags: []b6.Tag{
			{Key: "#gtfs", Value: b6.NewStringExpression("connection")},
			{
				Key:   b6.PathTag,
				Value: b6.NewExpressions([]b6.AnyExpression{b6.FeatureIDExpression(c.from.FeatureID()), b6.FeatureIDExpression(c.to.FeatureID())}),
			},
		},
	}

	h := fnv.New64()
	var buffer [8]byte
	binary.LittleEndian.PutUint64(buffer[0:], c.from.FeatureID().Value)
	h.Write(buffer[0:])
	binary.LittleEndian.PutUint64(buffer[0:], c.to.FeatureID().Value)
	h.Write(buffer[0:])

	if c.from.FeatureID().Value < c.to.FeatureID().Value {
		binary.LittleEndian.PutUint64(buffer[0:], 1)
	} else { // Distinguishing direction.
		binary.LittleEndian.PutUint64(buffer[0:], 2)
	}
	h.Write(buffer[0:])

	id := h.Sum64()

	path.ID = b6.FeatureID{Type: b6.FeatureTypePath, Namespace: b6.Namespace(b6.NamespaceGTFS.String() + c.operator), Value: id}

	if c.averagePeakTravelTime.denominator != 0 {
		path.AddTag(b6.Tag{
			Key: graph.GTFSPeakTimeTag,
			Value: b6.NewStringExpression(strconv.Itoa(int(
				math.Ceil(c.averagePeakTravelTime.numerator/
					float64(c.averagePeakTravelTime.denominator)/60) * 60))),
		})
	}

	if c.averageOffPeakTravelTime.denominator != 0 {
		path.AddTag(b6.Tag{
			Key: graph.GTFSOffPeakTimeTag,
			Value: b6.NewStringExpression(strconv.Itoa(int(
				math.Ceil(c.averageOffPeakTravelTime.numerator/
					float64(c.averageOffPeakTravelTime.denominator)/60) * 60))),
		})
	}

	return path
}

func emitFeatures(options ingest.ReadOptions, ch <-chan connection, emit ingest.Emit, goroutine int, ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case c, ok := <-ch:
			if !ok {
				return nil
			}

			if !options.SkipPoints {
				if err := emit(c.from, goroutine); err != nil {
					return err
				}
				if err := emit(c.to, goroutine); err != nil {
					return err
				}
			}

			if !options.SkipPaths {
				if err := emit(path(&c), goroutine); err != nil {
					return err
				}
			}
		}
	}
}

type TXTFilesGTFSSource struct {
	Directory string
	Operator  string
	// Expects routes, stops, stop_times, and trips [.txt] files under directory path.
	FailWhenNoFiles bool
}

type GTFSSource interface {
	Read(options ingest.ReadOptions, emit ingest.Emit, ctx context.Context) error
}

func (source *TXTFilesGTFSSource) Read(options ingest.ReadOptions, emit ingest.Emit, ctx context.Context) error {
	cores := options.Goroutines
	if cores < 1 {
		cores = 1
	}

	network, err := transit.ReadGTFS(source.Directory)
	if err != nil {
		return err
	}

	connections, err := tripConnections(network, source.Operator)
	if err != nil {
		return err
	}

	ch := make(chan connection, cores)
	g, gc := errgroup.WithContext(ctx)

	g.Go(func() error {
		for _, c := range connections {
			ch <- c
		}

		close(ch)
		return nil
	})
	for i := 0; i < cores; i++ {
		goroutine := i
		g.Go(func() error { return emitFeatures(options, ch, emit, goroutine, gc) })
	}
	g.Wait()

	return nil
}
