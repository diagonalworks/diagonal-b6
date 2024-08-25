---
sidebar_position: 1
---

# b6 API documentation
## Functions

  This is documentationed generated from the `b6-api` binary.

  Below are all the functions, with their Python function name, and the
  corresponding argument names and types. Note that the types are the **b6 go**
  type definitions; not the python ones; nevertheless it is indicative of what
  type to expect to construct on the python side.
  

### *b6.accessible_all* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[FeatureID,FeatureID]</span>

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

- `origins` of type `Collection[Any,Identifiable]`
- `destinations` of type [`Query`](#query)
- `duration` of type `float64`
- `options` of type `Collection[Any,Any]`

#### Returns
`Collection[FeatureID,FeatureID]`

### *b6.accessible_routes* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[FeatureID,Route]</span>


#### Arguments

- `origin` of type [`Identifiable`](#identifiable)
- `destinations` of type [`Query`](#query)
- `duration` of type `float64`
- `options` of type `Collection[Any,Any]`

#### Returns
`Collection[FeatureID,Route]`

### *b6.add* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Number</span>

Return a added to b.

#### Arguments

- `a` of type [`Number`](#number)
- `b` of type [`Number`](#number)

#### Returns
[`Number`](#number)

### *b6.add_collection* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Add a collection feature with the given id, tags and items.

#### Arguments

- `id` of type [`CollectionID`](#collectionid)
- `tags` of type `Collection[Any,Tag]`
- `collection` of type `Collection[Any,Any]`

#### Returns
[`Change`](#change)

### *b6.add_expression* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Add an expression feature with the given id, tags and expression.

#### Arguments

- `id` of type [`FeatureID`](#featureid)
- `tags` of type `Collection[Any,Tag]`
- `expresson` of type [`Expression`](#expression)

#### Returns
[`Change`](#change)

### *b6.add_ints* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: int</span>

Deprecated.

#### Arguments

- `a` of type `int`
- `b` of type `int`

#### Returns
`int`

### *b6.add_point* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Adds a point feature with the given id, tags and members.

#### Arguments

- `point` of type [`Geometry`](#geometry)
- `id` of type [`FeatureID`](#featureid)
- `tags` of type `Collection[Any,Tag]`

#### Returns
[`Change`](#change)

### *b6.add_relation* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Add a relation feature with the given id, tags and members.

#### Arguments

- `id` of type [`RelationID`](#relationid)
- `tags` of type `Collection[Any,Tag]`
- `members` of type `Collection[Identifiable,String]`

#### Returns
[`Change`](#change)

### *b6.add_tag* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Add the given tag to the given feature.

#### Arguments

- `id` of type [`Identifiable`](#identifiable)
- `tag` of type [`Tag`](#tag)

#### Returns
[`Change`](#change)

### *b6.add_tags* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Add the given tags to the given features.
The keys of the given collection specify the features to change, the
values provide the tag to be added.

#### Arguments

- `collection` of type `Collection[FeatureID,Tag]`

#### Returns
[`Change`](#change)

### *b6.add_world_with_change* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[FeatureID,FeatureID]</span>


#### Arguments

- `id` of type [`FeatureID`](#featureid)
- `change` of type [`Change`](#change)

#### Returns
`Collection[FeatureID,FeatureID]`

### *b6.all* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Query</span>

Return a query that will match any feature.

#### Arguments


#### Returns
[`Query`](#query)

### *b6.all_tags* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Int,Tag]</span>

Return a collection of all the tags on the given feature.
Keys are ordered integers from 0, values are tags.

#### Arguments

- `id` of type [`Identifiable`](#identifiable)

#### Returns
`Collection[Int,Tag]`

### *b6.and* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Query</span>

Return a query that will match features that match both given queries.

#### Arguments

- `a` of type [`Query`](#query)
- `b` of type [`Query`](#query)

#### Returns
[`Query`](#query)

### *b6.apply_to_area* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Callable</span>

Wrap the given function such that it will only be called when passed an area.

#### Arguments

- `f` of type [`Callable`](#callable)

#### Returns
[`Callable`](#callable)

### *b6.apply_to_path* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Callable</span>

Wrap the given function such that it will only be called when passed a path.

#### Arguments

- `f` of type [`Callable`](#callable)

#### Returns
[`Callable`](#callable)

### *b6.apply_to_point* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Callable</span>

Wrap the given function such that it will only be called when passed a point.

#### Arguments

- `f` of type [`Callable`](#callable)

#### Returns
[`Callable`](#callable)

### *b6.area* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: float64</span>

Return the area of the given polygon in m².

#### Arguments

- `area` of type [`Area`](#area)

#### Returns
`float64`

### *b6.building_access* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[FeatureID,FeatureID]</span>

Deprecated. Use accessible.

#### Arguments

- `origins` of type `Collection[Any,Feature]`
- `limit` of type `float64`
- `mode` of type `string`

#### Returns
`Collection[FeatureID,FeatureID]`

### *b6.call* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Any</span>


#### Arguments

- `f` of type [`Callable`](#callable)
- `args` of type [`Any`](#any)

#### Returns
[`Any`](#any)
#### Misc
 - [x] Function is _variadic_ (has a variable number of arguments.)

### *b6.cap_polygon* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Area</span>

Return a polygon approximating a spherical cap with the given center and radius in meters.

#### Arguments

- `center` of type [`Geometry`](#geometry)
- `radius` of type `float64`

#### Returns
[`Area`](#area)

### *b6.centroid* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Geometry</span>

Return the centroid of the given geometry.
For multipolygons, we return the centroid of the convex hull formed from
the points of those polygons.

#### Arguments

- `geometry` of type [`Geometry`](#geometry)

#### Returns
[`Geometry`](#geometry)

### *b6.changes_from_file* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Return the changes contained in the given file.
As the file is read by the b6 server process, the filename it relative
to the filesystems it sees. Reading from files on cloud storage is
supported.

#### Arguments

- `filename` of type `string`

#### Returns
[`Change`](#change)

### *b6.changes_to_file* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: string</span>

Export the changes that have been applied to the world to the given filename as yaml.
As the file is written by the b6 server process, the filename it relative
to the filesystems it sees. Writing files to cloud storage is
supported.

#### Arguments

- `filename` of type `string`

#### Returns
`string`

### *b6.clamp* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: int</span>

Return the given value, unless it falls outside the given inclusive bounds, in which case return the boundary.

#### Arguments

- `v` of type `int`
- `low` of type `int`
- `high` of type `int`

#### Returns
`int`

### *b6.closest* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Feature</span>

Return the closest feature from the given origin via the given mode, within the given distance in meters, matching the given query.
See accessible-all for options values.

#### Arguments

- `origin` of type [`Feature`](#feature)
- `options` of type `Collection[Any,Any]`
- `distance` of type `float64`
- `query` of type [`Query`](#query)

#### Returns
[`Feature`](#feature)

### *b6.closest_distance* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: float64</span>

Return the distance through the graph of the closest feature from the given origin via the given mode, within the given distance in meters, matching the given query.
See accessible-all for options values.

#### Arguments

- `origin` of type [`Feature`](#feature)
- `options` of type `Collection[Any,Any]`
- `distance` of type `float64`
- `query` of type [`Query`](#query)

#### Returns
`float64`

### *b6.collect_areas* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Area</span>

Return a single area containing all areas from the given collection.
If areas in the collection overlap, loops within the returned area
will overlap, which will likely cause undefined behaviour in many
functions.

#### Arguments

- `areas` of type `Collection[Any,Area]`

#### Returns
[`Area`](#area)

### *b6.collection* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Any]</span>

Return a collection of the given key value pairs.

#### Arguments

- `pairs` of type [`Any`](#any)

#### Returns
`Collection[Any,Any]`
#### Misc
 - [x] Function is _variadic_ (has a variable number of arguments.)

### *b6.connect* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Add a path that connects the two given points, if they're not already directly connected.

#### Arguments

- `a` of type [`Feature`](#feature)
- `b` of type [`Feature`](#feature)

#### Returns
[`Change`](#change)

### *b6.connect_to_network* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Add a path and point to connect given feature to the street network.
The street network is defined at the set of paths tagged #highway that
allow traversal of more than 500m. A point is added to the closest
network path at the projection of the origin point on that path, unless
that point is within 4m of an existing path point.

#### Arguments

- `feature` of type [`Feature`](#feature)

#### Returns
[`Change`](#change)

### *b6.connect_to_network_all* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Add paths and points to connect the given collection of features to the
network. See connect-to-network for connection details.
More efficient than using map with connect-to-network, as the street
network is only computed once.

#### Arguments

- `features` of type `Collection[Any,FeatureID]`

#### Returns
[`Change`](#change)

### *b6.containing_areas* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[FeatureID,AreaFeature]</span>


#### Arguments

- `points` of type `Collection[Any,Feature]`
- `q` of type [`Query`](#query)

#### Returns
`Collection[FeatureID,AreaFeature]`

### *b6.convex_hull* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Area</span>

Return the convex hull of the given geometries.

#### Arguments

- `c` of type `Collection[Any,Geometry]`

#### Returns
[`Area`](#area)

### *b6.count* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: int</span>

Return the number of items in the given collection.
The function will not evaluate and traverse the entire collection if it's possible to count
the collection efficiently.

#### Arguments

- `collection` of type `Collection[Any,Any]`

#### Returns
`int`

### *b6.count_keys* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Int]</span>

Return a collection of the number of occurances of each value in the given collection.

#### Arguments

- `collection` of type `Collection[Any,Any]`

#### Returns
`Collection[Any,Int]`

### *b6.count_tag_value* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Int]</span>

Deprecated.

#### Arguments

- `id` of type [`Identifiable`](#identifiable)
- `key` of type `string`

#### Returns
`Collection[Any,Int]`

### *b6.count_valid_ids* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: int</span>

Return the number of valid feature IDs in the given collection.

#### Arguments

- `collection` of type `Collection[Any,Identifiable]`

#### Returns
`int`

### *b6.count_valid_keys* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Int]</span>

Return a collection of the number of occurances of each valid value in the given collection.
Invalid values are not counted, but case the key to appear in the output.

#### Arguments

- `collection` of type `Collection[Any,Any]`

#### Returns
`Collection[Any,Int]`

### *b6.count_values* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Int]</span>

Return a collection of the number of occurances of each value in the given collection.

#### Arguments

- `collection` of type `Collection[Any,Any]`

#### Returns
`Collection[Any,Int]`

### *b6.debug_all_query* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Query</span>

Deprecated.

#### Arguments

- `token` of type `string`

#### Returns
[`Query`](#query)

### *b6.debug_tokens* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Int,String]</span>

Return the search index tokens generated for the given feature.
Intended for debugging use only.

#### Arguments

- `id` of type [`Identifiable`](#identifiable)

#### Returns
`Collection[Int,String]`

### *b6.degree* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: int</span>

Return the number of paths connected to the given point.
A single path will be counted twice if the point isn't at one of its
two ends - once in one direction, and once in the other.

#### Arguments

- `point` of type [`Feature`](#feature)

#### Returns
`int`

### *b6.distance_meters* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: float64</span>

Return the distance in meters between the given points.

#### Arguments

- `a` of type [`Geometry`](#geometry)
- `b` of type [`Geometry`](#geometry)

#### Returns
`float64`

### *b6.distance_to_point_meters* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: float64</span>

Return the distance in meters between the given path, and the project of the give point onto it.

#### Arguments

- `path` of type [`Geometry`](#geometry)
- `point` of type [`Geometry`](#geometry)

#### Returns
`float64`

### *b6.divide* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Number</span>

Return a divided by b.

#### Arguments

- `a` of type [`Number`](#number)
- `b` of type [`Number`](#number)

#### Returns
[`Number`](#number)

### *b6.divide_int* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: float64</span>

Deprecated.

#### Arguments

- `a` of type `int`
- `b` of type `float64`

#### Returns
`float64`

### *b6.entrance_approach* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Geometry</span>


#### Arguments

- `area` of type [`AreaFeature`](#areafeature)

#### Returns
[`Geometry`](#geometry)

### *b6.evaluate_feature* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Any</span>


#### Arguments

- `id` of type [`FeatureID`](#featureid)

#### Returns
[`Any`](#any)

### *b6.export_world* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: int</span>

Write the current world to the given filename in the b6 compact index format.
As the file is written by the b6 server process, the filename it relative
to the filesystems it sees. Writing files to cloud storage is
supported.

#### Arguments

- `filename` of type `string`

#### Returns
`int`

### *b6.filter* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Any]</span>

Return a collection of the items of the given collection for which the value of the given function applied to each value is true.

#### Arguments

- `collection` of type `Collection[Any,Any]`
- `function` of type [`Callable`](#callable)

#### Returns
`Collection[Any,Any]`

### *b6.filter_accessible* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[FeatureID,FeatureID]</span>

Return a collection containing only the values of the given collection that match the given query.
If no values for a key match the query, emit a single invalid feature ID
for that key, allowing callers to count the number of keys with no valid
values.
Keys are taken from the given collection.

#### Arguments

- `collection` of type `Collection[FeatureID,FeatureID]`
- `filter` of type [`Query`](#query)

#### Returns
`Collection[FeatureID,FeatureID]`

### *b6.find* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[FeatureID,Feature]</span>

Return a collection of the features present in the world that match the given query.
Keys are IDs, and values are features.

#### Arguments

- `query` of type [`Query`](#query)

#### Returns
`Collection[FeatureID,Feature]`

### *b6.find_area* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: AreaFeature</span>

Return the area feature with the given ID.

#### Arguments

- `id` of type [`FeatureID`](#featureid)

#### Returns
[`AreaFeature`](#areafeature)

### *b6.find_areas* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[FeatureID,AreaFeature]</span>

Return a collection of the area features present in the world that match the given query.
Keys are IDs, and values are features.

#### Arguments

- `query` of type [`Query`](#query)

#### Returns
`Collection[FeatureID,AreaFeature]`

### *b6.find_collection* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: CollectionFeature</span>

Return the collection feature with the given ID.

#### Arguments

- `id` of type [`FeatureID`](#featureid)

#### Returns
[`CollectionFeature`](#collectionfeature)

### *b6.find_feature* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Feature</span>

Return the feature with the given ID.

#### Arguments

- `id` of type [`FeatureID`](#featureid)

#### Returns
[`Feature`](#feature)

### *b6.find_relation* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: RelationFeature</span>

Return the relation feature with the given ID.

#### Arguments

- `id` of type [`FeatureID`](#featureid)

#### Returns
[`RelationFeature`](#relationfeature)

### *b6.find_relations* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[FeatureID,RelationFeature]</span>

Return a collection of the relation features present in the world that match the given query.
Keys are IDs, and values are features.

#### Arguments

- `query` of type [`Query`](#query)

#### Returns
`Collection[FeatureID,RelationFeature]`

### *b6.first* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Any</span>

Return the first value of the given pair.

#### Arguments

- `pair` of type [`Pair`](#pair)

#### Returns
[`Any`](#any)

### *b6.flatten* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Any]</span>

Return a collection with keys and values taken from the collections that form the values of the given collection.

#### Arguments

- `collection` of type `Collection[Any,Collection[Any,Any]]`

#### Returns
`Collection[Any,Any]`

### *b6.float_value* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: float64</span>

Return the value of the given tag as a float.
Propagates error if the value isn't a valid float.

#### Arguments

- `tag` of type [`Tag`](#tag)

#### Returns
`float64`

### *b6.geojson_areas* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Int,Area]</span>

Return the areas present in the given geojson.

#### Arguments

- `g` of type [`GeoJSON`](#geojson)

#### Returns
`Collection[Int,Area]`

### *b6.get* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Tag</span>

Return the tag with the given key on the given feature.
Returns a tag. To return the string value of a tag, use get-string.

#### Arguments

- `id` of type [`Identifiable`](#identifiable)
- `key` of type `string`

#### Returns
[`Tag`](#tag)

### *b6.get_float* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: float64</span>

Return the value of tag with the given key on the given feature as a float.
Returns error if there isn't a feature with that id, a tag with that key, or if the value isn't a valid float.

#### Arguments

- `id` of type [`Identifiable`](#identifiable)
- `key` of type `string`

#### Returns
`float64`

### *b6.get_int* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: int</span>

Return the value of tag with the given key on the given feature as an integer.
Returns error if there isn't a feature with that id, a tag with that key, or if the value isn't a valid integer.

#### Arguments

- `id` of type [`Identifiable`](#identifiable)
- `key` of type `string`

#### Returns
`int`

### *b6.get_string* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: string</span>

Return the value of tag with the given key on the given feature as a string.
Returns an empty string if there isn't a tag with that key.

#### Arguments

- `id` of type [`Identifiable`](#identifiable)
- `key` of type `string`

#### Returns
`string`

### *b6.gt* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: bool</span>

Return true if a is greater than b.

#### Arguments

- `a` of type [`Any`](#any)
- `b` of type [`Any`](#any)

#### Returns
[`bool`](#bool)

### *b6.histogram* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Return a change that adds a histogram for the given collection.

#### Arguments

- `collection` of type `Collection[Any,Any]`

#### Returns
[`Change`](#change)

### *b6.histogram_swatch* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Return a change that adds a histogram with only colour swatches for the given collection.

#### Arguments

- `collection` of type `Collection[Any,Any]`

#### Returns
[`Change`](#change)

### *b6.histogram_swatch_with_id* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Return a change that adds a histogram with only colour swatches for the given collection.

#### Arguments

- `collection` of type `Collection[Any,Any]`
- `id` of type [`CollectionID`](#collectionid)

#### Returns
[`Change`](#change)

### *b6.histogram_with_id* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Return a change that adds a histogram for the given collection with the given ID.

#### Arguments

- `collection` of type `Collection[Any,Any]`
- `id` of type [`CollectionID`](#collectionid)

#### Returns
[`Change`](#change)

### *b6.id_to_relation_id* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: FeatureID</span>

Deprecated.

#### Arguments

- `namespace` of type `string`
- `id` of type [`Identifiable`](#identifiable)

#### Returns
[`FeatureID`](#featureid)

### *b6.import_geojson* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Add features from the given geojson to the world.
IDs are formed from the given namespace, and the index of the feature
within the geojson collection (or 0, if a single feature is used).

#### Arguments

- `features` of type [`GeoJSON`](#geojson)
- `namespace` of type `string`

#### Returns
[`Change`](#change)

### *b6.import_geojson_file* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

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
[`Change`](#change)

### *b6.int_value* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: int</span>

Return the value of the given tag as an integer.
Propagates error if the value isn't a valid integer.

#### Arguments

- `tag` of type [`Tag`](#tag)

#### Returns
`int`

### *b6.interpolate* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Geometry</span>

Return the point at the given fraction along the given path.

#### Arguments

- `path` of type [`Geometry`](#geometry)
- `fraction` of type `float64`

#### Returns
[`Geometry`](#geometry)

### *b6.intersecting* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Query</span>

Return a query that will match features that intersect the given geometry.

#### Arguments

- `geometry` of type [`Geometry`](#geometry)

#### Returns
[`Query`](#query)

### *b6.intersecting_cap* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Query</span>

Return a query that will match features that intersect a spherical cap centred on the given point, with the given radius in meters.

#### Arguments

- `center` of type [`Geometry`](#geometry)
- `radius` of type `float64`

#### Returns
[`Query`](#query)

### *b6.join* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Geometry</span>

Return a path formed from the points of the two given paths, in the order they occur in those paths.

#### Arguments

- `pathA` of type [`Geometry`](#geometry)
- `pathB` of type [`Geometry`](#geometry)

#### Returns
[`Geometry`](#geometry)

### *b6.join_missing* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Any]</span>


#### Arguments

- `base` of type `Collection[Any,Any]`
- `joined` of type `Collection[Any,Any]`

#### Returns
`Collection[Any,Any]`

### *b6.keyed* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Query</span>

Return a query that will match features tagged with the given key independent of value.

#### Arguments

- `key` of type `string`

#### Returns
[`Query`](#query)

### *b6.length* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: float64</span>

Return the length of the given path in meters.

#### Arguments

- `path` of type [`Geometry`](#geometry)

#### Returns
`float64`

### *b6.list_feature* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Any]</span>


#### Arguments

- `id` of type [`CollectionID`](#collectionid)

#### Returns
`Collection[Any,Any]`

### *b6.ll* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Geometry</span>

Return a point at the given latitude and longitude, specified in degrees.

#### Arguments

- `lat` of type `float64`
- `lng` of type `float64`

#### Returns
[`Geometry`](#geometry)

### *b6.map* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Any]</span>

Return a collection with the result of applying the given function to each value.
Keys are unmodified.

#### Arguments

- `collection` of type `Collection[Any,Any]`
- `function` of type [`Callable`](#callable)

#### Returns
`Collection[Any,Any]`

### *b6.map_geometries* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: GeoJSON</span>

Return a geojson representing the result of applying the given function to each geometry in the given geojson.

#### Arguments

- `g` of type [`GeoJSON`](#geojson)
- `f` of type [`Callable`](#callable)

#### Returns
[`GeoJSON`](#geojson)

### *b6.map_items* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Any]</span>

Return a collection of the result of applying the given function to each pair(key, value).
Keys are unmodified.

#### Arguments

- `collection` of type `Collection[Any,Any]`
- `function` of type [`Callable`](#callable)

#### Returns
`Collection[Any,Any]`

### *b6.map_parallel* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Any]</span>

Return a collection with the result of applying the given function to each value.
Keys are unmodified, and function application occurs in parallel, bounded
by the number of CPU cores allocated to b6.

#### Arguments

- `collection` of type `Collection[Any,Any]`
- `function` of type [`Callable`](#callable)

#### Returns
`Collection[Any,Any]`

### *b6.matches* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: bool</span>

Return true if the given feature matches the given query.

#### Arguments

- `id` of type [`Identifiable`](#identifiable)
- `query` of type [`Query`](#query)

#### Returns
[`bool`](#bool)

### *b6.materialise* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Return a change that adds a collection feature to the world with the given ID, containing the result of calling the given function.
The given function isn't passed any arguments.
Also adds an expression feature (with the same namespace and value)
representing the given function.

#### Arguments

- `id` of type [`CollectionID`](#collectionid)
- `function` of type [`Callable`](#callable)

#### Returns
[`Change`](#change)

### *b6.materialise_map* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>


#### Arguments

- `collection` of type `Collection[Any,Feature]`
- `id` of type [`CollectionID`](#collectionid)
- `function` of type [`Callable`](#callable)

#### Returns
[`Change`](#change)

### *b6.merge_changes* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Return a change that will apply all the changes in the given collection.
Changes are applied transactionally. If the application of one change
fails (for example, because it includes a path that references a missing
point), then no changes will be applied.

#### Arguments

- `collection` of type `Collection[Any,Change]`

#### Returns
[`Change`](#change)

### *b6.or* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Query</span>

Return a query that will match features that match either of the given queries.

#### Arguments

- `a` of type [`Query`](#query)
- `b` of type [`Query`](#query)

#### Returns
[`Query`](#query)

### *b6.ordered_join* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Geometry</span>

Returns a path formed by joining the two given paths.
If necessary to maintain consistency, the order of points is reversed,
determined by which points are shared between the paths. Returns an error
if no endpoints are shared.

#### Arguments

- `pathA` of type [`Geometry`](#geometry)
- `pathB` of type [`Geometry`](#geometry)

#### Returns
[`Geometry`](#geometry)

### *b6.pair* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Pair</span>

Return a pair containing the given values.

#### Arguments

- `first` of type [`Any`](#any)
- `second` of type [`Any`](#any)

#### Returns
[`Pair`](#pair)

### *b6.parse_geojson* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: GeoJSON</span>

Return the geojson represented by the given string.

#### Arguments

- `s` of type `string`

#### Returns
[`GeoJSON`](#geojson)

### *b6.parse_geojson_file* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: GeoJSON</span>

Return the geojson contained in the given file.
As the file is read by the b6 server process, the filename it relative
to the filesystems it sees. Reading from files on cloud storage is
supported.

#### Arguments

- `filename` of type `string`

#### Returns
[`GeoJSON`](#geojson)

### *b6.paths_to_reach* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[FeatureID,Int]</span>

Return a collection of the paths used to reach all features matching the given query from the given origin via the given mode, within the given distance in meters.
Keys are the paths used, values are the number of times that path was used during traversal.
See accessible-all for options values.

#### Arguments

- `origin` of type [`Feature`](#feature)
- `options` of type `Collection[Any,Any]`
- `distance` of type `float64`
- `query` of type [`Query`](#query)

#### Returns
`Collection[FeatureID,Int]`

### *b6.percentiles* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Float64]</span>

Return a collection where values represent the perentile of the corresponding value in the given collection.
The returned collection is ordered by percentile, with keys drawn from the
given collection.

#### Arguments

- `collection` of type `Collection[Any,Float64]`

#### Returns
`Collection[Any,Float64]`

### *b6.point_features* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[FeatureID,PhysicalFeature]</span>

Return a collection of the point features referenced by the given feature.
Keys are ids of the respective value, values are point features. Area
features return the points referenced by their path features.

#### Arguments

- `f` of type [`Feature`](#feature)

#### Returns
`Collection[FeatureID,PhysicalFeature]`

### *b6.point_paths* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[FeatureID,PhysicalFeature]</span>

Return a collection of the path features referencing the given point.
Keys are the ids of the respective paths.

#### Arguments

- `id` of type [`Identifiable`](#identifiable)

#### Returns
`Collection[FeatureID,PhysicalFeature]`

### *b6.points* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Int,Geometry]</span>

Return a collection of the points of the given geometry.
Keys are ordered integers from 0, values are points.

#### Arguments

- `geometry` of type [`Geometry`](#geometry)

#### Returns
`Collection[Int,Geometry]`

### *b6.reachable* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[FeatureID,Feature]</span>

Return the a collection of the features reachable from the given origin via the given mode, within the given distance in meters, that match the given query.
See accessible-all for options values.
Deprecated. Use accessible-all.

#### Arguments

- `origin` of type [`Feature`](#feature)
- `options` of type `Collection[Any,Any]`
- `distance` of type `float64`
- `query` of type [`Query`](#query)

#### Returns
`Collection[FeatureID,Feature]`

### *b6.reachable_area* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: float64</span>

Return the area formed by the convex hull of the features matching the given query reachable from the given origin via the given mode specified in options, within the given distance in meters.
See accessible-all for options values.

#### Arguments

- `origin` of type [`Feature`](#feature)
- `options` of type `Collection[Any,Any]`
- `distance` of type `float64`

#### Returns
`float64`

### *b6.rectangle_polygon* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Area</span>

Return a rectangle polygon with the given top left and bottom right points.

#### Arguments

- `a` of type [`Geometry`](#geometry)
- `b` of type [`Geometry`](#geometry)

#### Returns
[`Area`](#area)

### *b6.remove_tag* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Remove the tag with the given key from the given feature.

#### Arguments

- `id` of type [`Identifiable`](#identifiable)
- `key` of type `string`

#### Returns
[`Change`](#change)

### *b6.remove_tags* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Change</span>

Remove the given tags from the given features.
The keys of the given collection specify the features to change, the
values provide the key of the tag to be removed.

#### Arguments

- `collection` of type `Collection[FeatureID,String]`

#### Returns
[`Change`](#change)

### *b6.s2_center* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Geometry</span>

Return a collection the center of the s2 cell with the given token.

#### Arguments

- `token` of type `string`

#### Returns
[`Geometry`](#geometry)

### *b6.s2_covering* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Int,String]</span>

Return a collection of of s2 cells tokens that cover the given area at the given level.

#### Arguments

- `area` of type [`Area`](#area)
- `minLevel` of type `int`
- `maxLevel` of type `int`

#### Returns
`Collection[Int,String]`

### *b6.s2_grid* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Int,String]</span>

Return a collection of points representing the centroids of s2 cells that cover the given area at the given level.

#### Arguments

- `area` of type [`Area`](#area)
- `level` of type `int`

#### Returns
`Collection[Int,String]`

### *b6.s2_points* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[String,Geometry]</span>

Return a collection of points representing the centroids of s2 cells that cover the given area between the given levels.

#### Arguments

- `area` of type [`Area`](#area)
- `minLevel` of type `int`
- `maxLevel` of type `int`

#### Returns
`Collection[String,Geometry]`

### *b6.s2_polygon* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Area</span>

Return the bounding area of the s2 cell with the given token.

#### Arguments

- `token` of type `string`

#### Returns
[`Area`](#area)

### *b6.sample_points* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Int,Geometry]</span>

Return a collection of points along the given path, with the given distance in meters between them.
Keys are ordered integers from 0, values are points.

#### Arguments

- `path` of type [`Geometry`](#geometry)
- `distanceMeters` of type `float64`

#### Returns
`Collection[Int,Geometry]`

### *b6.sample_points_along_paths* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Int,Geometry]</span>

Return a collection of points along the given paths, with the given distance in meters between them.
Keys are the id of the respective path, values are points.

#### Arguments

- `paths` of type `Collection[FeatureID,Geometry]`
- `distanceMeters` of type `float64`

#### Returns
`Collection[Int,Geometry]`

### *b6.second* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Any</span>

Return the second value of the given pair.

#### Arguments

- `pair` of type [`Pair`](#pair)

#### Returns
[`Any`](#any)

### *b6.sightline* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Area</span>


#### Arguments

- `from` of type [`Geometry`](#geometry)
- `radius` of type `float64`

#### Returns
[`Area`](#area)

### *b6.snap_area_edges* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Area</span>

Return an area formed by projecting the edges of the given polygon onto the paths present in the world matching the given query.
Paths beyond the given threshold in meters are ignored.

#### Arguments

- `area` of type [`Area`](#area)
- `query` of type [`Query`](#query)
- `threshold` of type `float64`

#### Returns
[`Area`](#area)

### *b6.sum* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: int</span>

Return the sum of all values in a given collection.

#### Arguments

- `collection` of type `Collection[Any,Int]`

#### Returns
`int`

### *b6.sum_by_key* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Int]</span>

Return a collection of the result of summing the values of each item with the same key.
Requires values to be integers.

#### Arguments

- `c` of type `Collection[Any,Int]`

#### Returns
`Collection[Any,Int]`

### *b6.tag* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Tag</span>

Return a tag with the given key and value.

#### Arguments

- `key` of type `string`
- `value` of type `string`

#### Returns
[`Tag`](#tag)

### *b6.tagged* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Query</span>

Return a query that will match features tagged with the given key and value.

#### Arguments

- `key` of type `string`
- `value` of type `string`

#### Returns
[`Query`](#query)

### *b6.take* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Any]</span>

Return a collection with the first n entries of the given collection.

#### Arguments

- `collection` of type `Collection[Any,Any]`
- `n` of type `int`

#### Returns
`Collection[Any,Any]`

### *b6.tile_ids* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[FeatureID,Int]</span>

Deprecated

#### Arguments

- `feature` of type [`Feature`](#feature)

#### Returns
`Collection[FeatureID,Int]`

### *b6.tile_ids_hex* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[FeatureID,String]</span>

Deprecated

#### Arguments

- `feature` of type [`Feature`](#feature)

#### Returns
`Collection[FeatureID,String]`

### *b6.tile_paths* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Int,String]</span>

Return the URL paths for the tiles containing the given geometry at the given zoom level.

#### Arguments

- `geometry` of type [`Geometry`](#geometry)
- `zoom` of type `int`

#### Returns
`Collection[Int,String]`

### *b6.to_geojson* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: GeoJSON</span>


#### Arguments

- `renderable` of type [`Geometry`](#geometry)

#### Returns
[`GeoJSON`](#geojson)

### *b6.to_geojson_collection* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: GeoJSON</span>


#### Arguments

- `renderables` of type `Collection[Any,Geometry]`

#### Returns
[`GeoJSON`](#geojson)

### *b6.to_str* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: string</span>


#### Arguments

- `a` of type `int`

#### Returns
`string`

### *b6.top* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Collection[Any,Any]</span>

Return a collection with the n entries from the given collection with the greatest values.
Requires the values of the given collection to be integers or floats.

#### Arguments

- `collection` of type `Collection[Any,Any]`
- `n` of type `int`

#### Returns
`Collection[Any,Any]`

### *b6.type_area* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: QueryProto</span>

Return a query that will match area features.

#### Arguments


#### Returns
[`QueryProto`](#queryproto)

### *b6.type_path* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: QueryProto</span>

Return a query that will match path features.

#### Arguments


#### Returns
[`QueryProto`](#queryproto)

### *b6.type_point* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: QueryProto</span>

Return a query that will match point features.

#### Arguments


#### Returns
[`QueryProto`](#queryproto)

### *b6.typed* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Query</span>

Wrap a query to only match features with the given feature type.

#### Arguments

- `typ` of type `string`
- `q` of type [`Query`](#query)

#### Returns
[`Query`](#query)

### *b6.value* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: string</span>

Return the value of the given tag as a string.

#### Arguments

- `tag` of type [`Tag`](#tag)

#### Returns
`string`

### *b6.with_change* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Any</span>

Return the result of calling the given function in a world in which the given change has been applied.
The underlying world used by the server is not modified.

#### Arguments

- `change` of type [`Change`](#change)
- `function` of type `FunctionAny`

#### Returns
[`Any`](#any)

### *b6.within* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Query</span>

Return a query that will match features that intersect the given area.
Deprecated. Use intersecting.

#### Arguments

- `a` of type [`Area`](#area)

#### Returns
[`Query`](#query)

### *b6.within_cap* <span style={{fontSize: 12 +'px', fontWeight: 'normal'}}>:: Query</span>

Return a query that will match features that intersect a spherical cap centred on the given point, with the given radius in meters.
Deprecated. Use intersecting-cap.

#### Arguments

- `point` of type [`Geometry`](#geometry)
- `radius` of type `float64`

#### Returns
[`Query`](#query)
## Interfaces

### *Any*


### *Area*

#### Implements
- [Geometry](#geometry)

### *AreaFeature*

#### Implements
- [Area](#area)
- [Feature](#feature)

### *Callable*


### *Change*


### *CollectionFeature*

#### Implements
- [Collection[Any,Any]](#collection[any,any])
- [Feature](#feature)

### *CollectionID*

#### Implements
- [Identifiable](#identifiable)

### *Expression*


### *Feature*

#### Implements
- [Identifiable](#identifiable)

### *FeatureID*

#### Implements
- [Identifiable](#identifiable)

### *GeoJSON*


### *Geometry*


### *Identifiable*


### *Number*


### *Pair*


### *Query*


### *Query*


### *QueryProto*


### *RelationFeature*

#### Implements
- [Feature](#feature)

### *RelationID*

#### Implements
- [Identifiable](#identifiable)

### *Tag*


### *bool*

