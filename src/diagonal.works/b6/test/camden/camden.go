package camden

import (
	"fmt"
	"sync"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/osm"
	"diagonal.works/b6/test"
)

const (
	StableStreetBridgeNorthEndNode osm.NodeID     = 1447052073
	StableStreetBridgeSouthEndNode osm.NodeID     = 1540349979
	StableStreetBridgeWay          osm.WayID      = 140633010
	SomersTownBridgeEastGateNode   osm.NodeID     = 4966136630
	VermuteriaNode                 osm.NodeID     = 6082053666
	CoalDropsYardEnclosureWay      osm.WayID      = 500008118
	CoalDropsYardWestBuildingWay   osm.WayID      = 222021572
	LightermanWay                  osm.WayID      = 427900370
	LightermanEntranceNode         osm.NodeID     = 4270651271
	GasholdersRelation             osm.RelationID = 7972217
	GranarySquareWay               osm.WayID      = 222021571
	GranarySquareBikeParkingNode   osm.NodeID     = 2713224957

	BuildingsInGranarySquare = 13
)

// TODO: Move the important features out into a separate configuration file, and use that to
// validate ID stability between OSM imports.
var (
	StableStreetBridgeNorthEndID = ingest.FromOSMNodeID(StableStreetBridgeNorthEndNode)
	StableStreetBridgeSouthEndID = ingest.FromOSMNodeID(StableStreetBridgeSouthEndNode)
	StableStreetBridgeID         = ingest.FromOSMWayID(StableStreetBridgeWay)
	SomersTownBridgeEastGateID   = ingest.FromOSMNodeID(SomersTownBridgeEastGateNode)
	VermuteriaID                 = ingest.FromOSMNodeID(VermuteriaNode)
	CoalDropsYardEnclosureID     = ingest.AreaIDFromOSMWayID(CoalDropsYardEnclosureWay)
	LightermanID                 = ingest.AreaIDFromOSMWayID(LightermanWay)
	GranarySquareID              = ingest.AreaIDFromOSMWayID(GranarySquareWay)
	GranarySquareBikeParkingID   = ingest.FromOSMNodeID(GranarySquareBikeParkingNode)
)

var (
	granarySquare     b6.World
	granarySquareLock sync.Mutex

	camden     b6.World
	camdenLock sync.Mutex
)

func BuildGranarySquareForTests(t *testing.T) b6.World {
	granarySquareLock.Lock()
	defer granarySquareLock.Unlock()
	if granarySquare == nil {
		granarySquare = build(test.Data(test.GranarySquarePBF), t)
	}
	return granarySquare
}

func BuildCamdenForTests(t *testing.T) b6.World {
	camdenLock.Lock()
	defer camdenLock.Unlock()
	if camden == nil {
		camden = build(test.Data(test.CamdenPBF), t)
	}
	return camden
}

func build(filename string, t *testing.T) b6.World {
	w, err := ingest.NewWorldFromPBFFile(filename, 2, ingest.FailOnInvalidFeatures)
	if err != nil {
		if t != nil {
			t.Errorf("Failed to build world: %s", err)
		} else {
			panic(fmt.Sprintf("Failed to build world: %s", err))
		}
	}
	return w
}
