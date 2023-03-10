package ingest

import (
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/osm"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

const idSetBuckets = 256
const idSetMask = 0xff

type IDSet struct {
	buckets []map[uint64]struct{}
	locks   []sync.Mutex
}

func NewIDSet() *IDSet {
	ids := &IDSet{buckets: make([]map[uint64]struct{}, idSetBuckets), locks: make([]sync.Mutex, idSetBuckets)}
	for i := range ids.buckets {
		ids.buckets[i] = make(map[uint64]struct{})
	}
	return ids
}

func (ids *IDSet) Add(id uint64) {
	bucket := id & idSetMask
	ids.locks[bucket].Lock()
	ids.buckets[bucket][id] = struct{}{}
	ids.locks[bucket].Unlock()
}

func (ids *IDSet) Has(id uint64) bool {
	_, ok := ids.buckets[id&idSetMask][id]
	return ok
}

func (ids *IDSet) Len() int {
	l := 0
	for _, ids := range ids.buckets {
		l += len(ids)
	}
	return l
}

const (
	// Nodes that are only used to represent points without tags, and hence dropped during indexing
	PrivateNodeIDsBegin osm.NodeID = 10000000000
)

const (
	// Loops generated from multi-way OSM boundaries
	BoundaryWayIDsBegin osm.WayID = 1000000000
)

const (
	// Polygons generated from multi-way OSM boundaries
	BoundaryRelationIDsBegin osm.RelationID = 100000000
)

func NewLatLngID(ll s2.LatLng) b6.PointID {
	id := (uint64(uint32(ll.Lat.E7())) << 32) | uint64(uint32(ll.Lng.E7()))
	return b6.MakePointID(b6.NamespaceLatLng, id)
}

func LatLngFromID(id b6.PointID) (s2.LatLng, bool) {
	if id.Namespace == b6.NamespaceLatLng {
		latE7 := int32((id.Value >> 32) & ((1 << 32) - 1))
		lngE7 := int32(id.Value & uint64((1<<32)-1))
		return s2.LatLng{Lat: s1.Angle(latE7) * s1.E7, Lng: s1.Angle(lngE7) * s1.E7}, true
	}
	return s2.LatLng{}, false
}
