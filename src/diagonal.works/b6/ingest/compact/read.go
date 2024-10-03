package compact

import (
	"os"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"runtime"
	"sort"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/encoding"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/osm"
	"golang.org/x/sync/errgroup"

	"github.com/apache/beam/sdks/go/pkg/beam/io/filesystem"
)

type pbfSource struct {
	Filename string
	FS       filesystem.Interface
	Ctx      context.Context
}

func (s *pbfSource) Read(options osm.ReadOptions, emit osm.EmitWithGoroutine, ctx context.Context) error {
	r, err := s.FS.OpenRead(s.Ctx, s.Filename)
	if err != nil {
		return err
	}
	defer r.Close()
	return osm.ReadPBFWithOptions(r, emit, options)
}

type readFormat int

const (
	readFormatCompact readFormat = iota
	readFormatOSM
	readFormatYAML
)

type toRead struct {
	Filename   string
	Filesystem filesystem.Interface
	Format     readFormat

	// Outputs
	World  b6.World
	Change ingest.Change
}

func (t *toRead) Read(w *World, status chan<- string, o *ingest.BuildOptions, ctx context.Context) error {
	switch t.Format {
	case readFormatCompact:
		m, err := encoding.Mmap(t.Filename)
		if err == nil {
			status <- fmt.Sprintf("Memory map %s", t.Filename)
			return w.Merge(m.Data)
		} else {
			status <- fmt.Sprintf("Read %s", t.Filename)
			m, err := encoding.ReadToMmappedBuffer(t.Filename, t.Filesystem, ctx, status)
			if err == nil {
				err = w.Merge(m.Data)
			}
			return err
		}
	case readFormatOSM:
		status <- fmt.Sprintf("Index PBF %s", t.Filename)
		pbf := pbfSource{
			Filename: t.Filename,
			FS:       t.Filesystem,
			Ctx:      ctx,
		}
		var err error
		t.World, err = ingest.NewWorldFromOSMSource(&pbf, o)
		return err
	case readFormatYAML:
		status <- fmt.Sprintf("Read YAML %s", t.Filename)
		f, err := t.Filesystem.OpenRead(ctx, t.Filename)
		if err == nil {
			var buffer bytes.Buffer
			if _, err = io.Copy(&buffer, f); err == nil {
				t.Change = ingest.IngestChangesFromYAML(&buffer)
			}
			f.Close()
		}
		return err
	}
	return fmt.Errorf("bad read format for %s", t.Filename)
}

func ReadWorld(input string, o *ingest.BuildOptions) (b6.World, error) {
	// Run the new function but with no additional world inputs.
	additional := make(map[b6.FeatureID]string)
	w, _, err := ReadWorldAndAdditionals(input, o, additional)
	return w, err
}

func ReadWorldAndAdditionals(input string, o *ingest.BuildOptions, additionals map[b6.FeatureID]string) (b6.World, map[b6.FeatureID]ingest.MutableWorld, error) {
	ctx := context.Background()
	sources := strings.Split(input, ",")
	trs := make([]*toRead, 0)
	toClose := make([]io.Closer, len(sources))

	for i, s := range sources {
		s = includeVersion(s)
		var hasCompactPrefix, hasOSMPRefix bool
		if hasCompactPrefix = strings.HasPrefix(s, "compact:"); hasCompactPrefix {
			s = strings.TrimPrefix(s, "compact:")
		} else if hasOSMPRefix = strings.HasPrefix(s, "osm:"); hasOSMPRefix {
			s = strings.TrimPrefix(s, "osm:")
		}
		fs, err := filesystem.New(ctx, s)
		if err != nil {
			return nil, nil, err
		}
		toClose[i] = fs
		var children []string
		if strings.Contains(s, "*") {
			children, err = fs.List(ctx, s)
		} else {
			children, err = fs.List(ctx, s+"/*")
			if len(children) == 0 {
				children = []string{s}
			}
		}
		if err != nil {
			return nil, nil, err
		}
		sort.Strings(children)
		for _, child := range children {
			format := readFormatCompact
			if hasCompactPrefix {
				format = readFormatCompact
			} else if hasOSMPRefix {
				format = readFormatOSM
			} else if strings.HasSuffix(child, ".pbf") {
				format = readFormatOSM
			} else if strings.HasSuffix(child, ".yaml") {
				format = readFormatYAML
			}
			trs = append(trs, &toRead{
				Filename:   child,
				Filesystem: fs,
				Format:     format,
			})
		}
	}


	cw := NewWorld()
	status := make(chan string)

	g, _ := errgroup.WithContext(ctx)
	for i := range trs {
		tr := trs[i]
		g.Go(func() error {
			return tr.Read(cw, status, o, ctx)
		})
	}
	go func() {
		for s := range status {
			log.Println(s)
		}
	}()

	err := g.Wait()
	close(status)
	if err != nil {
		return nil, nil, err
	}

	overlay := b6.World(cw)
	for _, tr := range trs {
		if tr.World != nil {
			log.Printf("Overlay %s", tr.Filename)
			overlay = ingest.NewOverlayWorld(tr.World, overlay)
		}
	}

	changed := false
	m := ingest.NewMutableOverlayWorld(overlay)
	for _, tr := range trs {
		if tr.Change != nil {
			changed = true
			log.Printf("Apply %s", tr.Filename)
			if _, err := tr.Change.Apply(m); err != nil {
				return nil, nil, fmt.Errorf("%s: %w", tr.Filename, err)
			}
			tr.Change = nil
			runtime.GC()
		}
	}

	for _, c := range toClose {
		c.Close()
	}

	additionalMutableWorlds := make(map[b6.FeatureID]ingest.MutableWorld)

	mkAdditionals := func (w b6.World) error {
		for featureId, worldStr := range additionals {
			world, nil := ReadWorld(worldStr, o)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				return err
			}
			log.Printf("Adding new world at %s", featureId)
			additionalMutableWorlds[featureId] = ingest.NewMutableOverlayWorld(ingest.NewOverlayWorld(world, m))
		}

		return nil
	}

	if changed {
		err = mkAdditionals(m)
		if err != nil {
			return nil, nil, err
		}
		return m, additionalMutableWorlds, nil
	} else {
		err = mkAdditionals(overlay)
		if err != nil {
			return nil, nil, err
		}
		return overlay, additionalMutableWorlds, nil
	}
}
