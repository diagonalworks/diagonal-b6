---
sidebar_position: 1
---

# b6 API documentation
## Functions

  This is documentation generated from the `b6-api` binary written assuming
  you are interacting with it via the Python API.

  Below are all the functions, with their Python function name, and the
  corresponding argument names and types. Note that the types are the **b6 go**
  type definitions; not the python ones; nevertheless it is indicative of what
  type to expect to construct on the Python side.
  

### <tt>accessible_all</tt> 
```python title='Indicative Python type signature'
def accessible_all(origins, destinations, duration, options) -> FeatureIDFeatureIDCollection
```

Return the a collection of the features reachable from the given origins, within the given duration in seconds, that match the given query.
Keys of the collection are origins, values are reachable destinations.
Options are passed as tags containing the mode, and mode specific values. Examples include:
Walking, with the default speed of 4.5km/h:
mode=walk
Walking, a speed of 3km/h:
mode=walk, walk:speed=3.0
Transit at peak times:
mode=transit
Transit at off-peak times:
mode=transit, peak=no
Walking, accounting for elevation:
elevation=true (optional: elevation:uphill=2.0 elevation:downhill=1.2)
Walking, accounting for elevation, adding double the penalty for uphill:
elevation=true, elevation:uphill=2.0
Walking, with the resulting collection flipped such that keys are
destinations and values are origins. Useful for efficiency if you assume
symmetry, and the number of destinations is considerably smaller than the
number of origins:
mode=walk, flip=yes

#### Arguments

- `origins` of type [AnyIdentifiableCollection](#anyidentifiablecollection)
- `destinations` of type [Query](#query)
- `duration` of type `float64`
- `options` of type [AnyAnyCollection](#anyanycollection)

#### Returns
- [FeatureIDFeatureIDCollection](#featureidfeatureidcollection)

### <tt>accessible_routes</tt> 
```python title='Indicative Python type signature'
def accessible_routes(origin, destinations, duration, options) -> FeatureIDRouteCollection
```


#### Arguments

- `origin` of type [Identifiable](#identifiable)
- `destinations` of type [Query](#query)
- `duration` of type `float64`
- `options` of type [AnyAnyCollection](#anyanycollection)

#### Returns
- [FeatureIDRouteCollection](#featureidroutecollection)

### <tt>add</tt> 
```python title='Indicative Python type signature'
def add(a, b) -> Number
```

Return a added to b.

#### Arguments

- `a` of type [Number](#number)
- `b` of type [Number](#number)

#### Returns
- [Number](#number)

### <tt>add_collection</tt> 
```python title='Indicative Python type signature'
def add_collection(id, tags, collection) -> Change
```

Add a collection feature with the given id, tags and items.

#### Arguments

- `id` of type [CollectionID](#collectionid)
- `tags` of type [AnyTagCollection](#anytagcollection)
- `collection` of type [AnyAnyCollection](#anyanycollection)

#### Returns
- [Change](#change)

### <tt>add_expression</tt> 
```python title='Indicative Python type signature'
def add_expression(id, tags, expresson) -> Change
```

Add an expression feature with the given id, tags and expression.

#### Arguments

- `id` of type [FeatureID](#featureid)
- `tags` of type [AnyTagCollection](#anytagcollection)
- `expresson` of type [Expression](#expression)

#### Returns
- [Change](#change)

### <tt>add_ints</tt> 
```python title='Indicative Python type signature'
def add_ints(a, b) -> int
```

Deprecated.

#### Arguments

- `a` of type `int`
- `b` of type `int`

#### Returns
- `int`

### <tt>add_point</tt> 
```python title='Indicative Python type signature'
def add_point(point, id, tags) -> Change
```

Adds a point feature with the given id, tags and members.

#### Arguments

- `point` of type [Geometry](#geometry)
- `id` of type [FeatureID](#featureid)
- `tags` of type [AnyTagCollection](#anytagcollection)

#### Returns
- [Change](#change)

### <tt>add_relation</tt> 
```python title='Indicative Python type signature'
def add_relation(id, tags, members) -> Change
```

Add a relation feature with the given id, tags and members.

#### Arguments

- `id` of type [RelationID](#relationid)
- `tags` of type [AnyTagCollection](#anytagcollection)
- `members` of type [IdentifiableStringCollection](#identifiablestringcollection)

#### Returns
- [Change](#change)

### <tt>add_tag</tt> 
```python title='Indicative Python type signature'
def add_tag(id, tag) -> Change
```

Add the given tag to the given feature.

#### Arguments

- `id` of type [Identifiable](#identifiable)
- `tag` of type [Tag](#tag)

#### Returns
- [Change](#change)

### <tt>add_tags</tt> 
```python title='Indicative Python type signature'
def add_tags(collection) -> Change
```

Add the given tags to the given features.
The keys of the given collection specify the features to change, the
values provide the tag to be added.

#### Arguments

- `collection` of type [FeatureIDTagCollection](#featureidtagcollection)

#### Returns
- [Change](#change)

### <tt>add_world_with_change</tt> 
```python title='Indicative Python type signature'
def add_world_with_change(id, change) -> FeatureIDFeatureIDCollection
```


#### Arguments

- `id` of type [FeatureID](#featureid)
- `change` of type [Change](#change)

#### Returns
- [FeatureIDFeatureIDCollection](#featureidfeatureidcollection)

### <tt>all</tt> 
```python title='Indicative Python type signature'
def all() -> Query
```

Return a query that will match any feature.

#### Arguments


#### Returns
- [Query](#query)

### <tt>all_tags</tt> 
```python title='Indicative Python type signature'
def all_tags(id) -> IntTagCollection
```

Return a collection of all the tags on the given feature.
Keys are ordered integers from 0, values are tags.

#### Arguments

- `id` of type [Identifiable](#identifiable)

#### Returns
- [IntTagCollection](#inttagcollection)

### <tt>and</tt> 
```python title='Indicative Python type signature'
def and(a, b) -> Query
```

Return a query that will match features that match both given queries.

#### Arguments

- `a` of type [Query](#query)
- `b` of type [Query](#query)

#### Returns
- [Query](#query)

### <tt>apply_to_area</tt> 
```python title='Indicative Python type signature'
def apply_to_area(f) -> Callable
```

Wrap the given function such that it will only be called when passed an area.

#### Arguments

- `f` of type [Callable](#callable)

#### Returns
- [Callable](#callable)

### <tt>apply_to_path</tt> 
```python title='Indicative Python type signature'
def apply_to_path(f) -> Callable
```

Wrap the given function such that it will only be called when passed a path.

#### Arguments

- `f` of type [Callable](#callable)

#### Returns
- [Callable](#callable)

### <tt>apply_to_point</tt> 
```python title='Indicative Python type signature'
def apply_to_point(f) -> Callable
```

Wrap the given function such that it will only be called when passed a point.

#### Arguments

- `f` of type [Callable](#callable)

#### Returns
- [Callable](#callable)

### <tt>area</tt> 
```python title='Indicative Python type signature'
def area(area) -> float64
```

Return the area of the given polygon in m².

#### Arguments

- `area` of type [Area](#area)

#### Returns
- `float64`

### <tt>building_access</tt> 
```python title='Indicative Python type signature'
def building_access(origins, limit, mode) -> FeatureIDFeatureIDCollection
```

Deprecated. Use accessible.

#### Arguments

- `origins` of type [AnyFeatureCollection](#anyfeaturecollection)
- `limit` of type `float64`
- `mode` of type `string`

#### Returns
- [FeatureIDFeatureIDCollection](#featureidfeatureidcollection)

### <tt>call</tt> 
```python title='Indicative Python type signature'
def call(f, args) -> Any
```


#### Arguments

- `f` of type [Callable](#callable)
- `args` of type [Any](#any)

#### Returns
- [Any](#any)
#### Misc
 - [x] Function is _variadic_ (has a variable number of arguments.)

### <tt>cap_polygon</tt> 
```python title='Indicative Python type signature'
def cap_polygon(center, radius) -> Area
```

Return a polygon approximating a spherical cap with the given center and radius in meters.

#### Arguments

- `center` of type [Geometry](#geometry)
- `radius` of type `float64`

#### Returns
- [Area](#area)

### <tt>centroid</tt> 
```python title='Indicative Python type signature'
def centroid(geometry) -> Geometry
```

Return the centroid of the given geometry.
For multipolygons, we return the centroid of the convex hull formed from
the points of those polygons.

#### Arguments

- `geometry` of type [Geometry](#geometry)

#### Returns
- [Geometry](#geometry)

### <tt>changes_from_file</tt> 
```python title='Indicative Python type signature'
def changes_from_file(filename) -> Change
```

Return the changes contained in the given file.
As the file is read by the b6 server process, the filename it relative
to the filesystems it sees. Reading from files on cloud storage is
supported.

#### Arguments

- `filename` of type `string`

#### Returns
- [Change](#change)

### <tt>changes_to_file</tt> 
```python title='Indicative Python type signature'
def changes_to_file(filename) -> string
```

Export the changes that have been applied to the world to the given filename as yaml.
As the file is written by the b6 server process, the filename it relative
to the filesystems it sees. Writing files to cloud storage is
supported.

#### Arguments

- `filename` of type `string`

#### Returns
- `string`

### <tt>clamp</tt> 
```python title='Indicative Python type signature'
def clamp(v, low, high) -> int
```

Return the given value, unless it falls outside the given inclusive bounds, in which case return the boundary.

#### Arguments

- `v` of type `int`
- `low` of type `int`
- `high` of type `int`

#### Returns
- `int`

### <tt>closest</tt> 
```python title='Indicative Python type signature'
def closest(origin, options, distance, query) -> Feature
```

Return the closest feature from the given origin via the given mode, within the given distance in meters, matching the given query.
See accessible-all for options values.

#### Arguments

- `origin` of type [Feature](#feature)
- `options` of type [AnyAnyCollection](#anyanycollection)
- `distance` of type `float64`
- `query` of type [Query](#query)

#### Returns
- [Feature](#feature)

### <tt>closest_distance</tt> 
```python title='Indicative Python type signature'
def closest_distance(origin, options, distance, query) -> float64
```

Return the distance through the graph of the closest feature from the given origin via the given mode, within the given distance in meters, matching the given query.
See accessible-all for options values.

#### Arguments

- `origin` of type [Feature](#feature)
- `options` of type [AnyAnyCollection](#anyanycollection)
- `distance` of type `float64`
- `query` of type [Query](#query)

#### Returns
- `float64`

### <tt>collect_areas</tt> 
```python title='Indicative Python type signature'
def collect_areas(areas) -> Area
```

Return a single area containing all areas from the given collection.
If areas in the collection overlap, loops within the returned area
will overlap, which will likely cause undefined behaviour in many
functions.

#### Arguments

- `areas` of type [AnyAreaCollection](#anyareacollection)

#### Returns
- [Area](#area)

### <tt>collection</tt> 
```python title='Indicative Python type signature'
def collection(pairs) -> AnyAnyCollection
```

Return a collection of the given key value pairs.

#### Arguments

- `pairs` of type [Any](#any)

#### Returns
- [AnyAnyCollection](#anyanycollection)
#### Misc
 - [x] Function is _variadic_ (has a variable number of arguments.)

### <tt>connect</tt> 
```python title='Indicative Python type signature'
def connect(a, b) -> Change
```

Add a path that connects the two given points, if they're not already directly connected.

#### Arguments

- `a` of type [Feature](#feature)
- `b` of type [Feature](#feature)

#### Returns
- [Change](#change)

### <tt>connect_to_network</tt> 
```python title='Indicative Python type signature'
def connect_to_network(feature) -> Change
```

Add a path and point to connect given feature to the street network.
The street network is defined at the set of paths tagged #highway that
allow traversal of more than 500m. A point is added to the closest
network path at the projection of the origin point on that path, unless
that point is within 4m of an existing path point.

#### Arguments

- `feature` of type [Feature](#feature)

#### Returns
- [Change](#change)

### <tt>connect_to_network_all</tt> 
```python title='Indicative Python type signature'
def connect_to_network_all(features) -> Change
```

Add paths and points to connect the given collection of features to the
network. See connect-to-network for connection details.
More efficient than using map with connect-to-network, as the street
network is only computed once.

#### Arguments

- `features` of type [AnyFeatureIDCollection](#anyfeatureidcollection)

#### Returns
- [Change](#change)

### <tt>containing_areas</tt> 
```python title='Indicative Python type signature'
def containing_areas(points, q) -> FeatureIDAreaFeatureCollection
```


#### Arguments

- `points` of type [AnyFeatureCollection](#anyfeaturecollection)
- `q` of type [Query](#query)

#### Returns
- [FeatureIDAreaFeatureCollection](#featureidareafeaturecollection)

### <tt>convex_hull</tt> 
```python title='Indicative Python type signature'
def convex_hull(c) -> Area
```

Return the convex hull of the given geometries.

#### Arguments

- `c` of type [AnyGeometryCollection](#anygeometrycollection)

#### Returns
- [Area](#area)

### <tt>count</tt> 
```python title='Indicative Python type signature'
def count(collection) -> int
```

Return the number of items in the given collection.
The function will not evaluate and traverse the entire collection if it's possible to count
the collection efficiently.

#### Arguments

- `collection` of type [AnyAnyCollection](#anyanycollection)

#### Returns
- `int`

### <tt>count_keys</tt> 
```python title='Indicative Python type signature'
def count_keys(collection) -> AnyIntCollection
```

Return a collection of the number of occurances of each value in the given collection.

#### Arguments

- `collection` of type [AnyAnyCollection](#anyanycollection)

#### Returns
- [AnyIntCollection](#anyintcollection)

### <tt>count_tag_value</tt> 
```python title='Indicative Python type signature'
def count_tag_value(id, key) -> AnyIntCollection
```

Deprecated.

#### Arguments

- `id` of type [Identifiable](#identifiable)
- `key` of type `string`

#### Returns
- [AnyIntCollection](#anyintcollection)

### <tt>count_valid_ids</tt> 
```python title='Indicative Python type signature'
def count_valid_ids(collection) -> int
```

Return the number of valid feature IDs in the given collection.

#### Arguments

- `collection` of type [AnyIdentifiableCollection](#anyidentifiablecollection)

#### Returns
- `int`

### <tt>count_valid_keys</tt> 
```python title='Indicative Python type signature'
def count_valid_keys(collection) -> AnyIntCollection
```

Return a collection of the number of occurances of each valid value in the given collection.
Invalid values are not counted, but case the key to appear in the output.

#### Arguments

- `collection` of type [AnyAnyCollection](#anyanycollection)

#### Returns
- [AnyIntCollection](#anyintcollection)

### <tt>count_values</tt> 
```python title='Indicative Python type signature'
def count_values(collection) -> AnyIntCollection
```

Return a collection of the number of occurances of each value in the given collection.

#### Arguments

- `collection` of type [AnyAnyCollection](#anyanycollection)

#### Returns
- [AnyIntCollection](#anyintcollection)

### <tt>debug_all_query</tt> 
```python title='Indicative Python type signature'
def debug_all_query(token) -> Query
```

Deprecated.

#### Arguments

- `token` of type `string`

#### Returns
- [Query](#query)

### <tt>debug_tokens</tt> 
```python title='Indicative Python type signature'
def debug_tokens(id) -> IntStringCollection
```

Return the search index tokens generated for the given feature.
Intended for debugging use only.

#### Arguments

- `id` of type [Identifiable](#identifiable)

#### Returns
- [IntStringCollection](#intstringcollection)

### <tt>degree</tt> 
```python title='Indicative Python type signature'
def degree(point) -> int
```

Return the number of paths connected to the given point.
A single path will be counted twice if the point isn't at one of its
two ends - once in one direction, and once in the other.

#### Arguments

- `point` of type [Feature](#feature)

#### Returns
- `int`

### <tt>distance_meters</tt> 
```python title='Indicative Python type signature'
def distance_meters(a, b) -> float64
```

Return the distance in meters between the given points.

#### Arguments

- `a` of type [Geometry](#geometry)
- `b` of type [Geometry](#geometry)

#### Returns
- `float64`

### <tt>distance_to_point_meters</tt> 
```python title='Indicative Python type signature'
def distance_to_point_meters(path, point) -> float64
```

Return the distance in meters between the given path, and the project of the give point onto it.

#### Arguments

- `path` of type [Geometry](#geometry)
- `point` of type [Geometry](#geometry)

#### Returns
- `float64`

### <tt>divide</tt> 
```python title='Indicative Python type signature'
def divide(a, b) -> Number
```

Return a divided by b.

#### Arguments

- `a` of type [Number](#number)
- `b` of type [Number](#number)

#### Returns
- [Number](#number)

### <tt>divide_int</tt> 
```python title='Indicative Python type signature'
def divide_int(a, b) -> float64
```

Deprecated.

#### Arguments

- `a` of type `int`
- `b` of type `float64`

#### Returns
- `float64`

### <tt>entrance_approach</tt> 
```python title='Indicative Python type signature'
def entrance_approach(area) -> Geometry
```


#### Arguments

- `area` of type [AreaFeature](#areafeature)

#### Returns
- [Geometry](#geometry)

### <tt>evaluate_feature</tt> 
```python title='Indicative Python type signature'
def evaluate_feature(id) -> Any
```


#### Arguments

- `id` of type [FeatureID](#featureid)

#### Returns
- [Any](#any)

### <tt>export_world</tt> 
```python title='Indicative Python type signature'
def export_world(filename) -> int
```

Write the current world to the given filename in the b6 compact index format.
As the file is written by the b6 server process, the filename it relative
to the filesystems it sees. Writing files to cloud storage is
supported.

#### Arguments

- `filename` of type `string`

#### Returns
- `int`

### <tt>filter</tt> 
```python title='Indicative Python type signature'
def filter(collection, function) -> AnyAnyCollection
```

Return a collection of the items of the given collection for which the value of the given function applied to each value is true.

#### Arguments

- `collection` of type [AnyAnyCollection](#anyanycollection)
- `function` of type [Callable](#callable)

#### Returns
- [AnyAnyCollection](#anyanycollection)

### <tt>filter_accessible</tt> 
```python title='Indicative Python type signature'
def filter_accessible(collection, filter) -> FeatureIDFeatureIDCollection
```

Return a collection containing only the values of the given collection that match the given query.
If no values for a key match the query, emit a single invalid feature ID
for that key, allowing callers to count the number of keys with no valid
values.
Keys are taken from the given collection.

#### Arguments

- `collection` of type [FeatureIDFeatureIDCollection](#featureidfeatureidcollection)
- `filter` of type [Query](#query)

#### Returns
- [FeatureIDFeatureIDCollection](#featureidfeatureidcollection)

### <tt>find</tt> 
```python title='Indicative Python type signature'
def find(query) -> FeatureIDFeatureCollection
```

Return a collection of the features present in the world that match the given query.
Keys are IDs, and values are features.

#### Arguments

- `query` of type [Query](#query)

#### Returns
- [FeatureIDFeatureCollection](#featureidfeaturecollection)

### <tt>find_area</tt> 
```python title='Indicative Python type signature'
def find_area(id) -> AreaFeature
```

Return the area feature with the given ID.

#### Arguments

- `id` of type [FeatureID](#featureid)

#### Returns
- [AreaFeature](#areafeature)

### <tt>find_areas</tt> 
```python title='Indicative Python type signature'
def find_areas(query) -> FeatureIDAreaFeatureCollection
```

Return a collection of the area features present in the world that match the given query.
Keys are IDs, and values are features.

#### Arguments

- `query` of type [Query](#query)

#### Returns
- [FeatureIDAreaFeatureCollection](#featureidareafeaturecollection)

### <tt>find_collection</tt> 
```python title='Indicative Python type signature'
def find_collection(id) -> CollectionFeature
```

Return the collection feature with the given ID.

#### Arguments

- `id` of type [FeatureID](#featureid)

#### Returns
- [CollectionFeature](#collectionfeature)

### <tt>find_feature</tt> 
```python title='Indicative Python type signature'
def find_feature(id) -> Feature
```

Return the feature with the given ID.

#### Arguments

- `id` of type [FeatureID](#featureid)

#### Returns
- [Feature](#feature)

### <tt>find_relation</tt> 
```python title='Indicative Python type signature'
def find_relation(id) -> RelationFeature
```

Return the relation feature with the given ID.

#### Arguments

- `id` of type [FeatureID](#featureid)

#### Returns
- [RelationFeature](#relationfeature)

### <tt>find_relations</tt> 
```python title='Indicative Python type signature'
def find_relations(query) -> FeatureIDRelationFeatureCollection
```

Return a collection of the relation features present in the world that match the given query.
Keys are IDs, and values are features.

#### Arguments

- `query` of type [Query](#query)

#### Returns
- [FeatureIDRelationFeatureCollection](#featureidrelationfeaturecollection)

### <tt>first</tt> 
```python title='Indicative Python type signature'
def first(pair) -> Any
```

Return the first value of the given pair.

#### Arguments

- `pair` of type [Pair](#pair)

#### Returns
- [Any](#any)

### <tt>flatten</tt> 
```python title='Indicative Python type signature'
def flatten(collection) -> AnyAnyCollection
```

Return a collection with keys and values taken from the collections that form the values of the given collection.

#### Arguments

- `collection` of type [AnyAnyAnyCollectionCollection](#anyanyanycollectioncollection)

#### Returns
- [AnyAnyCollection](#anyanycollection)

### <tt>float_value</tt> 
```python title='Indicative Python type signature'
def float_value(tag) -> float64
```

Return the value of the given tag as a float.
Propagates error if the value isn't a valid float.

#### Arguments

- `tag` of type [Tag](#tag)

#### Returns
- `float64`

### <tt>geojson_areas</tt> 
```python title='Indicative Python type signature'
def geojson_areas(g) -> IntAreaCollection
```

Return the areas present in the given geojson.

#### Arguments

- `g` of type [GeoJSON](#geojson)

#### Returns
- [IntAreaCollection](#intareacollection)

### <tt>get</tt> 
```python title='Indicative Python type signature'
def get(id, key) -> Tag
```

Return the tag with the given key on the given feature.
Returns a tag. To return the string value of a tag, use get-string.

#### Arguments

- `id` of type [Identifiable](#identifiable)
- `key` of type `string`

#### Returns
- [Tag](#tag)

### <tt>get_float</tt> 
```python title='Indicative Python type signature'
def get_float(id, key) -> float64
```

Return the value of tag with the given key on the given feature as a float.
Returns error if there isn't a feature with that id, a tag with that key, or if the value isn't a valid float.

#### Arguments

- `id` of type [Identifiable](#identifiable)
- `key` of type `string`

#### Returns
- `float64`

### <tt>get_int</tt> 
```python title='Indicative Python type signature'
def get_int(id, key) -> int
```

Return the value of tag with the given key on the given feature as an integer.
Returns error if there isn't a feature with that id, a tag with that key, or if the value isn't a valid integer.

#### Arguments

- `id` of type [Identifiable](#identifiable)
- `key` of type `string`

#### Returns
- `int`

### <tt>get_string</tt> 
```python title='Indicative Python type signature'
def get_string(id, key) -> string
```

Return the value of tag with the given key on the given feature as a string.
Returns an empty string if there isn't a tag with that key.

#### Arguments

- `id` of type [Identifiable](#identifiable)
- `key` of type `string`

#### Returns
- `string`

### <tt>gt</tt> 
```python title='Indicative Python type signature'
def gt(a, b) -> bool
```

Return true if a is greater than b.

#### Arguments

- `a` of type [Any](#any)
- `b` of type [Any](#any)

#### Returns
- [bool](#bool)

### <tt>histogram</tt> 
```python title='Indicative Python type signature'
def histogram(collection) -> Change
```

Return a change that adds a histogram for the given collection.

#### Arguments

- `collection` of type [AnyAnyCollection](#anyanycollection)

#### Returns
- [Change](#change)

### <tt>histogram_swatch</tt> 
```python title='Indicative Python type signature'
def histogram_swatch(collection) -> Change
```

Return a change that adds a histogram with only colour swatches for the given collection.

#### Arguments

- `collection` of type [AnyAnyCollection](#anyanycollection)

#### Returns
- [Change](#change)

### <tt>histogram_swatch_with_id</tt> 
```python title='Indicative Python type signature'
def histogram_swatch_with_id(collection, id) -> Change
```

Return a change that adds a histogram with only colour swatches for the given collection.

#### Arguments

- `collection` of type [AnyAnyCollection](#anyanycollection)
- `id` of type [CollectionID](#collectionid)

#### Returns
- [Change](#change)

### <tt>histogram_with_id</tt> 
```python title='Indicative Python type signature'
def histogram_with_id(collection, id) -> Change
```

Return a change that adds a histogram for the given collection with the given ID.

#### Arguments

- `collection` of type [AnyAnyCollection](#anyanycollection)
- `id` of type [CollectionID](#collectionid)

#### Returns
- [Change](#change)

### <tt>id_to_relation_id</tt> 
```python title='Indicative Python type signature'
def id_to_relation_id(namespace, id) -> FeatureID
```

Deprecated.

#### Arguments

- `namespace` of type `string`
- `id` of type [Identifiable](#identifiable)

#### Returns
- [FeatureID](#featureid)

### <tt>import_geojson</tt> 
```python title='Indicative Python type signature'
def import_geojson(features, namespace) -> Change
```

Add features from the given geojson to the world.
IDs are formed from the given namespace, and the index of the feature
within the geojson collection (or 0, if a single feature is used).

#### Arguments

- `features` of type [GeoJSON](#geojson)
- `namespace` of type `string`

#### Returns
- [Change](#change)

### <tt>import_geojson_file</tt> 
```python title='Indicative Python type signature'
def import_geojson_file(filename, namespace) -> Change
```

Add features from the given geojson file to the world.
IDs are formed from the given namespace, and the index of the feature
within the geojson collection (or 0, if a single feature is used).
As the file is read by the b6 server process, the filename it relative
to the filesystems it sees. Reading from files on cloud storage is
supported.

#### Arguments

- `filename` of type `string`
- `namespace` of type `string`

#### Returns
- [Change](#change)

### <tt>int_value</tt> 
```python title='Indicative Python type signature'
def int_value(tag) -> int
```

Return the value of the given tag as an integer.
Propagates error if the value isn't a valid integer.

#### Arguments

- `tag` of type [Tag](#tag)

#### Returns
- `int`

### <tt>interpolate</tt> 
```python title='Indicative Python type signature'
def interpolate(path, fraction) -> Geometry
```

Return the point at the given fraction along the given path.

#### Arguments

- `path` of type [Geometry](#geometry)
- `fraction` of type `float64`

#### Returns
- [Geometry](#geometry)

### <tt>intersecting</tt> 
```python title='Indicative Python type signature'
def intersecting(geometry) -> Query
```

Return a query that will match features that intersect the given geometry.

#### Arguments

- `geometry` of type [Geometry](#geometry)

#### Returns
- [Query](#query)

### <tt>intersecting_cap</tt> 
```python title='Indicative Python type signature'
def intersecting_cap(center, radius) -> Query
```

Return a query that will match features that intersect a spherical cap centred on the given point, with the given radius in meters.

#### Arguments

- `center` of type [Geometry](#geometry)
- `radius` of type `float64`

#### Returns
- [Query](#query)

### <tt>join</tt> 
```python title='Indicative Python type signature'
def join(pathA, pathB) -> Geometry
```

Return a path formed from the points of the two given paths, in the order they occur in those paths.

#### Arguments

- `pathA` of type [Geometry](#geometry)
- `pathB` of type [Geometry](#geometry)

#### Returns
- [Geometry](#geometry)

### <tt>join_missing</tt> 
```python title='Indicative Python type signature'
def join_missing(base, joined) -> AnyAnyCollection
```


#### Arguments

- `base` of type [AnyAnyCollection](#anyanycollection)
- `joined` of type [AnyAnyCollection](#anyanycollection)

#### Returns
- [AnyAnyCollection](#anyanycollection)

### <tt>keyed</tt> 
```python title='Indicative Python type signature'
def keyed(key) -> Query
```

Return a query that will match features tagged with the given key independent of value.

#### Arguments

- `key` of type `string`

#### Returns
- [Query](#query)

### <tt>length</tt> 
```python title='Indicative Python type signature'
def length(path) -> float64
```

Return the length of the given path in meters.

#### Arguments

- `path` of type [Geometry](#geometry)

#### Returns
- `float64`

### <tt>list_feature</tt> 
```python title='Indicative Python type signature'
def list_feature(id) -> AnyAnyCollection
```


#### Arguments

- `id` of type [CollectionID](#collectionid)

#### Returns
- [AnyAnyCollection](#anyanycollection)

### <tt>ll</tt> 
```python title='Indicative Python type signature'
def ll(lat, lng) -> Geometry
```

Return a point at the given latitude and longitude, specified in degrees.

#### Arguments

- `lat` of type `float64`
- `lng` of type `float64`

#### Returns
- [Geometry](#geometry)

### <tt>map</tt> 
```python title='Indicative Python type signature'
def map(collection, function) -> AnyAnyCollection
```

Return a collection with the result of applying the given function to each value.
Keys are unmodified.

#### Arguments

- `collection` of type [AnyAnyCollection](#anyanycollection)
- `function` of type [Callable](#callable)

#### Returns
- [AnyAnyCollection](#anyanycollection)

### <tt>map_geometries</tt> 
```python title='Indicative Python type signature'
def map_geometries(g, f) -> GeoJSON
```

Return a geojson representing the result of applying the given function to each geometry in the given geojson.

#### Arguments

- `g` of type [GeoJSON](#geojson)
- `f` of type [Callable](#callable)

#### Returns
- [GeoJSON](#geojson)

### <tt>map_items</tt> 
```python title='Indicative Python type signature'
def map_items(collection, function) -> AnyAnyCollection
```

Return a collection of the result of applying the given function to each pair(key, value).
Keys are unmodified.

#### Arguments

- `collection` of type [AnyAnyCollection](#anyanycollection)
- `function` of type [Callable](#callable)

#### Returns
- [AnyAnyCollection](#anyanycollection)

### <tt>map_parallel</tt> 
```python title='Indicative Python type signature'
def map_parallel(collection, function) -> AnyAnyCollection
```

Return a collection with the result of applying the given function to each value.
Keys are unmodified, and function application occurs in parallel, bounded
by the number of CPU cores allocated to b6.

#### Arguments

- `collection` of type [AnyAnyCollection](#anyanycollection)
- `function` of type [Callable](#callable)

#### Returns
- [AnyAnyCollection](#anyanycollection)

### <tt>matches</tt> 
```python title='Indicative Python type signature'
def matches(id, query) -> bool
```

Return true if the given feature matches the given query.

#### Arguments

- `id` of type [Identifiable](#identifiable)
- `query` of type [Query](#query)

#### Returns
- [bool](#bool)

### <tt>materialise</tt> 
```python title='Indicative Python type signature'
def materialise(id, function) -> Change
```

Return a change that adds a collection feature to the world with the given ID, containing the result of calling the given function.
The given function isn't passed any arguments.
Also adds an expression feature (with the same namespace and value)
representing the given function.

#### Arguments

- `id` of type [CollectionID](#collectionid)
- `function` of type [Callable](#callable)

#### Returns
- [Change](#change)

### <tt>materialise_map</tt> 
```python title='Indicative Python type signature'
def materialise_map(collection, id, function) -> Change
```


#### Arguments

- `collection` of type [AnyFeatureCollection](#anyfeaturecollection)
- `id` of type [CollectionID](#collectionid)
- `function` of type [Callable](#callable)

#### Returns
- [Change](#change)

### <tt>merge_changes</tt> 
```python title='Indicative Python type signature'
def merge_changes(collection) -> Change
```

Return a change that will apply all the changes in the given collection.
Changes are applied transactionally. If the application of one change
fails (for example, because it includes a path that references a missing
point), then no changes will be applied.

#### Arguments

- `collection` of type [AnyChangeCollection](#anychangecollection)

#### Returns
- [Change](#change)

### <tt>or</tt> 
```python title='Indicative Python type signature'
def or(a, b) -> Query
```

Return a query that will match features that match either of the given queries.

#### Arguments

- `a` of type [Query](#query)
- `b` of type [Query](#query)

#### Returns
- [Query](#query)

### <tt>ordered_join</tt> 
```python title='Indicative Python type signature'
def ordered_join(pathA, pathB) -> Geometry
```

Returns a path formed by joining the two given paths.
If necessary to maintain consistency, the order of points is reversed,
determined by which points are shared between the paths. Returns an error
if no endpoints are shared.

#### Arguments

- `pathA` of type [Geometry](#geometry)
- `pathB` of type [Geometry](#geometry)

#### Returns
- [Geometry](#geometry)

### <tt>pair</tt> 
```python title='Indicative Python type signature'
def pair(first, second) -> Pair
```

Return a pair containing the given values.

#### Arguments

- `first` of type [Any](#any)
- `second` of type [Any](#any)

#### Returns
- [Pair](#pair)

### <tt>parse_geojson</tt> 
```python title='Indicative Python type signature'
def parse_geojson(s) -> GeoJSON
```

Return the geojson represented by the given string.

#### Arguments

- `s` of type `string`

#### Returns
- [GeoJSON](#geojson)

### <tt>parse_geojson_file</tt> 
```python title='Indicative Python type signature'
def parse_geojson_file(filename) -> GeoJSON
```

Return the geojson contained in the given file.
As the file is read by the b6 server process, the filename it relative
to the filesystems it sees. Reading from files on cloud storage is
supported.

#### Arguments

- `filename` of type `string`

#### Returns
- [GeoJSON](#geojson)

### <tt>paths_to_reach</tt> 
```python title='Indicative Python type signature'
def paths_to_reach(origin, options, distance, query) -> FeatureIDIntCollection
```

Return a collection of the paths used to reach all features matching the given query from the given origin via the given mode, within the given distance in meters.
Keys are the paths used, values are the number of times that path was used during traversal.
See accessible-all for options values.

#### Arguments

- `origin` of type [Feature](#feature)
- `options` of type [AnyAnyCollection](#anyanycollection)
- `distance` of type `float64`
- `query` of type [Query](#query)

#### Returns
- [FeatureIDIntCollection](#featureidintcollection)

### <tt>percentiles</tt> 
```python title='Indicative Python type signature'
def percentiles(collection) -> AnyFloat64Collection
```

Return a collection where values represent the perentile of the corresponding value in the given collection.
The returned collection is ordered by percentile, with keys drawn from the
given collection.

#### Arguments

- `collection` of type [AnyFloat64Collection](#anyfloat64collection)

#### Returns
- [AnyFloat64Collection](#anyfloat64collection)

### <tt>point_features</tt> 
```python title='Indicative Python type signature'
def point_features(f) -> FeatureIDPhysicalFeatureCollection
```

Return a collection of the point features referenced by the given feature.
Keys are ids of the respective value, values are point features. Area
features return the points referenced by their path features.

#### Arguments

- `f` of type [Feature](#feature)

#### Returns
- [FeatureIDPhysicalFeatureCollection](#featureidphysicalfeaturecollection)

### <tt>point_paths</tt> 
```python title='Indicative Python type signature'
def point_paths(id) -> FeatureIDPhysicalFeatureCollection
```

Return a collection of the path features referencing the given point.
Keys are the ids of the respective paths.

#### Arguments

- `id` of type [Identifiable](#identifiable)

#### Returns
- [FeatureIDPhysicalFeatureCollection](#featureidphysicalfeaturecollection)

### <tt>points</tt> 
```python title='Indicative Python type signature'
def points(geometry) -> IntGeometryCollection
```

Return a collection of the points of the given geometry.
Keys are ordered integers from 0, values are points.

#### Arguments

- `geometry` of type [Geometry](#geometry)

#### Returns
- [IntGeometryCollection](#intgeometrycollection)

### <tt>reachable</tt> 
```python title='Indicative Python type signature'
def reachable(origin, options, distance, query) -> FeatureIDFeatureCollection
```

Return the a collection of the features reachable from the given origin via the given mode, within the given distance in meters, that match the given query.
See accessible-all for options values.
Deprecated. Use accessible-all.

#### Arguments

- `origin` of type [Feature](#feature)
- `options` of type [AnyAnyCollection](#anyanycollection)
- `distance` of type `float64`
- `query` of type [Query](#query)

#### Returns
- [FeatureIDFeatureCollection](#featureidfeaturecollection)

### <tt>reachable_area</tt> 
```python title='Indicative Python type signature'
def reachable_area(origin, options, distance) -> float64
```

Return the area formed by the convex hull of the features matching the given query reachable from the given origin via the given mode specified in options, within the given distance in meters.
See accessible-all for options values.

#### Arguments

- `origin` of type [Feature](#feature)
- `options` of type [AnyAnyCollection](#anyanycollection)
- `distance` of type `float64`

#### Returns
- `float64`

### <tt>rectangle_polygon</tt> 
```python title='Indicative Python type signature'
def rectangle_polygon(a, b) -> Area
```

Return a rectangle polygon with the given top left and bottom right points.

#### Arguments

- `a` of type [Geometry](#geometry)
- `b` of type [Geometry](#geometry)

#### Returns
- [Area](#area)

### <tt>remove_tag</tt> 
```python title='Indicative Python type signature'
def remove_tag(id, key) -> Change
```

Remove the tag with the given key from the given feature.

#### Arguments

- `id` of type [Identifiable](#identifiable)
- `key` of type `string`

#### Returns
- [Change](#change)

### <tt>remove_tags</tt> 
```python title='Indicative Python type signature'
def remove_tags(collection) -> Change
```

Remove the given tags from the given features.
The keys of the given collection specify the features to change, the
values provide the key of the tag to be removed.

#### Arguments

- `collection` of type [FeatureIDStringCollection](#featureidstringcollection)

#### Returns
- [Change](#change)

### <tt>s2_center</tt> 
```python title='Indicative Python type signature'
def s2_center(token) -> Geometry
```

Return a collection the center of the s2 cell with the given token.

#### Arguments

- `token` of type `string`

#### Returns
- [Geometry](#geometry)

### <tt>s2_covering</tt> 
```python title='Indicative Python type signature'
def s2_covering(area, minLevel, maxLevel) -> IntStringCollection
```

Return a collection of of s2 cells tokens that cover the given area at the given level.

#### Arguments

- `area` of type [Area](#area)
- `minLevel` of type `int`
- `maxLevel` of type `int`

#### Returns
- [IntStringCollection](#intstringcollection)

### <tt>s2_grid</tt> 
```python title='Indicative Python type signature'
def s2_grid(area, level) -> IntStringCollection
```

Return a collection of points representing the centroids of s2 cells that cover the given area at the given level.

#### Arguments

- `area` of type [Area](#area)
- `level` of type `int`

#### Returns
- [IntStringCollection](#intstringcollection)

### <tt>s2_points</tt> 
```python title='Indicative Python type signature'
def s2_points(area, minLevel, maxLevel) -> StringGeometryCollection
```

Return a collection of points representing the centroids of s2 cells that cover the given area between the given levels.

#### Arguments

- `area` of type [Area](#area)
- `minLevel` of type `int`
- `maxLevel` of type `int`

#### Returns
- [StringGeometryCollection](#stringgeometrycollection)

### <tt>s2_polygon</tt> 
```python title='Indicative Python type signature'
def s2_polygon(token) -> Area
```

Return the bounding area of the s2 cell with the given token.

#### Arguments

- `token` of type `string`

#### Returns
- [Area](#area)

### <tt>sample_points</tt> 
```python title='Indicative Python type signature'
def sample_points(path, distanceMeters) -> IntGeometryCollection
```

Return a collection of points along the given path, with the given distance in meters between them.
Keys are ordered integers from 0, values are points.

#### Arguments

- `path` of type [Geometry](#geometry)
- `distanceMeters` of type `float64`

#### Returns
- [IntGeometryCollection](#intgeometrycollection)

### <tt>sample_points_along_paths</tt> 
```python title='Indicative Python type signature'
def sample_points_along_paths(paths, distanceMeters) -> IntGeometryCollection
```

Return a collection of points along the given paths, with the given distance in meters between them.
Keys are the id of the respective path, values are points.

#### Arguments

- `paths` of type [FeatureIDGeometryCollection](#featureidgeometrycollection)
- `distanceMeters` of type `float64`

#### Returns
- [IntGeometryCollection](#intgeometrycollection)

### <tt>second</tt> 
```python title='Indicative Python type signature'
def second(pair) -> Any
```

Return the second value of the given pair.

#### Arguments

- `pair` of type [Pair](#pair)

#### Returns
- [Any](#any)

### <tt>sightline</tt> 
```python title='Indicative Python type signature'
def sightline(from, radius) -> Area
```


#### Arguments

- `from` of type [Geometry](#geometry)
- `radius` of type `float64`

#### Returns
- [Area](#area)

### <tt>snap_area_edges</tt> 
```python title='Indicative Python type signature'
def snap_area_edges(area, query, threshold) -> Area
```

Return an area formed by projecting the edges of the given polygon onto the paths present in the world matching the given query.
Paths beyond the given threshold in meters are ignored.

#### Arguments

- `area` of type [Area](#area)
- `query` of type [Query](#query)
- `threshold` of type `float64`

#### Returns
- [Area](#area)

### <tt>sum</tt> 
```python title='Indicative Python type signature'
def sum(collection) -> int
```

Return the sum of all values in a given collection.

#### Arguments

- `collection` of type [AnyIntCollection](#anyintcollection)

#### Returns
- `int`

### <tt>sum_by_key</tt> 
```python title='Indicative Python type signature'
def sum_by_key(c) -> AnyIntCollection
```

Return a collection of the result of summing the values of each item with the same key.
Requires values to be integers.

#### Arguments

- `c` of type [AnyIntCollection](#anyintcollection)

#### Returns
- [AnyIntCollection](#anyintcollection)

### <tt>tag</tt> 
```python title='Indicative Python type signature'
def tag(key, value) -> Tag
```

Return a tag with the given key and value.

#### Arguments

- `key` of type `string`
- `value` of type `string`

#### Returns
- [Tag](#tag)

### <tt>tagged</tt> 
```python title='Indicative Python type signature'
def tagged(key, value) -> Query
```

Return a query that will match features tagged with the given key and value.

#### Arguments

- `key` of type `string`
- `value` of type `string`

#### Returns
- [Query](#query)

### <tt>take</tt> 
```python title='Indicative Python type signature'
def take(collection, n) -> AnyAnyCollection
```

Return a collection with the first n entries of the given collection.

#### Arguments

- `collection` of type [AnyAnyCollection](#anyanycollection)
- `n` of type `int`

#### Returns
- [AnyAnyCollection](#anyanycollection)

### <tt>tile_ids</tt> 
```python title='Indicative Python type signature'
def tile_ids(feature) -> FeatureIDIntCollection
```

Deprecated

#### Arguments

- `feature` of type [Feature](#feature)

#### Returns
- [FeatureIDIntCollection](#featureidintcollection)

### <tt>tile_ids_hex</tt> 
```python title='Indicative Python type signature'
def tile_ids_hex(feature) -> FeatureIDStringCollection
```

Deprecated

#### Arguments

- `feature` of type [Feature](#feature)

#### Returns
- [FeatureIDStringCollection](#featureidstringcollection)

### <tt>tile_paths</tt> 
```python title='Indicative Python type signature'
def tile_paths(geometry, zoom) -> IntStringCollection
```

Return the URL paths for the tiles containing the given geometry at the given zoom level.

#### Arguments

- `geometry` of type [Geometry](#geometry)
- `zoom` of type `int`

#### Returns
- [IntStringCollection](#intstringcollection)

### <tt>to_geojson</tt> 
```python title='Indicative Python type signature'
def to_geojson(renderable) -> GeoJSON
```


#### Arguments

- `renderable` of type [Geometry](#geometry)

#### Returns
- [GeoJSON](#geojson)

### <tt>to_geojson_collection</tt> 
```python title='Indicative Python type signature'
def to_geojson_collection(renderables) -> GeoJSON
```


#### Arguments

- `renderables` of type [AnyGeometryCollection](#anygeometrycollection)

#### Returns
- [GeoJSON](#geojson)

### <tt>to_str</tt> 
```python title='Indicative Python type signature'
def to_str(a) -> string
```


#### Arguments

- `a` of type `int`

#### Returns
- `string`

### <tt>top</tt> 
```python title='Indicative Python type signature'
def top(collection, n) -> AnyAnyCollection
```

Return a collection with the n entries from the given collection with the greatest values.
Requires the values of the given collection to be integers or floats.

#### Arguments

- `collection` of type [AnyAnyCollection](#anyanycollection)
- `n` of type `int`

#### Returns
- [AnyAnyCollection](#anyanycollection)

### <tt>type_area</tt> 
```python title='Indicative Python type signature'
def type_area() -> QueryProto
```

Return a query that will match area features.

#### Arguments


#### Returns
- [QueryProto](#queryproto)

### <tt>type_path</tt> 
```python title='Indicative Python type signature'
def type_path() -> QueryProto
```

Return a query that will match path features.

#### Arguments


#### Returns
- [QueryProto](#queryproto)

### <tt>type_point</tt> 
```python title='Indicative Python type signature'
def type_point() -> QueryProto
```

Return a query that will match point features.

#### Arguments


#### Returns
- [QueryProto](#queryproto)

### <tt>typed</tt> 
```python title='Indicative Python type signature'
def typed(typ, q) -> Query
```

Wrap a query to only match features with the given feature type.

#### Arguments

- `typ` of type `string`
- `q` of type [Query](#query)

#### Returns
- [Query](#query)

### <tt>value</tt> 
```python title='Indicative Python type signature'
def value(tag) -> string
```

Return the value of the given tag as a string.

#### Arguments

- `tag` of type [Tag](#tag)

#### Returns
- `string`

### <tt>with_change</tt> 
```python title='Indicative Python type signature'
def with_change(change, function) -> Any
```

Return the result of calling the given function in a world in which the given change has been applied.
The underlying world used by the server is not modified.

#### Arguments

- `change` of type [Change](#change)
- `function` of type `FunctionAny`

#### Returns
- [Any](#any)

### <tt>within</tt> 
```python title='Indicative Python type signature'
def within(a) -> Query
```

Return a query that will match features that intersect the given area.
Deprecated. Use intersecting.

#### Arguments

- `a` of type [Area](#area)

#### Returns
- [Query](#query)

### <tt>within_cap</tt> 
```python title='Indicative Python type signature'
def within_cap(point, radius) -> Query
```

Return a query that will match features that intersect a spherical cap centred on the given point, with the given radius in meters.
Deprecated. Use intersecting-cap.

#### Arguments

- `point` of type [Geometry](#geometry)
- `radius` of type `float64`

#### Returns
- [Query](#query)
## Collections

### <tt>AnyAnyAnyCollectionCollection</tt>

|Key|Value|
|---|-----|
[Any](#any)|[AnyAnyCollection](#anyanycollection)

### <tt>AnyAnyCollection</tt>

|Key|Value|
|---|-----|
[Any](#any)|[Any](#any)

### <tt>AnyAnyCollection</tt>

|Key|Value|
|---|-----|
[Any](#any)|[Any](#any)

### <tt>AnyAreaCollection</tt>

|Key|Value|
|---|-----|
[Any](#any)|[Area](#area)

### <tt>AnyChangeCollection</tt>

|Key|Value|
|---|-----|
[Any](#any)|[Change](#change)

### <tt>AnyFeatureCollection</tt>

|Key|Value|
|---|-----|
[Any](#any)|[Feature](#feature)

### <tt>AnyFeatureIDCollection</tt>

|Key|Value|
|---|-----|
[Any](#any)|[FeatureID](#featureid)

### <tt>AnyFloat64Collection</tt>

|Key|Value|
|---|-----|
[Any](#any)|`float64`

### <tt>AnyGeometryCollection</tt>

|Key|Value|
|---|-----|
[Any](#any)|[Geometry](#geometry)

### <tt>AnyIdentifiableCollection</tt>

|Key|Value|
|---|-----|
[Any](#any)|[Identifiable](#identifiable)

### <tt>AnyIntCollection</tt>

|Key|Value|
|---|-----|
[Any](#any)|`int`

### <tt>AnyTagCollection</tt>

|Key|Value|
|---|-----|
[Any](#any)|[Tag](#tag)

### <tt>FeatureIDAreaFeatureCollection</tt>

|Key|Value|
|---|-----|
[FeatureID](#featureid)|[AreaFeature](#areafeature)

### <tt>FeatureIDFeatureCollection</tt>

|Key|Value|
|---|-----|
[FeatureID](#featureid)|[Feature](#feature)

### <tt>FeatureIDFeatureIDCollection</tt>

|Key|Value|
|---|-----|
[FeatureID](#featureid)|[FeatureID](#featureid)

### <tt>FeatureIDGeometryCollection</tt>

|Key|Value|
|---|-----|
[FeatureID](#featureid)|[Geometry](#geometry)

### <tt>FeatureIDIntCollection</tt>

|Key|Value|
|---|-----|
[FeatureID](#featureid)|`int`

### <tt>FeatureIDPhysicalFeatureCollection</tt>

|Key|Value|
|---|-----|
[FeatureID](#featureid)|`PhysicalFeature`

### <tt>FeatureIDRelationFeatureCollection</tt>

|Key|Value|
|---|-----|
[FeatureID](#featureid)|[RelationFeature](#relationfeature)

### <tt>FeatureIDRouteCollection</tt>

|Key|Value|
|---|-----|
[FeatureID](#featureid)|`Route`

### <tt>FeatureIDStringCollection</tt>

|Key|Value|
|---|-----|
[FeatureID](#featureid)|`string`

### <tt>FeatureIDTagCollection</tt>

|Key|Value|
|---|-----|
[FeatureID](#featureid)|[Tag](#tag)

### <tt>IdentifiableStringCollection</tt>

|Key|Value|
|---|-----|
[Identifiable](#identifiable)|`string`

### <tt>IntAreaCollection</tt>

|Key|Value|
|---|-----|
`int`|[Area](#area)

### <tt>IntGeometryCollection</tt>

|Key|Value|
|---|-----|
`int`|[Geometry](#geometry)

### <tt>IntStringCollection</tt>

|Key|Value|
|---|-----|
`int`|`string`

### <tt>IntTagCollection</tt>

|Key|Value|
|---|-----|
`int`|[Tag](#tag)

### <tt>StringGeometryCollection</tt>

|Key|Value|
|---|-----|
`string`|[Geometry](#geometry)
## Interfaces

### <tt>Any</tt>


### <tt>Area</tt>

#### Implements
- [Geometry](#geometry)

### <tt>AreaFeature</tt>

#### Implements
- [Area](#area)
- [Feature](#feature)

### <tt>Callable</tt>


### <tt>Change</tt>


### <tt>CollectionFeature</tt>

#### Implements
- [AnyAnyCollection](#anyanycollection)
- [Feature](#feature)

### <tt>CollectionID</tt>

#### Implements
- [Identifiable](#identifiable)

### <tt>Expression</tt>


### <tt>Feature</tt>

#### Implements
- [Identifiable](#identifiable)

### <tt>FeatureID</tt>

#### Implements
- [Identifiable](#identifiable)

### <tt>GeoJSON</tt>


### <tt>Geometry</tt>


### <tt>Identifiable</tt>


### <tt>Number</tt>


### <tt>Pair</tt>


### <tt>Query</tt>


### <tt>Query</tt>


### <tt>QueryProto</tt>


### <tt>RelationFeature</tt>

#### Implements
- [Feature](#feature)

### <tt>RelationID</tt>

#### Implements
- [Identifiable](#identifiable)

### <tt>Tag</tt>


### <tt>bool</tt>

