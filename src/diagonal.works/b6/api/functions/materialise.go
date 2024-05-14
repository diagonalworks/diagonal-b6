package functions

import (
	"fmt"
	"math/rand"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	"golang.org/x/sync/errgroup"
)

func materialiseCollection(id b6.CollectionID, collection b6.UntypedCollection) (*ingest.CollectionFeature, error) {
	c := b6.AdaptCollection[any, any](collection)
	keys, err := c.AllKeys(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fill keys: %s", err)
	}

	values, err := c.AllValues(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fill values: %s", err)
	}

	return &ingest.CollectionFeature{
		CollectionID: id,
		Keys:         keys,
		Values:       values,
	}, nil
}

// Return a change that adds a collection feature to the world with the given ID, containing the result of calling the given function.
// The given function isn't passed any arguments.
// Also adds an expression feature (with the same namespace and value)
// representing the given function.
func materialise(context *api.Context, id b6.CollectionID, function api.Callable) (ingest.Change, error) {
	if function.NumArgs() != 0 {
		return nil, fmt.Errorf("expected a function with no arguments, found %d", function.NumArgs())
	}

	result, err := context.VM.CallWithArgs(context, function, []interface{}{})
	if err != nil {
		return nil, err
	}

	untyped, ok := result.(b6.UntypedCollection)
	if !ok {
		return nil, fmt.Errorf("expected a collection, found %T", result)
	}
	collection, err := materialiseCollection(id, untyped)
	if err != nil {
		return nil, err
	}

	expression := ingest.ExpressionFeature{
		ExpressionID: b6.ExpressionID(id),
		Expression:   function.Expression(),
	}

	add := ingest.AddFeatures([]ingest.Feature{&expression, collection})
	return &add, nil
}

func materialiseMap(context *api.Context, collection b6.Collection[any, b6.Feature], id b6.CollectionID, function api.Callable) (ingest.Change, error) {
	if function.NumArgs() != 1 {
		return nil, fmt.Errorf("expected a function with one argument, found %d", function.NumArgs())
	}

	cores := context.Cores
	if cores == 0 {
		cores = 1
	}

	expression := function.Expression()
	results := &ingest.CollectionFeature{
		CollectionID: id,
	}
	change := ingest.AddFeatures{results}

	var lock sync.Mutex
	seen := make(map[b6.FeatureID]struct{})
	g, c := errgroup.WithContext(context.Context)
	contexts := context.Fork(cores)
	in := make(chan b6.Feature)
	for i := range contexts {
		context := contexts[i]
		g.Go(func() error {
			for f := range in {
				result, err := context.VM.CallWithArgs(context, function, []interface{}{f})
				if err != nil {
					return err
				}
				untyped, ok := result.(b6.UntypedCollection)
				if !ok {
					return fmt.Errorf("expected a collection, found %T", result)
				}
				id := b6.CollectionID{Namespace: b6.NamespaceMaterialised, Value: rand.Uint64()}
				materialised, err := materialiseCollection(id, untyped)
				if err != nil {
					return err
				}
				var bound *ingest.ExpressionFeature
				if expression.IsValid() {
					bound = &ingest.ExpressionFeature{
						ExpressionID: b6.ExpressionID(id),
						Expression: b6.NewCallExpression(
							expression,
							[]b6.Expression{
								b6.NewCallExpression(
									b6.NewSymbolExpression("find-feature"),
									[]b6.Expression{
										b6.NewFeatureIDExpression(f.FeatureID()),
									},
								),
							},
						),
					}
				}
				lock.Lock()
				results.Keys = append(results.Keys, f.FeatureID())
				results.Values = append(results.Values, materialised.FeatureID())
				change = append(change, materialised)
				if bound != nil {
					change = append(change, bound)
				}
				seen[f.FeatureID()] = struct{}{}
				lock.Unlock()
			}
			return nil
		})
	}

	g.Go(func() error {
		i := collection.Begin()
		for {
			ok, err := i.Next()
			if !ok || err != nil {
				close(in)
				return err
			}
			select {
			case in <- i.Value():
			case <-c.Done():
				close(in)
				return c.Err()
			}
		}
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	if existing := b6.FindCollectionByID(id, context.World); existing != nil {
		c := b6.AdaptCollection[b6.FeatureID, any](existing)
		i := c.Begin()
		for {
			ok, err := i.Next()
			if err != nil {
				return nil, err
			} else if !ok {
				break
			}
			if _, ok := seen[i.Key()]; !ok {
				results.Keys = append(results.Keys, i.Key())
				results.Values = append(results.Values, i.Value())
			}
		}
	}

	return &change, nil
}
