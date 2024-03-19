package ingest

import (
	"sync"

	"diagonal.works/b6"
	"github.com/golang/geo/s2"
)

type LockedWorld struct {
	World MutableWorld
	lock  sync.RWMutex
}

func NewLockedWorld(w MutableWorld) *LockedWorld {
	return &LockedWorld{World: w}
}

func (l *LockedWorld) FindFeatureByID(id b6.FeatureID) b6.Feature {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.World.FindFeatureByID(id)
}

func (l *LockedWorld) HasFeatureWithID(id b6.FeatureID) bool {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.World.HasFeatureWithID(id)
}

func (l *LockedWorld) FindLocationByID(id b6.FeatureID) (s2.LatLng, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.World.FindLocationByID(id)
}

func (l *LockedWorld) FindFeatures(query b6.Query) b6.Features {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.World.FindFeatures(query)
}

func (l *LockedWorld) FindRelationsByFeature(id b6.FeatureID) b6.RelationFeatures {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.World.FindRelationsByFeature(id)
}

func (l *LockedWorld) FindCollectionsByFeature(id b6.FeatureID) b6.CollectionFeatures {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.World.FindCollectionsByFeature(id)
}

func (l *LockedWorld) FindPathsByPoint(id b6.FeatureID) b6.PathFeatures {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.World.FindPathsByPoint(id)
}

func (l *LockedWorld) FindAreasByPoint(id b6.FeatureID) b6.AreaFeatures {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.World.FindAreasByPoint(id)
}

func (l *LockedWorld) FindReferences(id b6.FeatureID, typed ...b6.FeatureType) b6.Features {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.World.FindReferences(id, typed...)
}

func (l *LockedWorld) Traverse(id b6.FeatureID) b6.Segments {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.World.Traverse(id)
}

func (l *LockedWorld) EachFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.World.EachFeature(each, options)
}

func (l *LockedWorld) Tokens() []string {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.World.Tokens()
}

func (l *LockedWorld) AddFeature(f Feature) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.World.AddFeature(f)
}

func (l *LockedWorld) AddTag(id b6.FeatureID, tag b6.Tag) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.World.AddTag(id, tag)
}

func (l *LockedWorld) RemoveTag(id b6.FeatureID, key string) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.World.RemoveTag(id, key)
}

func (l *LockedWorld) EachModifiedFeature(each func(f b6.Feature, goroutine int) error, options *b6.EachFeatureOptions) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.World.EachModifiedFeature(each, options)
}

func (l *LockedWorld) EachModifiedTag(each func(f ModifiedTag, goroutine int) error, options *b6.EachFeatureOptions) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.World.EachModifiedTag(each, options)
}
