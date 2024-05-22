package uprn

import (
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"github.com/golang/geo/s2"

	"github.com/apache/beam/sdks/go/pkg/beam/io/filesystem"
)

type Filter func(c *api.Context, f b6.Feature) (bool, error)

type Source struct {
	Filename string
	Filter   Filter
	Context  *api.Context
	JoinTags ingest.JoinTags
}

func (s *Source) Read(options ingest.ReadOptions, emit ingest.Emit, ctx context.Context) error {
	fs, err := filesystem.New(ctx, s.Filename)
	if err != nil {
		return err
	}
	defer fs.Close()

	f, err := fs.OpenRead(ctx, s.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if options.SkipPoints {
		return nil
	}

	d, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	r := csv.NewReader(d)
	header, err := r.Read()
	if err != nil {
		return err
	}
	fields := []string{"UPRN", "LATITUDE", "LONGITUDE"}
	columns := make([]int, len(fields))
	for i, f := range fields {
		columns[i] = -1
		for j, column := range header {
			if f == strings.Trim(column, "\ufeff") { // Remove byte order mark
				columns[i] = j
				break
			}
		}
		if columns[i] < 0 {
			return fmt.Errorf("Failed to find field %q", f)
		}
	}

	goroutines := options.Goroutines
	if goroutines < 1 {
		goroutines = 1
	}
	filtered := filter(s.Filter, emit, options, s.Context)
	p, wait := ingest.ParalleliseEmit(filtered, goroutines, ctx)
	if err := s.read(r, p, columns, goroutines); err != nil {
		wait()
		return err
	}
	return wait()
}

func (s *Source) read(r *csv.Reader, emit ingest.Emit, columns []int, goroutines int) error {
	ps := make([]ingest.Feature, goroutines*2)
	for i := range ps {
		ps[i] = &ingest.GenericFeature{}
		ps[i].SetFeatureID(b6.FeatureID{Type: b6.FeatureTypePoint, Namespace: b6.NamespaceGBUPRN})
		ps[i].AddTag(b6.Tag{Key: "#place", Value: b6.StringExpression("uprn")})
	}

	uprns := 0
	joined := 0
	joined_tags := 0
	var err error
	for {
		var row []string
		row, err = r.Read()
		if err == io.EOF {
			log.Printf("uprns: %d joined: %d joined tags: %d", uprns, joined, joined_tags)
			return nil
		} else if err != nil {
			return err
		}
		slot := uprns % len(ps)
		uprns++

		value, err := strconv.ParseUint(row[columns[0]], 10, 64)
		if err != nil {
			return fmt.Errorf("Parsing id for %q: %s", row, err)
		}
		ps[slot].SetFeatureID(b6.FeatureID{Type: b6.FeatureTypePoint, Namespace: b6.NamespaceGBUPRN, Value: value})

		lat, err := strconv.ParseFloat(row[columns[1]], 64)
		if err != nil {
			return fmt.Errorf("Parsing latitude for %q: %s", row, err)
		}
		lng, err := strconv.ParseFloat(row[columns[2]], 64)
		if err != nil {
			return fmt.Errorf("Parsing longitude for %q: %s", row, err)
		}

		ps[slot].RemoveAllTags()
		ps[slot].AddTag(b6.Tag{Key: "#place", Value: b6.StringExpression("uprn")})
		ps[slot].ModifyOrAddTag(b6.Tag{Key: b6.PointTag, Value: b6.LatLng(s2.LatLngFromDegrees(lat, lng))})
		s.JoinTags.AddTags(row[columns[0]], ps[slot])
		if len(ps[slot].AllTags()) > 1 {
			joined++
			joined_tags += len(ps[slot].AllTags()) - 1
		}
		if err := emit(ps[slot], slot%goroutines); err != nil {
			return err
		}
	}
}

func filter(filter Filter, emit ingest.Emit, options ingest.ReadOptions, ctx *api.Context) ingest.Emit {
	goroutines := options.Goroutines
	if goroutines < 1 {
		goroutines = 1
	}
	ctxs := ctx.Fork(goroutines)
	return func(f ingest.Feature, goroutine int) error {
		var err error
		keep := true
		if f.FeatureID().Type == b6.FeatureTypePoint {
			if keep, err = filter(ctxs[goroutine], f); err != nil {
				return err
			}
		}
		if keep {
			return emit(f, goroutine)
		}
		return nil
	}
}

// ClusterSourceS2Level is the S2 cell level used for clustering nearby UPRNs.
// Level 25 has cells with edges around 30cm in length.
const ClusterSourceS2Level = 25

// ClusterSource returns single points for locations at which a number
// of UPRNs are present.
type ClusterSource struct {
	UPRNs     ingest.FeatureSource // Usually uprn.Source, above, except for testing
	centroids map[s2.CellID]uint16
}

func (s *ClusterSource) Read(options ingest.ReadOptions, emit ingest.Emit, ctx context.Context) error {
	if !options.SkipPoints {
		if s.centroids == nil {
			if err := s.fillCentroids(options, ctx); err != nil {
				return err
			}
		}

		goroutines := options.Goroutines
		if goroutines < 1 {
			goroutines = 1
		}
		features := make([]ingest.Feature, goroutines*2)
		for i := range features {
			features[i] = &ingest.GenericFeature{Tags: []b6.Tag{{Key: "#place", Value: b6.StringExpression("uprn_cluster")}, {Key: "uprn_cluster:size", Value: b6.StringExpression("0")}}}
		}
		parallelised, wait := ingest.ParalleliseEmit(emit, goroutines, ctx)
		clusters := 0
		for c, count := range s.centroids {
			slot := clusters % len(features)
			clusters++
			features[slot].SetFeatureID(b6.FeatureID{Type: b6.FeatureTypePoint, Namespace: b6.NamespaceDiagonalUPRNCluster, Value: uint64(c)})
			features[slot].ModifyOrAddTag(b6.Tag{Key: "uprn_cluster:size", Value: b6.StringExpression(strconv.Itoa(int(count)))})
			features[slot].ModifyOrAddTag(b6.Tag{Key: b6.PointTag, Value: b6.LatLng(c.LatLng())})
			if err := parallelised(features[slot], slot%goroutines); err != nil {
				wait()
				return err
			}
		}
		return wait()
	}
	return nil
}

type byCellID []s2.CellID

func (b byCellID) Len() int           { return len(b) }
func (b byCellID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byCellID) Less(i, j int) bool { return b[i] < b[j] }

func (s *ClusterSource) fillCentroids(options ingest.ReadOptions, ctx context.Context) error {
	goroutines := options.Goroutines
	if goroutines < 1 {
		goroutines = 1
	}
	centroids := make([][]s2.CellID, goroutines)
	for i := range centroids {
		centroids[i] = make([]s2.CellID, 0, 1024)
	}
	addUprn := func(f ingest.Feature, goroutine int) error {
		if f, ok := f.(b6.Geometry); ok && f.GeometryType() == b6.GeometryTypePoint {
			centroid := s2.CellIDFromLatLng(s2.LatLngFromPoint(f.Point())).Parent(ClusterSourceS2Level)
			centroids[goroutine] = append(centroids[goroutine], centroid)
		}
		return nil
	}
	s.UPRNs.Read(options, addUprn, ctx)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			sort.Sort(byCellID(centroids[i]))
		}(i)
	}
	wg.Wait()
	s.centroids = make(map[s2.CellID]uint16, len(centroids[0])*goroutines)
	for _, cs := range centroids {
		for _, c := range cs {
			s.centroids[c]++
		}
	}
	return nil
}
