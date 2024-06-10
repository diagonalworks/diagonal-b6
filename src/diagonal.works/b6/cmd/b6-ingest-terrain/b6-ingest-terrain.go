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
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"

	"github.com/golang/geo/s2"
	"github.com/lukeroth/gdal"

	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/gcs"
	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/local"
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

func readTerrainGrid(archive string, filename string, e emit) error {
	d, err := gdal.Open(fmt.Sprintf("/vsizip/%s/%s", archive, filename), 0)
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

func readTerrainZip(filename string, e emit) error {
	s, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("can't stat %s: %s", filename, err)
	}
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("can't open %s: %s", filename, err)
	}
	z, err := zip.NewReader(f, s.Size())
	ascs := make([]string, 0)
	for _, zf := range z.File {
		if strings.HasSuffix(zf.Name, ".asc") {
			ascs = append(ascs, zf.Name)
		}
	}
	f.Close()
	for _, asc := range ascs {
		if err := readTerrainGrid(filename, asc, e); err != nil {
			return err
		}
	}
	return nil
}

func readElevations(directory string, cores int) (b6.Elevations, error) {
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
			if err := readTerrainZip(filename, e); err != nil {
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
		Goroutines:    options.Goroutines,
	}
	points := uint64(0)
	elevations := uint64(0)
	each := func(f b6.Feature, goroutine int) error {
		if f, ok := f.(b6.PhysicalFeature); ok && f.GeometryType() == b6.GeometryTypePoint && !options.SkipTags {
			atomic.AddUint64(&points, 1)
			point := ingest.NewGenericFeatureFromWorld(f)
			paths := s.World.FindReferences(f.FeatureID(), b6.FeatureTypePath)
			for paths.Next() {
				path := paths.Feature()
				if path.Get("#highway").IsValid() {
					if e, ok := s.Elevations.Elevation(f.Point()); ok {
						atomic.AddUint64(&elevations, 1)
						point.AddTag(b6.Tag{Key: "ele", Value: b6.StringExpression(strconv.Itoa(int(math.Round(e))))})
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
	outputFlag := flag.String("output", "", "Output directory with OS terrain data")
	worldFlag := flag.String("world", "", "World to annotate with inclines")
	memory := flag.Bool("memory", true, "Use memory for intermediate data")
	scratch := flag.String("scratch", ".", "Directory for temporary files")
	coresFlag := flag.Int("cores", runtime.NumCPU(), "Number of cores available")
	flag.Parse()
	elevations, err := readElevations(*inputFlag, *coresFlag)
	if err != nil {
		log.Fatal(err)
	}

	w, err := compact.ReadWorld(*worldFlag, &ingest.BuildOptions{Cores: *coresFlag})
	if err != nil {
		log.Fatal(err)
	}

	points := compact.OutputTypeMemory
	if !*memory {
		points = compact.OutputTypeDisk
	}
	options := compact.Options{
		OutputFilename:          *outputFlag,
		Goroutines:              *coresFlag,
		ScratchDirectory:        *scratch,
		PointsScratchOutputType: points,
	}
	var finish func() error
	finish, err = compact.MaybeWriteToCloud(&options)
	if err == nil {
		source := elevationSource{World: w, Elevations: elevations}
		err = compact.Build(&source, &options)
		if err == nil {
			err = finish()
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}
