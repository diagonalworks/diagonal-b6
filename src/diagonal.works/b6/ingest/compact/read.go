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

type toRead struct {
	Filename   string
	Filesystem filesystem.Interface
	IsCompact  bool
}

func (t toRead) Read(w *World, status chan<- string, cores int, ctx context.Context) (b6.World, error) {
	if t.IsCompact {
		m, err := encoding.Mmap(t.Filename)
		if err == nil {
			status <- fmt.Sprintf("Memory map %s", t.Filename)
			return nil, w.Merge(m.Data)
		} else {
			status <- fmt.Sprintf("Read %s", t.Filename)
			m, err := encoding.ReadToMmappedBuffer(t.Filename, t.Filesystem, ctx, status)
			if err == nil {
				w.Merge(m.Data)
			}
			return nil, err
		}
	} else {
		status <- fmt.Sprintf("Index PBF %s", t.Filename)
		pbf := pbfSource{
			Filename: t.Filename,
			FS:       t.Filesystem,
			Ctx:      ctx,
		}
		o := ingest.BuildOptions{Cores: cores}
		return ingest.NewWorldFromOSMSource(&pbf, &o)
	}
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
			trs = append(trs, toRead{
				Filename:   child,
				Filesystem: fs,
				IsCompact:  (hasCompactPrefix || !strings.HasSuffix(child, ".pbf")) && !hasOSMPRefix,
			})
		}
	}

	worlds := make([]b6.World, 0)
	cw := NewWorld()
	status := make(chan string)
	var lock sync.Mutex

	g, _ := errgroup.WithContext(ctx)
	for i := range trs {
		tr := trs[i]
		g.Go(func() error {
			w, err := tr.Read(cw, status, cores, ctx)
			if err == nil && w != nil {
				lock.Lock()
				worlds = append(worlds, w)
				lock.Unlock()
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
	return overlay, nil
}
