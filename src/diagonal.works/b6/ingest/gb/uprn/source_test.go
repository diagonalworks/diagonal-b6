package uprn

import (
	"context"
	"reflect"
	"strconv"
	"sync"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	"github.com/golang/geo/s2"
)

func TestUPRNClusters(t *testing.T) {
	uprns := []ingest.Feature{
		&ingest.GenericFeature{
			ID:   b6.FeatureID{b6.FeatureTypePoint, b6.NamespaceGBUPRN, 5150460},
			Tags: []b6.Tag{{Key: b6.PointTag, Value: b6.LatLng(s2.LatLngFromDegrees(51.5349035, -0.1257194))}},
		},
		&ingest.GenericFeature{
			ID:   b6.FeatureID{b6.FeatureTypePoint, b6.NamespaceGBUPRN, 5150461},
			Tags: []b6.Tag{{Key: b6.PointTag, Value: b6.LatLng(s2.LatLngFromDegrees(51.5349035, -0.1257194))}},
		},
		&ingest.GenericFeature{
			ID:   b6.FeatureID{b6.FeatureTypePoint, b6.NamespaceGBUPRN, 5158495},
			Tags: []b6.Tag{{Key: b6.PointTag, Value: b6.LatLng(s2.LatLngFromDegrees(51.536685, -0.127258))}},
		}}

	source := ClusterSource{
		UPRNs: ingest.MemoryFeatureSource(uprns),
	}
	clusters := make(map[uint64]int)
	var lock sync.Mutex
	emit := func(f ingest.Feature, goroutine int) error {
		lock.Lock()
		defer lock.Unlock()
		clusters[f.FeatureID().Value], _ = strconv.Atoi(f.Get("uprn_cluster:size").Value.String())
		return nil
	}
	expected := map[uint64]int{
		5221390606888338432: 1,
		5221390769366334464: 2,
	}
	err := source.Read(ingest.ReadOptions{Goroutines: 4}, emit, context.Background())
	if err == nil {
		if !reflect.DeepEqual(expected, clusters) {
			t.Errorf("Expected %+v clusters, found %+v", expected, clusters)
		}
	} else {
		t.Errorf("Expected no error, found %v", err)
	}
}
