package functions

import (
	"context"
	"fmt"
	"reflect"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
)

var functions = api.FunctionSymbols{
	// map
	"map":          map_,
	"map-items":    mapItems,
	"map-parallel": mapParallel,
	"pair":         pair,
	"first":        first,
	"second":       second,
	// collections
	"collection":   collection,
	"count-values": countValues,
	"filter":       filter,
	"flattern":     flattern,
	"sum-by-key":   sumByKey,
	"take":         take,
	"top":          top,
	// search
	"find-feature":     findFeature,
	"find-point":       findPointFeature,
	"find-path":        findPathFeature,
	"find-area":        findAreaFeature,
	"find-relation":    findRelationFeature,
	"find":             Find,
	"find-points":      FindPointFeatures,
	"find-paths":       FindPathFeatures,
	"find-areas":       FindAreaFeatures,
	"find-relations":   FindRelationFeatures,
	"containing-areas": findAreasContainingPoints,
	"intersecting":     intersecting,
	"intersecting-cap": intersectingCap,
	"tagged":           tagged,
	"keyed":            keyed,
	"and":              and,
	"or":               or,
	"all":              all,
	"type-point":       typePoint,
	"type-path":        typePath,
	"type-area":        typeArea,
	"within":           within,
	"within-cap":       withinCap,
	// features
	"tag":                       tag,
	"value":                     value,
	"int-value":                 intValue,
	"float-value":               floatValue,
	"get":                       get,
	"get-string":                getString,
	"get-int":                   getInt,
	"get-float":                 getFloat,
	"has-key":                   hasKey,
	"all-tags":                  allTags,
	"count-tag-value":           countTagValue,
	"degree":                    pointDegree,
	"length":                    pathLengthMeters,
	"points":                    points,
	"point-features":            pointFeatures,
	"point-paths":               pointPaths,
	"sample-points":             samplePoints,
	"sample-points-along-paths": samplePointsAlongPaths,
	"join":                      join,
	"ordered-join":              orderedJoin,
	// s2
	"s2-points":   s2Points,
	"s2-covering": s2Covering,
	"s2-grid":     s2Grid,
	"s2-center":   s2Center,
	"s2-polygon":  s2Polygon,
	// math
	"gt":          gt,
	"divide":      divide,
	"divide-int":  divideInt,
	"add-ints":    addInts,
	"clamp":       clamp,
	"percentiles": percentiles,
	"count":       count,
	// graph
	"reachable-area":     ReachableArea,
	"reachable-points":   ReachablePoints,
	"reachable":          ReachableFeatures,
	"closest":            ClosestFeature,
	"closest-distance":   ClosestFeatureDistance,
	"paths-to-reach":     PathsToReachFeatures,
	"connect":            connect,
	"connect-to-network": connectToNetwork,
	// access
	"building-access": buildingAccess,
	// geometry
	"ll":                       ll,
	"distance-meters":          distanceMeters,
	"distance-to-point-meters": distanceToPointMeters,
	"interpolate":              interpolate,
	"area":                     areaArea,
	"rectangle-polygon":        rectanglePolygon,
	"cap-polygon":              capPolygon,
	"centroid":                 centroid,
	"sightline":                sightline,
	"snap-area-edges":          snapAreaEdges,
	"convex-hull":              convexHull,
	// tiles
	"tile-ids":     tileIDs,
	"tile-ids-hex": tileIDsHex,
	"tile-paths":   tilePaths,
	// collections
	"empty-points": emptyPointCollection,
	"add-point":    addPoint,
	// geojson
	"parse-geojson":         parseGeoJSON,
	"to-geojson":            toGeoJSON,
	"to-geojson-collection": toGeoJSONCollection,
	"import-geojson":        importGeoJSON,
	"import-geojson-file":   importGeoJSONFile,
	"geojson-areas":         geojsonAreas,
	"apply-to-point":        applyToPoint,
	"apply-to-path":         applyToPath,
	"apply-to-area":         applyToArea,
	"map-geometries":        MapGeometries,
	// change
	"add-tag":     addTag,
	"add-tags":    addTags,
	"remove-tag":  removeTag,
	"remove-tags": removeTags,
	"with-change": withChange,
	// debug
	"debug-tokens":    debugTokens,
	"debug-all-query": debugAllQuery,
	// export
	"export-world": exportWorld,
}

func Functions() api.FunctionSymbols {
	return functions // Validated in init()
}

var wrappers = []interface{}{
	func(c api.Callable) func(interface{}, *api.Context) (interface{}, error) {
		return func(v interface{}, context *api.Context) (interface{}, error) {
			return api.Call1(v, c, context)
		}
	},
	func(c api.Callable) func(api.Pair, *api.Context) (interface{}, error) {
		return func(pair api.Pair, context *api.Context) (interface{}, error) {
			return api.Call1(pair, c, context)
		}
	},
	func(c api.Callable) func(interface{}, *api.Context) (bool, error) {
		return func(v interface{}, context *api.Context) (bool, error) {
			r, err := api.Call1(v, c, context)
			if err != nil {
				return false, err
			}
			return api.IsTrue(r), nil
		}
	},
	func(c api.Callable) func(b6.Geometry, *api.Context) (b6.Geometry, error) {
		return func(g b6.Geometry, context *api.Context) (b6.Geometry, error) {
			if result, err := api.Call1(g, c, context); result != nil {
				return result.(b6.Geometry), err
			} else {
				return nil, err
			}
		}
	},
	func(c api.Callable) func(b6.Point, *api.Context) (b6.Geometry, error) {
		return func(p b6.Point, context *api.Context) (b6.Geometry, error) {
			if result, err := api.Call1(p, c, context); result != nil {
				return result.(b6.Geometry), err
			} else {
				return nil, err
			}
		}
	},
	func(c api.Callable) func(b6.Path, *api.Context) (b6.Geometry, error) {
		return func(p b6.Path, context *api.Context) (b6.Geometry, error) {
			if result, err := api.Call1(p, c, context); result != nil {
				return result.(b6.Geometry), err
			} else {
				return nil, err
			}
		}
	},
	func(c api.Callable) func(b6.Area, *api.Context) (b6.Geometry, error) {
		return func(a b6.Area, context *api.Context) (b6.Geometry, error) {
			if result, err := api.Call1(a, c, context); result != nil {
				return result.(b6.Geometry), err
			} else {
				return nil, err
			}
		}
	},
	func(c api.Callable) func(b6.Geometry, *api.Context) (bool, error) {
		return func(g b6.Geometry, context *api.Context) (bool, error) {
			if result, err := api.Call1(g, c, context); result != nil {
				return result.(bool), err
			} else {
				return false, err
			}
		}
	},
	func(c api.Callable) func(b6.Feature, *api.Context) (bool, error) {
		return func(f b6.Feature, context *api.Context) (bool, error) {
			if result, err := api.Call1(f, c, context); result != nil {
				return api.IsTrue(result), err
			} else {
				return false, err
			}
		}
	},
	func(c api.Callable) func(b6.PointFeature, *api.Context) (bool, error) {
		return func(f b6.PointFeature, context *api.Context) (bool, error) {
			if result, err := api.Call1(f, c, context); result != nil {
				return api.IsTrue(result), err
			} else {
				return false, err
			}
		}
	},
	func(c api.Callable) func(b6.PathFeature, *api.Context) (bool, error) {
		return func(f b6.PathFeature, context *api.Context) (bool, error) {
			if result, err := api.Call1(f, c, context); result != nil {
				return api.IsTrue(result), err
			} else {
				return false, err
			}
		}
	},
	func(c api.Callable) func(b6.AreaFeature, *api.Context) (bool, error) {
		return func(f b6.AreaFeature, context *api.Context) (bool, error) {
			if result, err := api.Call1(f, c, context); result != nil {
				return api.IsTrue(result), err
			} else {
				return false, err
			}
		}
	},
	func(c api.Callable) func(b6.RelationFeature, *api.Context) (bool, error) {
		return func(f b6.RelationFeature, context *api.Context) (bool, error) {
			if result, err := api.Call1(f, c, context); err == nil && result != nil {
				return api.IsTrue(result), err
			} else {
				return false, err
			}
		}
	},
	func(c api.Callable) func(b6.Feature, *api.Context) (api.Collection, error) {
		return func(f b6.Feature, context *api.Context) (api.Collection, error) {
			result, err := api.Call1(f, c, context)
			if err == nil {
				if collection, ok := result.(api.Collection); ok {
					return collection, nil
				}
				err = fmt.Errorf("%s: expected a collection, found %T", c, result)
			}
			return nil, err
		}
	},
	func(c api.Callable) func(b6.Feature, *api.Context) (api.Pair, error) {
		return func(f b6.Feature, context *api.Context) (api.Pair, error) {
			result, err := api.Call1(f, c, context)
			if err == nil {
				if p, ok := result.(api.Pair); ok {
					return p, nil
				}
				err = fmt.Errorf("%s: expected a pair, found %T", c, result)
			}
			return nil, err
		}
	},
	func(c api.Callable) func(b6.Feature, *api.Context) (b6.Tag, error) {
		return func(f b6.Feature, context *api.Context) (b6.Tag, error) {
			result, err := api.Call1(f, c, context)
			if err == nil {
				if tag, ok := result.(b6.Tag); ok {
					return tag, nil
				}
				err = fmt.Errorf("%s: expected a tag, found %T", c, result)
			}
			return b6.Tag{}, err
		}
	},
	func(c api.Callable) func(*api.Context) (interface{}, error) {
		return func(context *api.Context) (interface{}, error) {
			return api.Call0(c, context)
		}
	},
}

var wrappersByType api.FunctionWrappers

func Wrappers() api.FunctionWrappers {
	return wrappersByType
}

func makeWrapper(wrapper interface{}) func(c api.Callable) reflect.Value {
	return func(c api.Callable) reflect.Value {
		w := reflect.ValueOf(wrapper)
		return w.Call([]reflect.Value{reflect.ValueOf(c)})[0]
	}
}

func Validate(f interface{}, name string) error {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		panic(fmt.Sprintf("%s is not a function", t))
	}
	if t.NumIn() > api.MaxArgs {
		return fmt.Errorf("%s: has %d args, maximum allowed is %d", name, t.NumIn(), api.MaxArgs)
	}
	for i := 0; i < t.NumIn(); i++ {
		if t.In(i).Kind() == reflect.Func {
			if _, ok := wrappersByType[t.In(i)]; !ok {
				return fmt.Errorf("%s: no convertor for arg %d, %s", name, i, t.In(i))
			}
		}
	}
	for i := 0; i < t.NumOut(); i++ {
		if t.Out(i).Kind() == reflect.Func {
			if _, ok := wrappersByType[t.Out(i)]; !ok {
				return fmt.Errorf("%s: no convertor for result %d, %s", name, i, t.Out(i))
			}
		}
	}
	return nil
}

func NewContext(w b6.World) *api.Context {
	return &api.Context{
		World:            w,
		FunctionSymbols:  Functions(),
		FunctionWrappers: Wrappers(),
		Context:          context.Background(),
	}
}

func init() {
	wrappersByType = make(map[reflect.Type]func(api.Callable) reflect.Value)
	for _, c := range wrappers {
		t := reflect.TypeOf(c)
		wrappersByType[t.Out(0)] = makeWrapper(c)
	}

	for name, f := range Functions() {
		if err := Validate(f, name); err != nil {
			panic(err)
		}
	}
}
