package ingest

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Emit func(f Feature, goroutine int) error

type FeatureSource interface {
	Read(options ReadOptions, emit Emit, ctx context.Context) error
}

type Wait func() error

// ParalleliseEmit passes features concurrently to the given emit function,
// using the specified number of goroutines. Wait() blocks until all
// goroutines have finished emitting. This is useful to adapt a source
// that inherently uses a single goroutine (likely due to the input data
// format) to a sink that can make use of parallelism.
// The gorotuine specified in each emit call is used for that feature. If that
// goroutine is still emitting the previous feature, the emit call blocks. This
// behaviour allows a caller to use a buffer of 2*goroutines features to avoid
// dynamic allocation.
func ParalleliseEmit(emit Emit, goroutines int, ctx context.Context) (Emit, Wait) {
	if goroutines < 1 {
		goroutines = 1
	}
	c := make([]chan Feature, goroutines)
	for i := range c {
		c[i] = make(chan Feature, 0)
	}

	g, gc := errgroup.WithContext(ctx)
	for i := 0; i < goroutines; i++ {
		goroutine := i
		g.Go(func() error {
			for f := range c[goroutine] {
				if err := emit(f, goroutine); err != nil {
					return err
				}
			}
			return nil
		})
	}

	running := true
	wait := func() error {
		if running {
			for i := 0; i < goroutines; i++ {
				close(c[i])
			}
			running = false
		}
		return g.Wait()
	}
	parallelised := func(f Feature, goroutine int) error {
		if running {
			select {
			case <-gc.Done():
				return wait()
			case c[goroutine] <- f:
				return nil
			}
		} else {
			return wait()
		}
	}
	return parallelised, wait
}

// MergedSource reads from each of given Sources, reading as many in parallel as
// the number of goroutines allows. The number of goroutines available to the
// underlying Source is divided accordingly
type MergedFeatureSource []FeatureSource

func (m MergedFeatureSource) Read(options ReadOptions, emit Emit, ctx context.Context) error {
	if len(m) == 0 {
		return nil
	}
	if options.Goroutines < 1 {
		options.Goroutines = 1
	}
	perFile := options
	perFile.Goroutines /= len(m)
	if perFile.Goroutines < 1 {
		perFile.Goroutines = 1
	}
	g, gc := errgroup.WithContext(ctx)
	c := make(chan FeatureSource)
	for i := 0; i < options.Goroutines/perFile.Goroutines; i++ {
		i := i
		g.Go(func() error {
			offset := func(f Feature, g int) error {
				return emit(f, g+(i*perFile.Goroutines))
			}
			for {
				select {
				case <-gc.Done():
					return nil
				case s, ok := <-c:
					if ok {
						if err := s.Read(perFile, offset, ctx); err != nil {
							return err
						}
					} else {
						return nil
					}
				}
			}
		})
	}
	for _, s := range m {
		select {
		case <-gc.Done():
			return g.Wait()
		case c <- s:
		}
	}
	close(c)
	return g.Wait()
}
