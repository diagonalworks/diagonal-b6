package ingest

import (
	"sync"

	"diagonal.works/b6"
)

var DefaultWorldFeatureID = b6.FeatureID{
	Type:      b6.FeatureTypeCollection,
	Namespace: "diagonal.works/world",
	Value:     0,
}

type Worlds interface {
	FindOrCreateWorld(id b6.FeatureID) MutableWorld
}

type MutableWorlds struct {
	Base    b6.World
	Mutable map[b6.FeatureID]MutableWorld
	lock    sync.Mutex
}

func (m *MutableWorlds) FindOrCreateWorld(id b6.FeatureID) MutableWorld {
	if !id.IsValid() {
		id = DefaultWorldFeatureID
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	w, ok := m.Mutable[id]
	if !ok {
		if m.Mutable == nil {
			m.Mutable = make(map[b6.FeatureID]MutableWorld)
		}
		w = NewMutableOverlayWorld(m.Base)
		m.Mutable[id] = w
	}
	return w
}

type ReadOnlyWorlds struct {
	Base b6.World
}

func (r ReadOnlyWorlds) FindOrCreateWorld(id b6.FeatureID) MutableWorld {
	return ReadOnlyWorld{World: r.Base}
}
