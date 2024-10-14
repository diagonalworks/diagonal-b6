---
sidebar_position: 1
---

# Worlds

Worlds are the central organising element in b6. Everything exists within a
particular world, either an explicit one, or the default one, called
`DefaultWorldFeatureID` in the codebase and taking on the value:

```
/collection/diagonal.works/world/0
```

## The `World` interface

The central `World` interface.

```go
// In file world.go
type World interface {
  FindFeatureByID(id FeatureID) Feature
  HasFeatureWithID(id FeatureID) bool
  FindLocationByID(id FeatureID) (s2.LatLng, error)
  FindFeatures(query Query) Features
  FindRelationsByFeature(id FeatureID) RelationFeatures
  FindCollectionsByFeature(id FeatureID) CollectionFeatures
  FindAreasByPoint(id FeatureID) AreaFeatures
  FindReferences(id FeatureID, typed ...FeatureType) Features
  Traverse(id FeatureID) Segments
  EachFeature(
      each func(f Feature, goroutine int) error
    , options *EachFeatureOptions
    )
    error
  Tokens() []string
}
```

### Instances of `World`


The central world type, containing a list of features, and indices.

```go
// In file ingest/compact/world.go
type  World struct {
  byID    *FeaturesByID
  indices []*Index
  status  string
  lock    sync.Mutex
}
```

#### <tt>OverlayWorld</tt>


```go
// In file ingest/overlay.go
type OverlayWorld struct {
  overlay b6.World
  base    b6.World
}
```


#### <tt>ReadOnlyWorld</tt>

A very simple wrapper over `World`; basically just calls out to `World`.

```go
// In file ingest/mutable.go
type ReadOnlyWorld struct {
  World b6.World
}
```

#### <tt>BasicMutableWorld</tt>

```go
// In file ingest/mutable.go
type BasicMutableWorld struct {
  features   *FeaturesByID
  references *FeatureReferencesByID
  index      *mutableFeatureIndex
}
```

#### <tt>MutableOverlayWorld</tt>

```go
// In file ingest/mutable.go
type MutableOverlayWorld struct {
  features   *FeaturesByID
  references *FeatureReferencesByID
  index      *mutableFeatureIndex
  base       b6.World
  tags       ModifiedTags
  epoch      int
}
```

#### <tt>MutableTagsOverlayWorld</tt>

```go
// In file ingest/mutable.go
type MutableTagsOverlayWorld struct {
  tags     ModifiedTags
  base     b6.World
  watchers []*watcher
}
```

## Worlds

The `Worlds` interface represents a list of worlds.

```go
type Worlds interface {
  FindOrCreateWorld(id b6.FeatureID) MutableWorld
  ListWorlds() []b6.FeatureID
  DeleteWorld(id b6.FeatureID)
}
```

### Instances of `Worlds`

There are only two choices:

- <tt>MutableWorlds</tt> - For worlds that can be "changed".
- <tt>ReadOnlyWorlds</tt> - A single world that can never change.

```go
// In file ingest/worlds.go
type MutableWorlds struct {
  Base    b6.World
  Mutable map[b6.FeatureID]MutableWorld
  lock    sync.Mutex
}

// ...

type ReadOnlyWorlds struct {
  Base b6.World
}
```

## The `MutableWorld` interface


```go
type MutableWorld interface {
  b6.World

  AddFeature(
      f Feature
    )
    error

  AddTag(
      id b6.FeatureID
    , tag b6.Tag
    )
    error

  RemoveTag(
      id b6.FeatureID
    , key string
    )
    error

  EachModifiedFeature(
      each func(f b6.Feature, goroutine int) error
    , options *b6.EachFeatureOptions
    )
    error

  EachModifiedTag(
      each func(f ModifiedTag, goroutine int) error
    , options *b6.EachFeatureOptions
    )
    error
}
```

## Construction of `World`

See [Ingest](./ingest).
