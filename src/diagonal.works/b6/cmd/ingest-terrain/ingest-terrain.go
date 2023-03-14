package main

// Ingest data from the OS Terrain 50 open dataset
// https://www.ordnancesurvey.co.uk/business-government/products/terrain-50
// Currently builds an S2 cell based point index from the terrain data,
// and outputs paths where there's a significant incline between the grid
// points closest to two consecutive path points.
// TODO: Since the terrain data is sparse, we need to smooth it to be
// useful.

import (
	"archive/zip"
	"context"
	"flag"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"

	"github.com/golang/geo/s2"
	"github.com/lukeroth/gdal"
)

func transform(x int, y int, t [6]float64) (float64, float64) {
	xx := t[0] + (float64(x) * t[1]) + float64(y)*t[2]
	yy := t[3] + (float64(x) * t[4]) + float64(y)*t[5]
	return xx, yy
}

func extractZip(filename string, directory string) error {
	z, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer z.Close()
	for _, f := range z.File {
		target := filepath.Join(directory, f.Name)
		r, err := f.Open()
		if err != nil {
			return err
		}
		w, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			r.Close()
			return err
		}
		_, err = io.Copy(w, r)
		r.Close()
		if err != nil {
			w.Close()
		} else {
			err = w.Close()
		}
		if err != nil {
			return err
		}
	}
	return nil
}

const EPSGCodeWGS84 = 4326

type emit func(s2.LatLng, float64) error

func readTerrainGrid(filename string, e emit) error {
	d, err := gdal.Open(filename, 0)
	if err != nil {
		return err
	}
	defer d.Close()
	wgs84 := gdal.CreateSpatialReference("")
	if err := wgs84.FromEPSG(EPSGCodeWGS84); err != nil {
		return err
	}
	source := gdal.CreateSpatialReference(d.Projection())
	defer source.Destroy()
	t := gdal.CreateCoordinateTransform(source, wgs84)
	defer t.Destroy()
	buffer := make([]float64, d.RasterXSize())
	xs := make([]float64, d.RasterXSize())
	ys := make([]float64, d.RasterXSize())
	zs := make([]float64, d.RasterXSize())
	for y := 0; y < d.RasterYSize(); y++ {
		if err := d.IO(0, 0, y, d.RasterXSize(), 1, buffer, d.RasterXSize(), 1, 1, []int{1}, 0, 0, 0); err != nil {
			return err
		}
		for x := 0; x < d.RasterXSize(); x++ {
			xx, yy := transform(x, y, d.GeoTransform())
			xs[x] = xx
			ys[x] = yy
		}
		t.Transform(len(xs), xs, ys, zs)
		for x := 0; x < d.RasterXSize(); x++ {
			if err := e(s2.LatLngFromDegrees(xs[x], ys[x]), buffer[x]); err != nil {
				return err
			}
		}
	}
	return nil
}

func readTerrainZip(filename string, tmp string, e emit) error {
	d, err := os.MkdirTemp(tmp, "ingest-terrain")
	if err != nil {
		return err
	}
	defer os.RemoveAll(d)
	if err := extractZip(filename, d); err != nil {
		return err
	}
	ascs, err := filepath.Glob(filepath.Join(d, "*.asc"))
	if err != nil {
		return err
	}
	for _, asc := range ascs {
		readTerrainGrid(asc, e)
	}
	return nil
}

func readElevations(directory string, tmp string, cores int) (b6.Elevations, error) {
	filenames, err := filepath.Glob(filepath.Join(directory, "data/??/*.zip"))
	if err != nil {
		return nil, err
	}
	elevations := b6.ElevationField{Radius: b6.MetersToAngle(25.0)}
	var lock sync.Mutex
	e := func(ll s2.LatLng, v float64) error {
		lock.Lock()
		elevations.Add(ll, v)
		lock.Unlock()
		return nil
	}
	log.Printf("Read %s", directory)
	c := make(chan string, cores)
	var wg sync.WaitGroup
	f := func() {
		defer wg.Done()
		for filename := range c {
			if err := readTerrainZip(filename, tmp, e); err != nil {
				log.Println(err.Error())
				return
			}
		}
	}
	wg.Add(cores)
	for i := 0; i < cores; i++ {
		go f()
	}

	for _, filename := range filenames {
		c <- filename
	}
	close(c)
	wg.Wait()
	log.Print("Finished read")
	elevations.Finish()
	return &elevations, nil
}

type elevationSource struct {
	World      b6.World
	Elevations b6.Elevations
}

func (s *elevationSource) Read(options ingest.ReadOptions, emit ingest.Emit, ctx context.Context) error {
	o := b6.EachFeatureOptions{
		SkipPoints:    options.SkipPoints,
		SkipPaths:     options.SkipPaths,
		SkipAreas:     options.SkipAreas,
		SkipRelations: options.SkipRelations,
		Parallelism:   options.Parallelism,
	}
	points := uint64(0)
	elevations := uint64(0)
	each := func(f b6.Feature, goroutine int) error {
		if f, ok := f.(b6.PointFeature); ok && !options.SkipTags {
			atomic.AddUint64(&points, 1)
			point := ingest.NewPointFeatureFromWorld(f)
			paths := s.World.FindPathsByPoint(f.PointID())
			for paths.Next() {
				path := paths.PathSegment().PathFeature
				if path.Get("#highway").IsValid() {
					if e, ok := s.Elevations.Elevation(f.Point()); ok {
						atomic.AddUint64(&elevations, 1)
						point.AddTag(b6.Tag{Key: "ele", Value: strconv.Itoa(int(math.Round(e)))})
					}
					break
				}
			}
			return emit(point, goroutine)
		}
		return emit(ingest.NewFeatureFromWorld(f), goroutine)
	}
	log.Printf("elevationSource: world features")
	if err := s.World.EachFeature(each, &o); err != nil {
		return err
	}
	log.Printf("elevationSource: %d points, %d elevations", points, elevations)
	return nil
}

func main() {
	inputFlag := flag.String("input", "", "Input directory with OS terrain data")
	outputFlag := flag.String("output", "", "Input directory with OS terrain data")
	tmpFlag := flag.String("tmp", "/tmp", "Temporary directory")
	worldFlag := flag.String("world", "", "World to annotate with inclines")
	coresFlag := flag.Int("cores", runtime.NumCPU(), "Number of cores available")
	flag.Parse()
	elevations, err := readElevations(*inputFlag, *tmpFlag, *coresFlag)
	if err != nil {
		log.Fatal(err)
	}

	w, err := compact.ReadWorld(*worldFlag, *coresFlag)
	if err != nil {
		log.Fatal(err)
	}

	options := compact.Options{
		OutputFilename:       *outputFlag,
		Cores:                *coresFlag,
		WorkDirectory:        "",
		PointsWorkOutputType: compact.OutputTypeMemory,
	}
	source := elevationSource{World: w, Elevations: elevations}
	if compact.Build(&source, &options); err != nil {
		log.Fatal(err)
	}
}
