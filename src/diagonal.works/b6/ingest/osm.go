package ingest

import (
	"context"
	"fmt"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/osm"
	"github.com/apache/beam/sdks/go/pkg/beam/io/filesystem"
	_ "github.com/apache/beam/sdks/go/pkg/beam/io/filesystem/local"
)

func FromOSMNodeID(id osm.NodeID) b6.PointID {
	return b6.MakePointID(b6.NamespaceOSMNode, uint64(id))
}

func FromOSMWayID(id osm.WayID) b6.PathID {
	return b6.MakePathID(b6.NamespaceOSMWay, uint64(id))
}

func AreaIDFromOSMWayID(id osm.WayID) b6.AreaID {
	return b6.MakeAreaID(b6.NamespaceOSMWay, uint64(id))
}

func FromOSMRelationID(id osm.RelationID) b6.RelationID {
	return b6.MakeRelationID(b6.NamespaceOSMRelation, uint64(id))
}

func AreaIDFromOSMRelationID(id osm.RelationID) b6.AreaID {
	return b6.MakeAreaID(b6.NamespaceOSMRelation, uint64(id))
}

type OSMSource interface {
	Read(options osm.ReadOptions, emit osm.EmitWithGoroutine, ctx context.Context) error
}

type PBFFilesOSMSource struct {
	Glob            string
	FailWhenNoFiles bool
}

func (s *PBFFilesOSMSource) Read(options osm.ReadOptions, emit osm.EmitWithGoroutine, ctx context.Context) error {
	read := func(ctx context.Context, filename string, fs filesystem.Interface) error {
		f, err := fs.OpenRead(ctx, filename)
		if err != nil {
			return err
		}
		defer f.Close()
		return osm.ReadPBFWithOptions(f, emit, options)
	}
	return mapFilenames(ctx, read, s.Glob, s.FailWhenNoFiles, options.Parallelism)
}

func mapFilenames(ctx context.Context, f func(ctx context.Context, filename string, fs filesystem.Interface) error, glob string, failWhenNoFiles bool, goroutines int) error {
	fs, err := filesystem.New(ctx, glob)
	if err != nil {
		return err
	}
	matches, err := fs.List(ctx, glob)
	if err != nil {
		return err
	} else if len(matches) == 0 && failWhenNoFiles {
		return fmt.Errorf("No files matched %s", glob)
	}
	return mapNWithLimit(func(i int) error { return f(ctx, matches[i], fs) }, len(matches), goroutines)
}

// TODO: Refactor
func mapNWithLimit(f func(i int) error, n int, goroutines int) error {
	wg := sync.WaitGroup{}
	wg.Add(goroutines)
	errors := make([]error, n)
	i := 0
	iLock := sync.Mutex{}
	for j := 0; j < goroutines; j++ {
		go func() {
			defer wg.Done()
			for {
				iLock.Lock()
				if i >= n {
					iLock.Unlock()
					return
				}
				run := i
				i++
				iLock.Unlock()
				errors[run] = f(run)
			}
		}()
	}
	wg.Wait()

	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

type MemoryOSMSource struct {
	Nodes     []osm.Node
	Ways      []osm.Way
	Relations []osm.Relation
}

func (s *MemoryOSMSource) Read(options osm.ReadOptions, emit osm.EmitWithGoroutine, ctx context.Context) error {
	if !options.SkipNodes {
		for _, n := range s.Nodes {
			if err := emit(&n, 0); err != nil {
				return err
			}
		}
	}
	if !options.SkipWays {
		for _, w := range s.Ways {
			if err := emit(&w, 0); err != nil {
				return err
			}
		}
	}
	if !options.SkipRelations {
		for _, r := range s.Relations {
			if err := emit(&r, 0); err != nil {
				return err
			}
		}
	}
	return nil
}

var osmTagMapping = map[string]string{
	"amenity":   "#amenity",
	"barrier":   "#barrier",
	"boundary":  "#boundary",
	"bridge":    "#bridge",
	"building":  "#building",
	"highway":   "#highway",
	"landuse":   "#landuse",
	"leisure":   "#leisure",
	"natural":   "#natural",
	"network":   "#network",
	"place":     "#place",
	"railway":   "#railway",
	"route":     "#route",
	"shop":      "#shop",
	"water":     "#water",
	"waterway":  "#waterway",
	"fhrs:id":   "@fhrs:id",
	"wikidata":  "@wikidata",
	"wikipedia": "@wikipedia",
}

func NewTagsFromOSM(o []osm.Tag) Tags {
	tags := make(Tags, len(o))
	for i, tag := range o {
		key := tag.Key
		if mapped, ok := osmTagMapping[tag.Key]; ok {
			key = mapped
		}
		tags[i] = b6.Tag{Key: key, Value: tag.Value}
	}
	return tags
}

func KeyForOSMKey(key string) string {
	if k, ok := osmTagMapping[key]; ok {
		return k
	}
	return key
}

func NewWorldFromPBFFile(filename string, cores int, validity FeatureValidity) (b6.World, error) {
	source := PBFFilesOSMSource{Glob: filename, FailWhenNoFiles: true}
	return NewWorldFromOSMSource(&source, cores, validity)
}

func BuildWorldFromOSM(nodes []osm.Node, ways []osm.Way, relations []osm.Relation, cores int, validity FeatureValidity) (b6.World, error) {
	source := MemoryOSMSource{Nodes: nodes, Ways: ways, Relations: relations}
	return NewWorldFromOSMSource(&source, cores, validity)
}

func NewWorldFromOSMSource(o OSMSource, cores int, validity FeatureValidity) (b6.World, error) {
	source, err := NewFeatureSourceFromPBF(o, cores, context.Background())
	if err != nil {
		return nil, err
	}
	return NewWorldFromSource(source, cores, validity)
}

func NewMutableWorldFromOSMSource(o OSMSource, cores int, validity FeatureValidity) (b6.World, error) {
	source, err := NewFeatureSourceFromPBF(o, cores, context.Background())
	if err != nil {
		return nil, err
	}
	return NewMutableWorldFromSource(source, cores, validity)
}

func BuildMutableWorldFromOSM(nodes []osm.Node, ways []osm.Way, relations []osm.Relation, cores int, validity FeatureValidity) (b6.World, error) {
	source := MemoryOSMSource{Nodes: nodes, Ways: ways, Relations: relations}
	return NewWorldFromOSMSource(&source, cores, validity)
}
