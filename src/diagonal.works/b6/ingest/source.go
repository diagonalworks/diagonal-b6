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
