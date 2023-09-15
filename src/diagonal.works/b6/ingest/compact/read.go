package compact

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

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
}

func (t toRead) Read(w *World, status chan<- string, cores int, ctx context.Context) (b6.World, ingest.Change, error) {
	switch t.Format {
	case readFormatCompact:
		m, err := encoding.Mmap(t.Filename)
		if err == nil {
			status <- fmt.Sprintf("Memory map %s", t.Filename)
			return nil, nil, w.Merge(m.Data)
		} else {
			status <- fmt.Sprintf("Read %s", t.Filename)
			m, err := encoding.ReadToMmappedBuffer(t.Filename, t.Filesystem, ctx, status)
			if err == nil {
				w.Merge(m.Data)
			}
			return nil, nil, err
		}
	case readFormatOSM:
		status <- fmt.Sprintf("Index PBF %s", t.Filename)
		pbf := pbfSource{
			Filename: t.Filename,
			FS:       t.Filesystem,
			Ctx:      ctx,
		}
		o := ingest.BuildOptions{Cores: cores}
		w, err := ingest.NewWorldFromOSMSource(&pbf, &o)
		return w, nil, err
	case readFormatYAML:
		status <- fmt.Sprintf("Overlay YAML %s", t.Filename)
		f, err := t.Filesystem.OpenRead(ctx, t.Filename)
		return nil, ingest.IngestChangesFromYAML(f), err
	}
	return nil, nil, fmt.Errorf("bad read format for %s", t.Filename)
}

func ReadWorld(input string, cores int) (b6.World, error) {
	ctx := context.Background()
	sources := strings.Split(input, ",")
	trs := make([]toRead, 0)
	toClose := make([]io.Closer, len(sources))

	for i, s := range sources {
		var hasCompactPrefix, hasOSMPRefix bool
		if hasCompactPrefix = strings.HasPrefix(s, "compact:"); hasCompactPrefix {
			s = strings.TrimPrefix(s, "compact:")
		} else if hasOSMPRefix = strings.HasPrefix(s, "osm:"); hasOSMPRefix {
			s = strings.TrimPrefix(s, "osm:")
		}
		fs, err := filesystem.New(ctx, s)
		if err != nil {
			return nil, err
		}
		toClose[i] = fs
		children, err := fs.List(ctx, s+"/*")
		if err != nil {
			return nil, err
		}
		if len(children) == 0 {
			children = []string{s}
		}
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
			trs = append(trs, toRead{
				Filename:   child,
				Filesystem: fs,
				Format:     format,
			})
		}
	}

	worlds := make([]b6.World, 0)
	changes := make([]ingest.Change, 0)
	cw := NewWorld()
	status := make(chan string)
	var lock sync.Mutex

	g, _ := errgroup.WithContext(ctx)
	for i := range trs {
		tr := trs[i]
		g.Go(func() error {
			w, c, err := tr.Read(cw, status, cores, ctx)
			if err == nil {
				if w != nil {
					lock.Lock()
					worlds = append(worlds, w)
					lock.Unlock()
				}
				if c != nil {
					lock.Lock()
					changes = append(changes, c)
					lock.Unlock()
				}
			}
			return err
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
		return nil, err
	}

	overlay := b6.World(cw)
	for i := len(worlds) - 1; i >= 0; i-- {
		overlay = ingest.NewOverlayWorld(worlds[i], overlay)
	}

	for _, c := range toClose {
		c.Close()
	}

	if len(changes) > 0 {
		m := ingest.NewMutableOverlayWorld(overlay)
		for _, c := range changes {
			if _, err := c.Apply(m); err != nil {
				return nil, err
			}
		}
		return m, nil
	} else {
		return overlay, nil
	}
}
