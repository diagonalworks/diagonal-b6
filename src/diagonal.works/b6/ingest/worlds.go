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
	ListWorlds() []b6.FeatureID
	DeleteWorld(id b6.FeatureID)
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

func (m *MutableWorlds) ListWorlds() []b6.FeatureID {
	m.lock.Lock()
	defer m.lock.Unlock()
	if len(m.Mutable) == 0 {
		return []b6.FeatureID{DefaultWorldFeatureID}
	}
	ids := make([]b6.FeatureID, 0, len(m.Mutable)+1)
	for id := range m.Mutable {
		ids = append(ids, id)
	}
	return ids
}

func (m *MutableWorlds) DeleteWorld(id b6.FeatureID) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.Mutable, id)
}

type ReadOnlyWorlds struct {
	Base b6.World
}

func (r ReadOnlyWorlds) FindOrCreateWorld(id b6.FeatureID) MutableWorld {
	return ReadOnlyWorld{World: r.Base}
}

func (r ReadOnlyWorlds) ListWorlds() []b6.FeatureID {
	return []b6.FeatureID{DefaultWorldFeatureID}
}

func (r ReadOnlyWorlds) DeleteWorld(id b6.FeatureID) {}
