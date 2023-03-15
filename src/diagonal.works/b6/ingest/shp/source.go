package shp

import (
	"archive/zip"
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"unicode"

	"diagonal.works/b6"
	"diagonal.works/b6/geometry"
	"diagonal.works/b6/ingest"

	"github.com/golang/geo/s2"
	"github.com/lukeroth/gdal"
)

type IDStrategy func(value string, i int) (uint64, error)

var (
	IndexIDStrategy IDStrategy = func(value string, i int) (uint64, error) {
		return uint64(i), nil
	}

	StripNonDigitsIDStrategy IDStrategy = func(value string, i int) (uint64, error) {
		stripped := ""
		for _, r := range value {
			if unicode.IsDigit(r) {
				stripped += string(r)
			}
		}
		return strconv.ParseUint(stripped, 10, 64)
	}

	HashIDStrategy IDStrategy = func(value string, i int) (uint64, error) {
		h := fnv.New64()
		h.Write([]byte(value))
		return h.Sum64(), nil
	}
)

type Source struct {
	Filename   string
	Bounds     s2.Rect
	Namespace  b6.Namespace
	IDField    string
	IDStrategy IDStrategy
	Tags       []b6.Tag

	wkt string
	ts  []gdal.CoordinateTransform
}

func newFeatureFromS2Region(r s2.Region) ingest.Feature {
	switch g := r.(type) {
	case s2.Point:
		return &ingest.PointFeature{
			PointID:  b6.PointIDInvalid,
			Tags:     []b6.Tag{},
			Location: s2.LatLngFromPoint(g),
		}
	case *s2.Polyline:
		f := ingest.NewPathFeature(len(*g))
		for i, p := range *g {
			f.SetLatLng(i, s2.LatLngFromPoint(p))
		}
		return f
	case *s2.Polygon:
		f := ingest.NewAreaFeature(1)
		f.SetPolygon(0, g)
		return f
	case geometry.MultiPolygon:
		f := ingest.NewAreaFeature(len(g))
		for i, p := range g {
			f.SetPolygon(i, p)
		}
		return f
	}
	return nil
}

type feature struct {
	Feature *gdal.Feature
	Index   int
}

func layerBounds(l gdal.Layer, t gdal.CoordinateTransform) s2.Rect {
	e, err := l.Extent(false)
	if err == nil {
		xs := []float64{
			e.MinX(), e.MaxX(), e.MinX(), e.MaxX(),
		}
		ys := []float64{
			e.MinY(), e.MinY(), e.MaxY(), e.MaxY(),
		}
		zs := []float64{0.0, 0.0, 0.0, 0.0}
		t.Transform(len(xs), xs, ys, zs)
		var bounds s2.Rect
		for i := range xs {
			bounds = bounds.AddPoint(s2.LatLngFromDegrees(xs[i], ys[i]))
		}
		return bounds
	}
	return s2.FullRect()
}

func intersects(a s2.Rect, b s2.Rect) bool {
	for _, r := range []s2.Rect{a, b} {
		if !r.IsValid() || r.IsFull() {
			return true
		}
	}
	return a.Intersects(b)
}

func (s *Source) Read(options ingest.ReadOptions, emit ingest.Emit, ctx context.Context) error {
	source := gdal.OpenDataSource(s.Filename, 0)
	defer source.Destroy()

	// TODO: Find a better way of error checking - this is the only way we can tell
	// the file exists?
	if layers := source.LayerCount(); layers != 1 {
		return fmt.Errorf("Expected 1 layer in %s, found %d", s.Filename, layers)
	}

	layer := source.LayerByIndex(0)
	if err := s.fillTransformers(layer, options.Cores); err != nil {
		return err
	}

	if !intersects(layerBounds(layer, s.ts[0]), s.Bounds) {
		return nil
	}

	definition := layer.Definition()
	names := make([]string, definition.FieldCount())
	types := make([]gdal.FieldType, definition.FieldCount())
	for i := 0; i < definition.FieldCount(); i++ {
		d := definition.FieldDefinition(i)
		names[i] = d.Name()
		types[i] = d.Type()
	}

	idField := -1
	if s.IDField != "" {
		idField = definition.FieldIndex(s.IDField)
		if idField < 0 {
			return fmt.Errorf("Can't find ID field %q", s.IDField)
		}
	}

	outside := uint64(0)
	done := make(chan error, options.Cores)
	convert := func(toConvert <-chan feature, goroutine int) {
		var err error
		for feature := range toConvert {
			g := feature.Feature.Geometry()
			if options.SkipPoints && (g.Type() == gdal.GT_Point || g.Type() == gdal.GT_MultiPoint) {
				continue
			}
			if options.SkipPaths && (g.Type() == gdal.GT_LineString || g.Type() == gdal.GT_MultiLineString) {
				continue
			}
			if options.SkipAreas && (g.Type() == gdal.GT_Polygon || g.Type() == gdal.GT_MultiPolygon) {
				continue
			}
			err = g.Transform(s.ts[goroutine])
			if err != nil {
				break
			}
			var gg s2.Region
			gg, err = GDALGeometryToS2Region(g)
			if err != nil {
				break
			}
			if !intersects(gg.RectBound(), s.Bounds) {
				atomic.AddUint64(&outside, 1)
				continue
			}
			f := newFeatureFromS2Region(gg)
			for i, t := range types {
				switch t {
				case gdal.FT_Integer:
					f.AddTag(b6.Tag{Key: names[i], Value: strconv.Itoa(feature.Feature.FieldAsInteger(i))})
				case gdal.FT_String:
					f.AddTag(b6.Tag{Key: names[i], Value: feature.Feature.FieldAsString(i)})
				}
			}
			idFieldValue := ""
			if idField >= 0 {
				switch types[idField] {
				case gdal.FT_Integer:
					idFieldValue = strconv.Itoa(feature.Feature.FieldAsInteger(idField))
				case gdal.FT_String:
					idFieldValue = feature.Feature.FieldAsString(idField)
				}
			}
			var id uint64
			id, err = s.IDStrategy(idFieldValue, feature.Index)
			if err != nil {
				break
			}
			switch f := f.(type) {
			case *ingest.PointFeature:
				f.PointID = b6.MakePointID(s.Namespace, uint64(id))
			case *ingest.PathFeature:
				f.PathID = b6.MakePathID(s.Namespace, uint64(id))
			case *ingest.AreaFeature:
				f.AreaID = b6.MakeAreaID(s.Namespace, uint64(id))
			}
			for _, tag := range s.Tags {
				f.AddTag(tag)
			}
			emit(f, goroutine)
		}
		if err != nil {
			log.Println(err.Error())
		}
		done <- err
	}

	c := make(chan feature)
	cores := options.Cores
	if cores < 1 {
		cores = 1
	}
	for i := 0; i < cores; i++ {
		go convert(c, i)
	}

	i := 0
	running := cores
	// TODO: also read context cancel channel
	var err error
	for err == nil && running > 0 {
		f := layer.NextFeature()
		if f == nil {
			break
		}
		select {
		case err = <-done:
			running--
			break
		case c <- feature{Feature: f, Index: i}:
		}
		i++
	}
	close(c)
	for running > 0 {
		<-done
		running--
	}
	if !s.Bounds.IsFull() {
		log.Printf("%d features, %d outside bounds", i, outside)
	}
	return err
}

func (s *Source) fillTransformers(layer gdal.Layer, n int) error {
	wgs84 := gdal.CreateSpatialReference("")
	defer wgs84.Destroy()
	if err := wgs84.FromEPSG(EPSGCodeWGS84); err != nil {
		return err
	}

	r := layer.SpatialReference()
	wkt, err := r.ToWKT()
	if err != nil || s.wkt != wkt {
		s.wkt = wkt
		for _, t := range s.ts {
			t.Destroy()
		}
		if s.ts == nil {
			s.ts = s.ts[0:]
		} else {
			s.ts = make([]gdal.CoordinateTransform, 0, n)
		}
	}

	for len(s.ts) < n {
		// CoordinateTransform's aren't thread safe - make one per goroutine
		t := gdal.CreateCoordinateTransform(r, wgs84)
		s.ts = append(s.ts, t)
	}
	return nil
}

func xwalkZip(z *zip.ReadCloser, tmp string, emit func(filename string) error) error {
	d, err := os.MkdirTemp(tmp, "shp")
	if err != nil {
		return err
	}
	defer os.RemoveAll(d)
	for _, f := range z.File {
		target := filepath.Join(d, f.Name)
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
	return AllZippedShapefiles(d, tmp, emit)
}

// AllZippedShapefiles called emit() once for every shapefile under directory.
// If there's a zipfile containing a shapefile, it's extracted into a temporary
// directory under tmp first.
func AllZippedShapefiles(directory string, tmp string, emit func(filename string) error) error {
	f := func(path string, info fs.FileInfo, err error) error {
		switch filepath.Ext(path) {
		case ".shp":
			return emit(path)
		case ".zip":
			z, err := zip.OpenReader(path)
			if err != nil {
				return err
			}
			for _, f := range z.File {
				if filepath.Ext(f.Name) == ".shp" {
					//emit(filepath.Join("/vsizip", path, f.Name))
					if err := emit(fmt.Sprintf("/vsizip/%s/%s", path, f.Name)); err != nil {
						z.Close()
						return err
					}
				}
			}
			z.Close()
		}
		return nil
	}
	return filepath.Walk(directory, f)
}
