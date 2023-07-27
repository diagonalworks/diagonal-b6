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
	"find":             find,
	"find-points":      findPointFeatures,
	"find-paths":       findPathFeatures,
	"find-areas":       findAreaFeatures,
	"find-relations":   findRelationFeatures,
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
	"add":         add,
	"add-ints":    addInts,
	"clamp":       clamp,
	"percentiles": percentiles,
	"count":       count,
	// graph
	"reachable-area":     reachableArea,
	"reachable-points":   reachablePoints,
	"reachable":          reachableFeatures,
	"closest":            closestFeature,
	"closest-distance":   closestFeatureDistance,
	"paths-to-reach":     pathsToReachFeatures,
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
	"parse-geojson-file":    parseGeoJSONFile,
	"to-geojson":            toGeoJSON,
	"to-geojson-collection": toGeoJSONCollection,
	"import-geojson":        importGeoJSON,
	"import-geojson-file":   importGeoJSONFile,
	"geojson-areas":         geojsonAreas,
	"apply-to-point":        applyToPoint,
	"apply-to-path":         applyToPath,
	"apply-to-area":         applyToArea,
	"map-geometries":        mapGeometries,
	// change
	"add-tag":       addTag,
	"add-tags":      addTags,
	"remove-tag":    removeTag,
	"remove-tags":   removeTags,
	"merge-changes": mergeChanges,
	"with-change":   withChange,
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
	func(c api.Callable) func(*api.Context, interface{}) (interface{}, error) {
		return func(context *api.Context, v interface{}) (interface{}, error) {
			return api.Call1(context, v, c)
		}
	},
	func(c api.Callable) func(*api.Context, api.Pair) (interface{}, error) {
		return func(context *api.Context, pair api.Pair) (interface{}, error) {
			return api.Call1(context, pair, c)
		}
	},
	func(c api.Callable) func(*api.Context, interface{}) (bool, error) {
		return func(context *api.Context, v interface{}) (bool, error) {
			r, err := api.Call1(context, v, c)
			if err != nil {
				return false, err
			}
			return api.IsTrue(r), nil
		}
	},
	func(c api.Callable) func(*api.Context, b6.Geometry) (b6.Geometry, error) {
		return func(context *api.Context, g b6.Geometry) (b6.Geometry, error) {
			if result, err := api.Call1(context, g, c); result != nil {
				return result.(b6.Geometry), err
			} else {
				return nil, err
			}
		}
	},
	func(c api.Callable) func(*api.Context, b6.Point) (b6.Geometry, error) {
		return func(context *api.Context, p b6.Point) (b6.Geometry, error) {
			if result, err := api.Call1(context, p, c); result != nil {
				return result.(b6.Geometry), err
			} else {
				return nil, err
			}
		}
	},
	func(c api.Callable) func(*api.Context, b6.Path) (b6.Geometry, error) {
		return func(context *api.Context, p b6.Path) (b6.Geometry, error) {
			if result, err := api.Call1(context, p, c); result != nil {
				return result.(b6.Geometry), err
			} else {
				return nil, err
			}
		}
	},
	func(c api.Callable) func(*api.Context, b6.Area) (b6.Geometry, error) {
		return func(context *api.Context, a b6.Area) (b6.Geometry, error) {
			if result, err := api.Call1(context, a, c); result != nil {
				return result.(b6.Geometry), err
			} else {
				return nil, err
			}
		}
	},
	func(c api.Callable) func(*api.Context, b6.Geometry) (bool, error) {
		return func(context *api.Context, g b6.Geometry) (bool, error) {
			if result, err := api.Call1(context, g, c); result != nil {
				return result.(bool), err
			} else {
				return false, err
			}
		}
	},
	func(c api.Callable) func(*api.Context, b6.Feature) (bool, error) {
		return func(context *api.Context, f b6.Feature) (bool, error) {
			if result, err := api.Call1(context, f, c); result != nil {
				return api.IsTrue(result), err
			} else {
				return false, err
			}
		}
	},
	func(c api.Callable) func(*api.Context, b6.PointFeature) (bool, error) {
		return func(context *api.Context, f b6.PointFeature) (bool, error) {
			if result, err := api.Call1(context, f, c); result != nil {
				return api.IsTrue(result), err
			} else {
				return false, err
			}
		}
	},
	func(c api.Callable) func(*api.Context, b6.PathFeature) (bool, error) {
		return func(context *api.Context, f b6.PathFeature) (bool, error) {
			if result, err := api.Call1(context, f, c); result != nil {
				return api.IsTrue(result), err
			} else {
				return false, err
			}
		}
	},
	func(c api.Callable) func(*api.Context, b6.AreaFeature) (bool, error) {
		return func(context *api.Context, f b6.AreaFeature) (bool, error) {
			if result, err := api.Call1(context, f, c); result != nil {
				return api.IsTrue(result), err
			} else {
				return false, err
			}
		}
	},
	func(c api.Callable) func(*api.Context, b6.RelationFeature) (bool, error) {
		return func(context *api.Context, f b6.RelationFeature) (bool, error) {
			if result, err := api.Call1(context, f, c); err == nil && result != nil {
				return api.IsTrue(result), err
			} else {
				return false, err
			}
		}
	},
	func(c api.Callable) func(*api.Context, b6.Feature) (api.Collection, error) {
		return func(context *api.Context, f b6.Feature) (api.Collection, error) {
			result, err := api.Call1(context, f, c)
			if err == nil {
				if collection, ok := result.(api.Collection); ok {
					return collection, nil
				}
				err = fmt.Errorf("%s: expected a collection, found %T", c, result)
			}
			return nil, err
		}
	},
	func(c api.Callable) func(*api.Context, b6.Feature) (api.Pair, error) {
		return func(context *api.Context, f b6.Feature) (api.Pair, error) {
			result, err := api.Call1(context, f, c)
			if err == nil {
				if p, ok := result.(api.Pair); ok {
					return p, nil
				}
				err = fmt.Errorf("%s: expected a pair, found %T", c, result)
			}
			return nil, err
		}
	},
	func(c api.Callable) func(*api.Context, b6.Feature) (b6.Tag, error) {
		return func(context *api.Context, f b6.Feature) (b6.Tag, error) {
			result, err := api.Call1(context, f, c)
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
			return api.Call0(context, c)
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
	if t.NumIn() < 1 || t.In(0) != reflect.TypeOf(&api.Context{}) {
		return fmt.Errorf("%s: expected *Context as first argument", name)
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
