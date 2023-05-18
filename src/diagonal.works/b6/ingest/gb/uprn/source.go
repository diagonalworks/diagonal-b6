package uprn

import (
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/geojson"
	"diagonal.works/b6/ingest"
	"github.com/golang/geo/s2"

	"github.com/apache/beam/sdks/go/pkg/beam/io/filesystem"
)

type Filter func(f b6.Feature, c *api.Context) (bool, error)

type Source struct {
	Filename string
	Filter   func(f b6.Feature, c *api.Context) (bool, error)
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
	err = s.read(r, p, columns, goroutines)
	return wait()
}

func (s *Source) read(r *csv.Reader, emit ingest.Emit, columns []int, goroutines int) error {
	ps := make([]ingest.PointFeature, goroutines*2)
	for i := range ps {
		ps[i] = ingest.PointFeature{
			PointID: b6.PointID{
				Namespace: b6.NamespaceGBUPRN,
			},
			Tags: []b6.Tag{{Key: "#place", Value: "uprn"}},
		}
	}

	uprns := 0
	joins := 0
	var err error
	for {
		var row []string
		row, err = r.Read()
		if err == io.EOF {
			log.Printf("uprns: %d joined: %d", uprns, joins)
			return nil
		} else if err != nil {
			return err
		}
		slot := uprns % len(ps)
		uprns++
		ps[slot].PointID.Value, err = strconv.ParseUint(row[columns[0]], 10, 64)
		if err != nil {
			return fmt.Errorf("Parsing id for %q: %s", row, err)
		}
		lat, err := strconv.ParseFloat(row[columns[1]], 64)
		if err != nil {
			return fmt.Errorf("Parsing latitude for %q: %s", row, err)
		}
		lng, err := strconv.ParseFloat(row[columns[2]], 64)
		if err != nil {
			return fmt.Errorf("Parsing longitude for %q: %s", row, err)
		}
		ps[slot].Location = s2.LatLngFromDegrees(lat, lng)
		ps[slot].Tags = ps[slot].Tags[0:1] // Keep #place=uprn
		s.JoinTags.AddTags(row[columns[0]], &ps[slot])
		if len(ps[slot].Tags) > 1 {
			joins++
		}
		if err := emit(&ps[slot], slot%goroutines); err != nil {
			return err
		}
	}
}

type point struct {
	*ingest.PointFeature
}

func (p point) PointID() b6.PointID {
	return p.PointFeature.PointID
}

func (p point) Covering(coverer s2.RegionCoverer) s2.CellUnion {
	return s2.CellUnion([]s2.CellID{p.CellID().Parent(coverer.MaxLevel)})
}

func (p point) ToGeoJSON() geojson.GeoJSON {
	return b6.PointFeatureToGeoJSON(p)
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
		if p, ok := f.(*ingest.PointFeature); ok {
			if keep, err = filter(point{p}, ctxs[goroutine]); err != nil {
				return err
			}
		}
		if keep {
			return emit(f, goroutine)
		}
		return nil
	}
}
