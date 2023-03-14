package compact

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"diagonal.works/b6"
	"diagonal.works/b6/encoding"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/osm"

	"github.com/apache/beam/sdks/go/pkg/beam/io/filesystem"
)

type beamPBFSouce struct {
	Filename string
	FS       filesystem.Interface
	Ctx      context.Context
}

func (s *beamPBFSouce) Read(options osm.ReadOptions, emit osm.EmitWithGoroutine, ctx context.Context) error {
	r, err := s.FS.OpenRead(s.Ctx, s.Filename)
	if err != nil {
		return err
	}
	defer r.Close()
	return osm.ReadPBFWithOptions(r, emit, options)
}

func ReadWorld(input string, cores int) (b6.World, error) {
	ctx := context.Background()
	sources := strings.Split(input, ",")
	expanded := make([]string, 0, len(sources))
	isRegion := make([]bool, 0, len(sources))
	fss := make([]filesystem.Interface, 0, len(sources))
	close := make([]io.Closer, len(sources))
	for i, s := range sources {
		var hasRegionPrefix, hasOSMPRefix bool
		if hasRegionPrefix = strings.HasPrefix(s, "region:"); hasRegionPrefix {
			s = strings.TrimPrefix(s, "region:")
		} else if hasOSMPRefix = strings.HasPrefix(s, "osm:"); hasOSMPRefix {
			s = strings.TrimPrefix(s, "osm:")
		}
		fs, err := filesystem.New(ctx, s)
		if err != nil {
			return nil, err
		}
		close[i] = fs
		children, err := fs.List(ctx, s+"/*")
		if err != nil {
			return nil, err
		}
		if len(children) > 0 {
			for _, c := range children {
				expanded = append(expanded, c)
				fss = append(fss, fs)
				isRegion = append(isRegion, (hasRegionPrefix || !strings.HasSuffix(c, ".pbf")) && !hasOSMPRefix)
			}
		} else {
			expanded = append(expanded, s)
			fss = append(fss, fs)
			isRegion = append(isRegion, (hasRegionPrefix || !strings.HasSuffix(s, ".pbf")) && !hasOSMPRefix)
		}
	}

	worlds := make([]b6.World, 0)
	rw := NewWorld()
	for i, s := range expanded {
		if isRegion[i] {
			m, err := encoding.Mmap(s)
			if err == nil {
				log.Printf("Memory map %s", s)
				if err := rw.Merge(m.Data); err != nil {
					return nil, err
				}
			} else {
				log.Printf("Read %s", s)
				r, err := fss[i].OpenRead(ctx, s)
				if err != nil {
					return nil, err
				}
				defer r.Close()
				// TODO: Parallelise reading
				data, err := ioutil.ReadAll(r)
				if err != nil {
					return nil, err
				}
				if err := rw.Merge(data); err != nil {
					return nil, err
				}
			}
		} else {
			log.Printf("Index PBF %s", s)
			pbf := beamPBFSouce{
				Filename: s,
				FS:       fss[i],
				Ctx:      ctx,
			}
			w, err := ingest.NewWorldFromOSMSource(&pbf, cores, ingest.DeleteInvalidFeatures)
			if err != nil {
				return nil, err
			}
			worlds = append(worlds, w)
		}
	}

	overlay := b6.World(rw)
	for i := len(worlds) - 1; i >= 0; i-- {
		overlay = ingest.NewOverlayWorld(worlds[i], overlay)
	}

	for _, c := range close {
		c.Close()
	}
	return overlay, nil
}
