package functions

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
)

func (d Doc) ToGoLiteral() string {
	argNames := make([]string, len(d.ArgNames))
	for i := range d.ArgNames {
		argNames[i] = fmt.Sprintf("%q", d.ArgNames[i])
	}
	return fmt.Sprintf("Doc{Doc: %q, ArgNames: []string{%s}}", d.Doc, strings.Join(argNames, ","))
}

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
	"flatten":      flatten,
	"sum-by-key":   sumByKey,
	"take":         take,
	"top":          top,
	"histogram":    histogram,
	// search
	"find-feature":     findFeature,
	"find-path":        findPathFeature,
	"find-area":        findAreaFeature,
	"find-relation":    findRelationFeature,
	"find":             find,
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
	"all-tags":                  allTags,
	"matches":                   matches,
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
	"reachable-area":         reachableArea,
	"reachable":              reachable,
	"accessible-all":         accessibleAll,
	"accessible-routes":      accessibleRoutes,
	"filter-accessible":      filterAccessible,
	"closest":                closestFeature,
	"closest-distance":       closestFeatureDistance,
	"paths-to-reach":         pathsToReachFeatures,
	"connect":                connect,
	"connect-to-network":     connectToNetwork,
	"connect-to-network-all": connectToNetworkAll,
	// access
	"building-access": buildingAccess,
	// geometry
	"ll":                       ll,
	"collect-areas":            collectAreas,
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
	"id-to-relation-id": idToRelationID,
	"add-tag":           addTag,
	"add-tags":          addTags,
	"remove-tag":        removeTag,
	"remove-tags":       removeTags,
	"add-relation":      addRelation,
	"add-collection":    addCollection,
	"merge-changes":     mergeChanges,
	"with-change":       withChange,
	"changes-to-file":   changesToFile,
	"changes-from-file": changesFromFile,
	// materialise
	"materialise": materialise,
	// debug
	"debug-tokens":    debugTokens,
	"debug-all-query": debugAllQuery,
	// export
	"export-world": exportWorld,
}

func Functions() api.FunctionSymbols {
	return functions // Validated in init()
}

type Doc struct {
	Doc      string
	ArgNames []string
}

func FunctionDocs() map[string]Doc {
	return functionDocs
}

var functionAdaptors = []interface{}{
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
	func(c api.Callable) func(*api.Context, b6.Feature) (interface{}, error) {
		return func(context *api.Context, f b6.Feature) (interface{}, error) {
			return api.Call1(context, f, c)
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
	func(c api.Callable) func(*api.Context, b6.PhysicalFeature) (bool, error) {
		return func(context *api.Context, f b6.PhysicalFeature) (bool, error) {
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
	func(c api.Callable) func(*api.Context, b6.Feature) (b6.Collection[any, any], error) {
		return func(context *api.Context, f b6.Feature) (b6.Collection[any, any], error) {
			result, err := api.Call1(context, f, c)
			if err == nil {
				if collection, ok := result.(b6.Collection[any, any]); ok {
					return collection, nil
				}
				err = fmt.Errorf("%s: expected a collection, found %T", c, result)
			}
			return b6.Collection[any, any]{}, err
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

// The b6 VM is dynamically typed, allowing you to pass Collection[any, any]
// to a function that expects, eg, Collection[FeatureID, Feature]. These
// functions wrap a collection to match a different type, causing a runtime
// error if type conversion isn't possible.
var collectionAdaptors = []interface{}{
	b6.AdaptCollection[any, any],
	b6.AdaptCollection[any, b6.Area],
	b6.AdaptCollection[any, b6.AreaFeature],
	b6.AdaptCollection[any, b6.Feature],
	b6.AdaptCollection[any, b6.FeatureID],
	b6.AdaptCollection[any, b6.Geometry],
	b6.AdaptCollection[any, b6.Identifiable],
	b6.AdaptCollection[any, b6.Path],
	b6.AdaptCollection[any, b6.PathFeature],
	b6.AdaptCollection[any, b6.PhysicalFeature],
	b6.AdaptCollection[any, b6.Tag],
	b6.AdaptCollection[any, b6.UntypedCollection],
	b6.AdaptCollection[any, ingest.Change],
	b6.AdaptCollection[any, float64],
	b6.AdaptCollection[any, int],
	b6.AdaptCollection[b6.FeatureID, b6.FeatureID],
	b6.AdaptCollection[b6.FeatureID, b6.Area],
	b6.AdaptCollection[b6.FeatureID, b6.AreaFeature],
	b6.AdaptCollection[b6.FeatureID, b6.Feature],
	b6.AdaptCollection[b6.FeatureID, b6.Identifiable],
	b6.AdaptCollection[b6.FeatureID, b6.Path],
	b6.AdaptCollection[b6.FeatureID, b6.PathFeature],
	b6.AdaptCollection[b6.FeatureID, b6.Geometry],
	b6.AdaptCollection[b6.FeatureID, b6.PhysicalFeature],
	b6.AdaptCollection[b6.FeatureID, b6.Tag],
	b6.AdaptCollection[b6.FeatureID, string],
	b6.AdaptCollection[b6.Identifiable, string],
	b6.AdaptCollection[int, b6.Area],
	b6.AdaptCollection[int, b6.AreaFeature],
	b6.AdaptCollection[int, b6.Path],
	b6.AdaptCollection[int, b6.PathFeature],
	b6.AdaptCollection[int, b6.Geometry],
	b6.AdaptCollection[int, b6.PhysicalFeature],
}

var defaultAdaptors api.Adaptors

func Adaptors() api.Adaptors {
	return defaultAdaptors
}

var untypedCollectionInterface = reflect.TypeOf((*b6.UntypedCollection)(nil)).Elem()

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
			if _, ok := defaultAdaptors.Functions[t.In(i)]; !ok {
				return fmt.Errorf("%s: no adaptor for arg %d, %s", name, i, t.In(i))
			}
		} else if t.In(i).Implements(untypedCollectionInterface) && t.In(i) != untypedCollectionInterface {
			if _, ok := defaultAdaptors.Collections[t.In(i)]; !ok {
				return fmt.Errorf("%s: no adaptor for arg %d, %s", name, i, t.In(i))
			}
		}
	}
	for i := 0; i < t.NumOut(); i++ {
		if t.Out(i).Kind() == reflect.Func {
			if _, ok := defaultAdaptors.Functions[t.Out(i)]; !ok {
				return fmt.Errorf("%s: no convertor for result %d, %s", name, i, t.Out(i))
			}
		}
	}
	return nil
}

func NewContext(w b6.World) *api.Context {
	return &api.Context{
		World:           w,
		FunctionSymbols: Functions(),
		Adaptors:        Adaptors(),
		Context:         context.Background(),
	}
}

func makeFunctionAdaptor(adaptor interface{}) func(c api.Callable) reflect.Value {
	return func(c api.Callable) reflect.Value {
		w := reflect.ValueOf(adaptor)
		return w.Call([]reflect.Value{reflect.ValueOf(c)})[0]
	}
}

func makeCollectionAdaptor(adaptor interface{}) func(c b6.UntypedCollection) reflect.Value {
	return func(c b6.UntypedCollection) reflect.Value {
		w := reflect.ValueOf(adaptor)
		return w.Call([]reflect.Value{reflect.ValueOf(c)})[0]
	}
}

func init() {
	defaultAdaptors.Functions = make(map[reflect.Type]func(api.Callable) reflect.Value)
	for _, c := range functionAdaptors {
		t := reflect.TypeOf(c)
		defaultAdaptors.Functions[t.Out(0)] = makeFunctionAdaptor(c)
	}

	defaultAdaptors.Collections = make(map[reflect.Type]func(b6.UntypedCollection) reflect.Value)
	for _, f := range collectionAdaptors {
		t := reflect.TypeOf(f)
		defaultAdaptors.Collections[t.Out(0)] = makeCollectionAdaptor(f)
	}

	for name, f := range Functions() {
		if err := Validate(f, name); err != nil {
			panic(err)
		}
	}
}
