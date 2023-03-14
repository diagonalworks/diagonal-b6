package shp

import (
	"fmt"
	"reflect"

	"diagonal.works/b6/geometry"

	"github.com/golang/geo/s2"
	"github.com/lukeroth/gdal"
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

func GDALGeometryToS2Region(g gdal.Geometry) (s2.Region, error) {
	switch g.Type() {
	case gdal.GT_Point:
		return geometryToS2Point(g), nil
	case gdal.GT_LineString:
		return geometryToS2Polyline(g), nil
	case gdal.GT_Polygon:
		return geometryToS2Polygon(g), nil
	case gdal.GT_MultiPolygon:
		return geometryToS2MultiPolygon(g), nil
	}
	return nil, fmt.Errorf("Can't convert geometry type %s", geometryTypeToString(g.Type()))
}

func ReadFeatures(filename string, attributes interface{}) ([]s2.Region, error) {
	t := reflect.TypeOf(attributes)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Slice {
		return nil, fmt.Errorf("Expected attributes to be a pointer to a slice")
	}
	t = t.Elem().Elem()

	source := gdal.OpenDataSource(filename, 0)
	defer source.Destroy()
	// TODO: Find a better way of error checking - this is the only way we can tell
	// the file exists, unless there's a was to convert a DataSet to a DataSource?
	if layers := source.LayerCount(); layers != 1 {
		return nil, fmt.Errorf("Expected 1 layer in %s, found %d", filename, layers)
	}

	wgs84 := gdal.CreateSpatialReference("")
	if err := wgs84.FromEPSG(EPSGCodeWGS84); err != nil {
		return nil, err
	}

	layer := source.LayerByIndex(0)
	definition := layer.Definition()
	shpFields := make([]int, 0, t.NumField())
	structFields := make([]int, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag, ok := field.Tag.Lookup("shp"); ok {
			if index := definition.FieldIndex(tag); index >= 0 {
				shpField := definition.FieldDefinition(index)
				switch shpField.Type() {
				case gdal.FT_Integer:
					if field.Type.Kind() != reflect.Int {
						return nil, fmt.Errorf("Field %s, expected %s, found %s", field.Name, reflect.Int, field.Type.Kind())
					}
				case gdal.FT_String:
					if field.Type.Kind() != reflect.String {
						return nil, fmt.Errorf("Field %s, expected %s, found %s", field.Name, reflect.String, field.Type.Kind())
					}
				default:
					return nil, fmt.Errorf("Can't handle type %s for %s", shpField.Type().Name(), tag)
				}
				structFields = append(structFields, i)
				shpFields = append(shpFields, index)
			}
		}
	}

	geometries := make([]s2.Region, 0)
	as := reflect.ValueOf(attributes)
	for {
		feature := layer.NextFeature()
		if feature == nil {
			break
		}
		geometry := feature.Geometry()
		geometry.TransformTo(wgs84)
		s2Geometry, err := GDALGeometryToS2Region(geometry)
		if err != nil {
			return nil, err
		}
		geometries = append(geometries, s2Geometry)
		a := reflect.New(t).Elem()
		for i := 0; i < len(structFields); i++ {
			shpField := definition.FieldDefinition(shpFields[i])
			switch shpField.Type() {
			case gdal.FT_Integer:
				a.Field(structFields[i]).Set(reflect.ValueOf(feature.FieldAsInteger(shpFields[i])))
			case gdal.FT_String:
				a.Field(structFields[i]).Set(reflect.ValueOf(feature.FieldAsString(shpFields[i])))
			default:
				panic("Type check failure")
			}
		}
		reflect.Indirect(as).Set(reflect.Append(as.Elem(), a))
	}
	return geometries, nil
}
