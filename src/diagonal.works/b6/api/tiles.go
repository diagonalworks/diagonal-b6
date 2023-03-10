package api

import (
	"encoding/binary"
	"hash/fnv"

	"diagonal.works/b6"
)

func TileFeatureID(id b6.FeatureID) uint64 {
	h := fnv.New64()
	var buffer [8]byte
	binary.LittleEndian.PutUint64(buffer[0:], uint64(id.Type))
	h.Write(buffer[0:])
	h.Write([]byte(id.Namespace))
	binary.LittleEndian.PutUint64(buffer[0:], id.Value)
	h.Write(buffer[0:])
	return h.Sum64()
}

func TileFeatureIDForPolygon(id b6.FeatureID, polygon int) uint64 {
	h := fnv.New64()
	var buffer [8]byte
	binary.LittleEndian.PutUint64(buffer[0:], uint64(id.Type))
	h.Write(buffer[0:])
	h.Write([]byte(id.Namespace))
	binary.LittleEndian.PutUint64(buffer[0:], id.Value)
	h.Write(buffer[0:])
	binary.LittleEndian.PutUint64(buffer[0:], uint64(polygon))
	h.Write(buffer[0:])
	return h.Sum64()
}

func TileFeatureIDsForArea(area b6.AreaFeature) []uint64 {
	ids := make([]uint64, area.Len())
	for i := range ids {
		ids[i] = TileFeatureIDForPolygon(area.FeatureID(), i)
	}
	return ids
}
