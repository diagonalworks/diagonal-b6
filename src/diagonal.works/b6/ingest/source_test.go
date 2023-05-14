package ingest

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"diagonal.works/b6"
)

func TestParalliseEmit(t *testing.T) {
	const cores = 8
	const nIDs = 1024
	const ns = b6.Namespace("diagonal.works/ns/test")

	emitted := make([][]b6.FeatureID, cores)
	emit := func(f Feature, goroutine int) error {
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
		emitted[goroutine] = append(emitted[goroutine], f.FeatureID())
		return nil
	}

	ps := make([]PointFeature, cores*2)
	parallelised, wait := ParalleliseEmit(emit, cores, context.Background())
	for i := 0; i < nIDs; i++ {
		slot := i % len(ps)
		ps[slot].PointID = b6.MakePointID(ns, uint64(i))
		if err := parallelised(&ps[slot], i%cores); err != nil {
			break
		}
	}
	if err := wait(); err != nil {
		t.Errorf("Expected no error, found %s", err)
		return
	}

	expected := make(map[b6.FeatureID]struct{})
	for i := 0; i < nIDs; i++ {
		expected[b6.MakePointID(ns, uint64(i)).FeatureID()] = struct{}{}
	}
	for _, ids := range emitted {
		for _, id := range ids {
			delete(expected, id)
		}
	}
	if len(expected) != 0 {
		t.Errorf("Expected all ids to be emitted")
	}
}

func TestParalliseEmitWithError(t *testing.T) {
	const cores = 8
	const nIDs = 1024
	const ns = b6.Namespace("diagonal.works/ns/test")

	broken := errors.New("Broken")
	emit := func(f Feature, goroutine int) error {
		if f.FeatureID().Value == 42 {
			return broken
		}
		return nil
	}

	ps := make([]PointFeature, cores*2)
	parallelised, wait := ParalleliseEmit(emit, cores, context.Background())
	for i := 0; i < nIDs; i++ {
		slot := i % len(ps)
		ps[slot].PointID = b6.MakePointID(ns, uint64(i))
		if err := parallelised(&ps[slot], i%cores); err != nil {
			break
		}
	}
	if err := wait(); err != broken {
		t.Errorf("Expected %s, found %s", broken, err)
		return
	}
}
