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

func FromOSMNodeID(id osm.NodeID) b6.FeatureID {
	return b6.FeatureID{b6.FeatureTypePoint, b6.NamespaceOSMNode, uint64(id)}
}

func FromOSMWayID(id osm.WayID) b6.FeatureID {
	return b6.FeatureID{b6.FeatureTypePath, b6.NamespaceOSMWay, uint64(id)}
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
	return mapFilenames(ctx, read, s.Glob, s.FailWhenNoFiles, options.Cores)
}

func mapFilenames(ctx context.Context, f func(ctx context.Context, filename string, fs filesystem.Interface) error, glob string, failWhenNoFiles bool, goroutines int) error {
	if goroutines < 1 {
		goroutines = 1
	}
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
	"tourism":   "#tourism",
	"water":     "#water",
	"waterway":  "#waterway",
	"fhrs:id":   "@fhrs:id",
	"wikidata":  "@wikidata",
	"wikipedia": "@wikipedia",
}

func NewTagsFromOSM(o osm.Tags) b6.Tags {
	tags := make(b6.Tags, 0, len(o))
	FillTagsFromOSM(&tags, o)
	return tags
}

func FillTagsFromOSM(t *b6.Tags, o osm.Tags) {
	*t = (*t)[0:0]
	for _, tag := range o {
		key := tag.Key
		if mapped, ok := osmTagMapping[tag.Key]; ok {
			key = mapped
		}
		*t = append(*t, b6.Tag{Key: key, Value: b6.String(tag.Value)})
	}
}

func KeyForOSMKey(key string) string {
	if k, ok := osmTagMapping[key]; ok {
		return k
	}
	return key
}

func NewWorldFromPBFFile(filename string, o *BuildOptions) (b6.World, error) {
	source := PBFFilesOSMSource{Glob: filename, FailWhenNoFiles: true}
	return NewWorldFromOSMSource(&source, o)
}

func BuildWorldFromOSM(nodes []osm.Node, ways []osm.Way, relations []osm.Relation, o *BuildOptions) (b6.World, error) {
	source := MemoryOSMSource{Nodes: nodes, Ways: ways, Relations: relations}
	return NewWorldFromOSMSource(&source, o)
}

func NewWorldFromOSMSource(s OSMSource, o *BuildOptions) (b6.World, error) {
	source, err := NewFeatureSourceFromPBF(s, o, context.Background())
	if err != nil {
		return nil, err
	}
	return NewWorldFromSource(source, o)
}

func NewMutableWorldFromOSMSource(s OSMSource, o *BuildOptions) (b6.World, error) {
	source, err := NewFeatureSourceFromPBF(s, o, context.Background())
	if err != nil {
		return nil, err
	}
	return NewMutableWorldFromSource(o, source)
}

func BuildMutableWorldFromOSM(nodes []osm.Node, ways []osm.Way, relations []osm.Relation, o *BuildOptions) (b6.World, error) {
	source := MemoryOSMSource{Nodes: nodes, Ways: ways, Relations: relations}
	return NewWorldFromOSMSource(&source, o)
}

func isWayClosed(way *osm.Way) bool {
	return way.Nodes[0] == way.Nodes[len(way.Nodes)-1]
}

func isRelationArea(relation *osm.Relation) bool {
	if t, ok := relation.Tag("type"); ok {
		return t == "multipolygon"
	}
	return false
}

type pbfSource struct {
	pbf              OSMSource
	areaWays         *IDSet                // IDs of ways that represent areas
	areaRelations    *IDSet                // IDs of relations that represent areas
	multipolygonWays map[osm.WayID]osm.Way // Ways that aren't closed, but are referenced by multipolygons
}

// NewFeatureSourceFromPBF returns a FeatureSource with features from pbf.
// Note that during the creation of the source, all ways and relations are
// read, as we need to know in advance which of them represent Areas,
// so we can correctly build FeatureIDs from an OSM ID reference.
func NewFeatureSourceFromPBF(pbf OSMSource, o *BuildOptions, ctx context.Context) (FeatureSource, error) {
	s := &pbfSource{
		pbf:              pbf,
		areaWays:         NewIDSet(),
		areaRelations:    NewIDSet(),
		multipolygonWays: make(map[osm.WayID]osm.Way),
	}

	multipolygonWays := NewIDSet()
	emit := func(element osm.Element, g int) error {
		switch e := element.(type) {
		case *osm.Relation:
			if isRelationArea(e) {
				for _, m := range e.Members {
					if m.Type == osm.ElementTypeWay {
						multipolygonWays.Add(uint64(m.ID))
					}
				}
			}
		}
		return nil
	}

	options := osm.ReadOptions{SkipNodes: true, SkipWays: true, Cores: o.Cores}
	if err := pbf.Read(options, emit, ctx); err != nil {
		return nil, err
	}

	var lock sync.Mutex
	emit = func(element osm.Element, g int) error {
		switch e := element.(type) {
		case *osm.Way:
			if isWayClosed(e) {
				s.areaWays.Add(uint64(e.ID))
			} else if multipolygonWays.Has(uint64(e.ID)) {
				lock.Lock()
				s.multipolygonWays[e.ID] = e.Clone()
				lock.Unlock()
			}
		case *osm.Relation:
			if isRelationArea(e) {
				s.areaRelations.Add(uint64(e.ID))
			}
		}
		return nil
	}
	options = osm.ReadOptions{SkipNodes: true, Cores: o.Cores}
	if err := pbf.Read(options, emit, ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func reassembleMultiPolygon(relation *osm.Relation, areaWays *IDSet, ways map[osm.WayID]osm.Way, goroutine int, emit Emit) {
	polygons := make([][]osm.WayID, 0)
	loops := make([]osm.WayID, 0)
	for _, m := range relation.Members {
		if m.Type == osm.ElementTypeWay {
			if m.Role == "outer" || m.Role == "" {
				// TODO: use existing multipolygon assembly code rather than relying on
				// outer/inner tags and ordering.
				if len(loops) > 0 {
					polygons = append(polygons, loops)
					loops = make([]osm.WayID, 0)
				}
			}
			if areaWays.Has(uint64(m.ID)) {
				loops = append(loops, osm.WayID(m.ID))
			} else {
				// This could be because the way isn't closed, or because the way doesn't
				// fall within the area we're looking at
				// TODO: Reassemble polygons from unclosed ways
				return
			}
		}
	}
	if len(loops) > 0 {
		polygons = append(polygons, loops)
	}
	area := NewAreaFeature(len(polygons))
	FillTagsFromOSM(&area.Tags, relation.Tags)
	area.AreaID = AreaIDFromOSMRelationID(relation.ID)
	for i, loops := range polygons {
		ids := make([]b6.FeatureID, len(loops))
		for j, loop := range loops {
			ids[j] = FromOSMWayID(loop)
		}
		area.SetPathIDs(i, ids)
	}
	emit(area, goroutine)
}

func (s *pbfSource) Read(options ReadOptions, emit Emit, ctx context.Context) error {
	cores := options.Goroutines
	if cores < 1 {
		cores = 1
	}
	o := osm.ReadOptions{
		SkipTags:      options.SkipTags,
		SkipNodes:     options.SkipPoints,
		SkipWays:      options.SkipPaths && options.SkipAreas,
		SkipRelations: options.SkipRelations && options.SkipAreas,
		Cores:         cores,
	}
	points := make([]GenericFeature, cores)
	paths := make([]GenericFeature, cores)
	areas := make([]AreaFeature, cores)
	relations := make([]RelationFeature, cores)
	f := func(element osm.Element, g int) error {
		switch e := element.(type) {
		case *osm.Node:
			points[g].FillFromOSM(OSMFeature{Node: e})
			return emit(&points[g], g)
		case *osm.Way:
			if !options.SkipPaths {
				paths[g].FillFromOSM(OSMFeature{Way: e, ClosedWay: isWayClosed(e)})
				if err := emit(&paths[g], g); err != nil {
					return err
				}
			}

			if isWayClosed(e) && !options.SkipAreas {
				areas[g].FillFromOSMWay(e)
				return emit(&areas[g], g)
			}
		case *osm.Relation:
			if isRelationArea(e) {
				if !options.SkipAreas {
					reassembleMultiPolygon(e, s.areaWays, s.multipolygonWays, g, emit)
				}
			} else if !options.SkipRelations {
				relations[g].RelationID = FromOSMRelationID(e.ID)
				FillTagsFromOSM(&relations[g].Tags, e.Tags)
				relations[g].Members = relations[g].Members[0:0]
				for _, m := range e.Members {
					var id b6.FeatureID
					switch m.Type {
					case osm.ElementTypeNode:
						id = FromOSMNodeID(m.NodeID()).FeatureID()
					case osm.ElementTypeWay:
						if s.areaWays.Has(uint64(e.ID)) {
							id = AreaIDFromOSMWayID(m.WayID()).FeatureID()
						} else {
							id = FromOSMWayID(m.WayID()).FeatureID()
						}
					case osm.ElementTypeRelation:
						if s.areaRelations.Has(uint64(e.ID)) {
							id = AreaIDFromOSMRelationID(m.RelationID()).FeatureID()
						} else {
							id = FromOSMRelationID(m.RelationID()).FeatureID()
						}
					}
					relations[g].Members = append(relations[g].Members, b6.RelationMember{ID: id, Role: m.Role})
				}
				return emit(&relations[g], g)
			}
		}
		return nil
	}
	return s.pbf.Read(o, f, ctx)
}
