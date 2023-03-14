package main

import (
	"archive/zip"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	_ "net/http/pprof"
	"path/filepath"
	"runtime"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/ingest/compact"
	"diagonal.works/b6/ingest/shp"
	"github.com/golang/geo/s2"
)

var idStrategies = map[string]shp.IDStrategy{
	"":      shp.IndexIDStrategy,
	"strip": shp.StripNonDigitsIDStrategy,
	"hash":  shp.HashIDStrategy,
}

type recursiveSource struct {
	Root   string
	Zip    bool
	Source *shp.Source
}

func (s *recursiveSource) Read(options ingest.ReadOptions, emit ingest.Emit, ctx context.Context) error {
	read := func(filename string) error {
		s.Source.Filename = filename
		return s.Source.Read(options, emit, ctx)
	}
	walk := func(path string, info fs.FileInfo, err error) error {
		switch filepath.Ext(path) {
		case ".shp":
			log.Printf("Read: %s", path)
			return read(path)
		case ".zip":
			if s.Zip {
				z, err := zip.OpenReader(path)
				if err != nil {
					return err
				}
				fs := make([]string, 0)
				for _, f := range z.File {
					if filepath.Ext(f.Name) == ".shp" {
						fs = append(fs, f.Name)
					}
				}
				z.Close()
				for _, f := range fs {
					log.Printf("Read: %s:%s", path, f)
					if err := read(fmt.Sprintf("/vsizip/%s/%s", path, f)); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	return filepath.Walk(s.Root, walk)
}

func main() {
	addr := flag.String("addr", "", "Address to listen on for status over HTTP")
	input := flag.String("input", "", "Input shapefile")
	output := flag.String("output", "", "Output index")
	namespace := flag.String("namespace", "", "Namespace for features")
	id := flag.String("id", "", "Field to use for ID generation")
	idStategy := flag.String("id-strategy", "", "Strategy to use for ID generation")
	addTags := flag.String("add-tags", "", "Tags to add to imported data, eg #boundary=datazone,year=2011")
	recurse := flag.Bool("recurse", false, "Recurse into directories")
	zip := flag.Bool("zip", false, "Read shapefiles within zipfiles")
	boundingBox := flag.String("bounding-box", "", "lat,lng,lat,lng bounding box to crop points outside")
	cores := flag.Int("cores", runtime.NumCPU(), "Number of cores available")
	flag.Parse()

	if *addr != "" {
		go func() {
			log.Println(http.ListenAndServe(*addr, nil))
		}()
		log.Printf("Listening on %s", *addr)
	}

	strategy, ok := idStrategies[*idStategy]
	if !ok {
		ss := make([]string, 0, len(idStrategies))
		for key := range idStrategies {
			if key != "" {
				ss = append(ss, key)
			}
		}
		log.Fatalf("No ID strategy %q - try one of %v", *idStategy, ss)
	}

	bounds := s2.FullRect()
	if *boundingBox != "" {
		var err error
		bounds, err = ingest.ParseBoundingBox(*boundingBox)
		if err != nil {
			log.Fatal(err)
		}
	}

	s := &shp.Source{
		Filename:   *input,
		Namespace:  b6.Namespace(*namespace),
		IDField:    *id,
		IDStrategy: strategy,
		Bounds:     bounds,
	}

	for _, tag := range strings.Split(*addTags, ",") {
		parts := strings.Split(tag, "=")
		if len(parts) == 2 {
			s.Tags = append(s.Tags, b6.Tag{Key: parts[0], Value: parts[1]})
		}
	}

	options := compact.Options{
		OutputFilename:       *output,
		Cores:                *cores,
		WorkDirectory:        "",
		PointsWorkOutputType: compact.OutputTypeMemory,
	}

	var source ingest.FeatureSource
	if !*recurse {
		s.Filename = *input
		source = s
	} else {
		source = &recursiveSource{Root: *input, Zip: *zip, Source: s}
	}

	if err := compact.Build(source, &options); err != nil {
		log.Fatal(err)
	}
}
