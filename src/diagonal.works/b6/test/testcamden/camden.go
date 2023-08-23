package testcamden

import (
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

func BuildGranarySquare(t *testing.T) b6.World {
	granarySquareLock.Lock()
	defer granarySquareLock.Unlock()
	if granarySquare == nil {
		granarySquare = build(t, test.Data(test.GranarySquarePBF))
	}
	return granarySquare
}

func BuildCamden(t *testing.T) b6.World {
	camdenLock.Lock()
	defer camdenLock.Unlock()
	if camden == nil {
		camden = build(t, test.Data(test.CamdenPBF))
	}
	return camden
}

func build(t *testing.T, filename string) b6.World {
	o := &ingest.BuildOptions{Cores: 2}
	w, err := ingest.NewWorldFromPBFFile(filename, o)
	if err != nil {
		t.Fatalf("Failed to build world: %v", err)
	}
	return w
}
