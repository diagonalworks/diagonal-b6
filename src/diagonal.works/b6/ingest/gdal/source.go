package gdal

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"strconv"
	"strings"
	"unicode"

	"diagonal.works/b6"
	"diagonal.works/b6/geometry"
	"diagonal.works/b6/ingest"

	"github.com/golang/geo/s2"
	"github.com/lukeroth/gdal"
)

type IDStrategy func(value string, i int, t b6.FeatureType, ns b6.Namespace) (b6.FeatureID, error)

var (
	IndexIDStrategy IDStrategy = func(value string, i int, t b6.FeatureType, ns b6.Namespace) (b6.FeatureID, error) {
		return b6.FeatureID{Type: t, Namespace: ns, Value: uint64(i)}, nil
	}

	StripNonDigitsIDStrategy IDStrategy = func(value string, i int, t b6.FeatureType, ns b6.Namespace) (b6.FeatureID, error) {
		stripped := ""
		for _, r := range value {
			if unicode.IsDigit(r) {
				stripped += string(r)
			}
		}
		v, err := strconv.ParseUint(stripped, 10, 64)
		return b6.FeatureID{Type: t, Namespace: ns, Value: v}, err
	}

	HashIDStrategy IDStrategy = func(value string, i int, t b6.FeatureType, ns b6.Namespace) (b6.FeatureID, error) {
		h := fnv.New64()
		h.Write([]byte(value))
		return b6.FeatureID{Type: t, Namespace: ns, Value: h.Sum64()}, nil
	}

	UKONS2011IDStrategy IDStrategy = func(value string, i int, t b6.FeatureType, ns b6.Namespace) (b6.FeatureID, error) {
		return b6.FeatureIDFromUKONSCode(value, 2011, t), nil
	}

	UKONS2022IDStrategy IDStrategy = func(value string, i int, t b6.FeatureType, ns b6.Namespace) (b6.FeatureID, error) {
		return b6.FeatureIDFromUKONSCode(value, 2022, t), nil
	}

	UKONS2023IDStrategy IDStrategy = func(value string, i int, t b6.FeatureType, ns b6.Namespace) (b6.FeatureID, error) {
		return b6.FeatureIDFromUKONSCode(value, 2023, t), nil
	}
)

const (
	EPSGCodeWGS84 = 4326
)

func geometryTypeToString(t gdal.GeometryType) string {
	switch t {
	case gdal.GT_Unknown:
		return "GT_Unknown"
	case gdal.GT_Point:
		return "GT_Point"
	case gdal.GT_LineString:
		return "GT_LineString"
	case gdal.GT_Polygon:
		return "GT_Polygon"
	case gdal.GT_MultiPoint:
		return "GT_MultiPoint"
	case gdal.GT_MultiLineString:
		return "GT_MultiLineString"
	case gdal.GT_MultiPolygon:
		return "GT_MultiPolygon"
	case gdal.GT_GeometryCollection:
		return "GT_GeometryCollection"
	case gdal.GT_None:
		return "GT_None"
	case gdal.GT_LinearRing:
		return "GT_LinearRing"
	case gdal.GT_Point25D:
		return "GT_Point25D"
	case gdal.GT_LineString25D:
		return "GT_LineString25D"
	case gdal.GT_Polygon25D:
		return "GT_Polygon25D"
	case gdal.GT_MultiPoint25D:
		return "GT_MultiPoint25D"
	case gdal.GT_MultiLineString25D:
		return "GT_MultiLineString25D"
	case gdal.GT_MultiPolygon25D:
		return "GT_MultiPolygon25D"
	case gdal.GT_GeometryCollection25D:
		return "GT_GeometryCollection25D"
	default:
		return "Invalid"
	}
}

func geometryToS2Point(g gdal.Geometry) s2.Point {
	lat, lng, _ := g.Point(0)
	return s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng))
}

func geometryToS2Polyline(g gdal.Geometry) *s2.Polyline {
	points := make(s2.Polyline, g.PointCount())
	for i := 0; i < g.PointCount(); i++ {
		lat, lng, _ := g.Point(i)
		points[g.PointCount()-i-1] = s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng))
	}
	return &points
}

func geometryToS2Loop(g gdal.Geometry) *s2.Loop {
	points := make([]s2.Point, g.PointCount())
	for i := 0; i < g.PointCount(); i++ {
		lat, lng, _ := g.Point(i)
		points[g.PointCount()-i-1] = s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng))
	}
	return s2.LoopFromPoints(points)
}

func geometryToS2Polygon(g gdal.Geometry) *s2.Polygon {
	loops := make([]*s2.Loop, g.GeometryCount())
	for i := 0; i < g.GeometryCount(); i++ {
		loops[i] = geometryToS2Loop(g.Geometry(i))
	}
	return s2.PolygonFromLoops(loops)
}

func geometryToS2MultiPolygon(g gdal.Geometry) geometry.MultiPolygon {
	polygons := make(geometry.MultiPolygon, g.GeometryCount())
	for i := 0; i < g.GeometryCount(); i++ {
		polygons[i] = geometryToS2Polygon(g.Geometry(i))
	}
	return polygons
}

func geometryToS2Region(g gdal.Geometry) (s2.Region, error) {
	switch g.Type() {
	case gdal.GT_Point, gdal.GT_Point25D:
		return geometryToS2Point(g), nil
	case gdal.GT_LineString, gdal.GT_LineString25D:
		return geometryToS2Polyline(g), nil
	case gdal.GT_Polygon, gdal.GT_Polygon25D:
		return geometryToS2Polygon(g), nil
	case gdal.GT_MultiPolygon, gdal.GT_MultiPolygon25D:
		return geometryToS2MultiPolygon(g), nil
	}
	return nil, fmt.Errorf("Can't convert geometry type %s", geometryTypeToString(g.Type()))
}

func geometryTypeToFeatureType(t gdal.GeometryType) (b6.FeatureType, error) {
	switch t {
	case gdal.GT_Point, gdal.GT_Point25D:
		return b6.FeatureTypePoint, nil
	case gdal.GT_LineString, gdal.GT_LineString25D:
		return b6.FeatureTypePath, nil
	case gdal.GT_Polygon, gdal.GT_Polygon25D:
		return b6.FeatureTypeArea, nil
	case gdal.GT_MultiPolygon, gdal.GT_MultiPolygon25D:
		return b6.FeatureTypeArea, nil
	}
	return b6.FeatureTypeInvalid, fmt.Errorf("Can't convert geometry type %s", geometryTypeToString(t))
}

const batchTransformerBatchSize = 20000

type batchTransformer struct {
	Bounds           s2.Rect
	CopyTags         []CopyTag
	AddTags          []b6.Tag
	JoinTags         ingest.JoinTags
	Goroutines       int
	Emit             ingest.Emit
	SpatialReference gdal.SpatialReference

	geometries  gdal.Geometry
	originalIDs []string
	b6IDs       []b6.FeatureID
	tags        [][]b6.Tag
}

func (b *batchTransformer) TakeOwnershipAndTransform(g gdal.Geometry, originalID string, b6ID b6.FeatureID, tags []b6.Tag, ctx context.Context) error {
	if b.geometries.IsNull() {
		b.geometries = gdal.Create(gdal.GT_GeometryCollection)
		b.geometries.SetSpatialReference(b.SpatialReference)
	}
	b.geometries.AddGeometryDirectly(g)
	b.originalIDs = append(b.originalIDs, originalID)
	b.b6IDs = append(b.b6IDs, b6ID)
	b.tags = append(b.tags, tags)
	if len(b.originalIDs) > batchTransformerBatchSize {
		if err := b.Flush(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (b *batchTransformer) Flush(ctx context.Context) error {
	if b.geometries.IsNull() {
		return nil
	}

	wgs84 := gdal.CreateSpatialReference("")
	if err := wgs84.FromEPSG(EPSGCodeWGS84); err != nil {
		return err
	}
	b.geometries.TransformTo(wgs84)

	for i := 0; i < b.geometries.GeometryCount(); i++ {
		region, err := geometryToS2Region(b.geometries.Geometry(i))
		if err == nil {
			if !intersects(region.RectBound(), b.Bounds) {
				continue
			}
			f := newFeatureFromS2Region(region)
			f.SetFeatureID(b.b6IDs[i])
			for _, tag := range b.tags[i] {
				f.AddTag(tag)
			}
			for _, tag := range b.AddTags {
				f.AddTag(tag)
			}
			b.JoinTags.AddTags(b.originalIDs[i], f)
			err = b.Emit(f, i%b.Goroutines)
		}
		if err != nil {
			return err
		}
	}

	b.geometries.Destroy()
	b.geometries = gdal.Geometry{}
	wgs84.Destroy()
	b.originalIDs = b.originalIDs[0:0]
	b.b6IDs = b.b6IDs[0:0]
	b.tags = b.tags[0:0]
	return nil
}

type CopyTag struct {
	Field string
	Key   string
}

type Source struct {
	Filename      string
	Layer         string
	Bounds        s2.Rect
	Namespace     b6.Namespace
	IDField       string
	IDStrategy    IDStrategy
	CopyAllFields bool
	CopyTags      []CopyTag
	AddTags       []b6.Tag
	JoinTags      ingest.JoinTags
}

func newFeatureFromS2Region(r s2.Region) ingest.Feature {
	switch g := r.(type) {
	case s2.Point:
		return &ingest.GenericFeature{
			ID:   b6.FeatureIDInvalid,
			Tags: []b6.Tag{{Key: b6.PointTag, Value: b6.LatLng(s2.LatLngFromPoint(g))}},
		}
	case *s2.Polyline:
		f := &ingest.GenericFeature{}
		for i, p := range *g {
			f.ModifyOrAddTagAt(b6.Tag{b6.PathTag, b6.LatLng(s2.LatLngFromPoint(p))}, i)
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

func intersects(a s2.Rect, b s2.Rect) bool {
	for _, r := range []s2.Rect{a, b} {
		if !r.IsValid() || r.IsFull() {
			return true
		}
	}
	return a.Intersects(b)
}

type copyField struct {
	FieldName  string
	FieldIndex int
	Key        string
	Type       gdal.FieldType
}

func (c copyField) Value(f *gdal.Feature) (string, error) {
	if c.FieldIndex < 0 || !f.IsFieldSet(c.FieldIndex) {
		return "", nil
	}
	switch c.Type {
	case gdal.FT_String:
		return f.FieldAsString(c.FieldIndex), nil
	case gdal.FT_Integer:
		return strconv.Itoa(f.FieldAsInteger(c.FieldIndex)), nil
	case gdal.FT_Integer64:
		return fmt.Sprintf("%d", f.FieldAsInteger64(c.FieldIndex)), nil
	case gdal.FT_Real:
		return fmt.Sprintf("%f", f.FieldAsFloat64(c.FieldIndex)), nil
	case gdal.FT_Date, gdal.FT_Time, gdal.FT_DateTime:
		if t, ok := f.FieldAsDateTime(c.FieldIndex); ok {
			return t.String(), nil
		}
	}
	return "", fmt.Errorf("Can't convert field %s, type %v, to string", c.FieldName, c.Type)
}

type copyFields []copyField

func (c copyFields) Fill(f *gdal.Feature, tags []b6.Tag) ([]b6.Tag, error) {
	for _, cc := range c {
		if value, err := cc.Value(f); err == nil {
			tags = append(tags, b6.Tag{Key: cc.Key, Value: b6.StringExpression(value)})
		} else {
			return nil, err
		}
	}
	return tags, nil
}

// Returns fields to be copied to features, and the field to be used as the ID
func (s *Source) makeCopyFields(d gdal.FeatureDefinition) (copyFields, copyField, error) {
	cfs := make(copyFields, 0, len(s.CopyTags))
	id := copyField{FieldIndex: -1}
	for _, c := range s.CopyTags {
		i := d.FieldIndex(c.Field)
		if i < 0 {
			found := make([]string, d.FieldCount())
			for j := 0; j < d.FieldCount(); j++ {
				found[j] = d.FieldDefinition(j).Name()
			}
			return cfs, id, fmt.Errorf("No field named %q; found: %s", c.Field, strings.Join(found, ","))
		}
		d := d.FieldDefinition(i)
		cf := copyField{FieldName: d.Name(), FieldIndex: i, Key: c.Key, Type: d.Type()}
		cfs = append(cfs, cf)
	}

	if s.IDField != "" {
		i := d.FieldIndex(s.IDField)
		if i < 0 {
			return cfs, id, fmt.Errorf("No field named %q for ID", s.IDField)
		}
		d := d.FieldDefinition(i)
		id = copyField{FieldName: d.Name(), FieldIndex: i, Key: d.Name(), Type: d.Type()}
	}

	if s.CopyAllFields {
		for i := 0; i < d.FieldCount(); i++ {
			copied := i == id.FieldIndex
			if !copied {
				for _, cf := range cfs {
					if cf.FieldIndex == i {
						copied = true
						break
					}
				}
			}
			if !copied {
				d := d.FieldDefinition(i)
				cf := copyField{FieldName: d.Name(), FieldIndex: i, Key: d.Name(), Type: d.Type()}
				cfs = append(cfs, cf)
			}
		}
	}
	return cfs, id, nil
}

func shouldSkip(t gdal.GeometryType, options *ingest.ReadOptions) bool {
	tt, err := geometryTypeToFeatureType(t)
	if err != nil {
		log.Printf("%s", err)
		return true
	}
	switch tt {
	case b6.FeatureTypePoint:
		return options.SkipPoints
	case b6.FeatureTypePath:
		return options.SkipPaths
	case b6.FeatureTypeArea:
		return options.SkipAreas
	}
	return false
}

func (s *Source) Read(options ingest.ReadOptions, emit ingest.Emit, ctx context.Context) error {
	if options.Goroutines < 1 {
		options.Goroutines = 1
	}

	source := gdal.OpenDataSource(s.Filename, 0)
	defer source.Destroy()
	parallelised, wait := ingest.ParalleliseEmit(emit, options.Goroutines, ctx)

	names := make([]string, 0)
	layers := make([]gdal.Layer, 0)
	for i := 0; i < source.LayerCount(); i++ {
		layer := source.LayerByIndex(i)
		if s.Layer == "" || layer.Name() == s.Layer {
			layers = append(layers, layer)
		}
		names = append(names, layer.Name())
	}
	if len(layers) == 0 {
		if s.Layer != "" {
			return fmt.Errorf("No layer named %s, found %s", s.Layer, strings.Join(names, ","))
		} else {
			return fmt.Errorf("Input contains no readable layers")
		}
	}

	featureIndex := 0
	for _, layer := range layers {
		log.Printf("layer: %s", layer.Name())
		b := batchTransformer{
			Bounds:           s.Bounds,
			AddTags:          s.AddTags,
			JoinTags:         s.JoinTags,
			Goroutines:       options.Goroutines,
			Emit:             parallelised,
			SpatialReference: layer.SpatialReference(),
		}

		if shouldSkip(layer.Type(), &options) {
			continue
		}

		definition := layer.Definition()
		copyTags, copyID, err := s.makeCopyFields(definition)
		if err != nil {
			return err
		}
		for {
			feature := layer.NextFeature()
			if feature == nil {
				break
			}
			geometry := feature.StealGeometry()
			if shouldSkip(geometry.Type(), &options) {
				continue
			}
			var originalID string
			var b6ID b6.FeatureID
			if originalID, err = copyID.Value(feature); err == nil {
				var t b6.FeatureType
				if t, err = geometryTypeToFeatureType(geometry.Type()); err == nil {
					b6ID, err = s.IDStrategy(originalID, featureIndex, t, s.Namespace)
				}
			}
			if err != nil {
				return fmt.Errorf("can't make ID for feature from field %q with value %q: %s", copyID.FieldName, originalID, err)
			}
			tags, err := copyTags.Fill(feature, []b6.Tag{})
			if err != nil {
				return err
			}
			if err := b.TakeOwnershipAndTransform(geometry, originalID, b6ID, tags, ctx); err != nil {
				return err
			}
			feature.Destroy()
			featureIndex++
		}
		b.Flush(ctx)
	}
	return wait()
}
