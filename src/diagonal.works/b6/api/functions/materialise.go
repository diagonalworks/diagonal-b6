package functions

import (
	"fmt"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
)

func materialise(context *api.Context, id b6.CollectionID, c api.Callable) (ingest.Change, error) {
	if c.NumArgs() != 0 {
		return nil, fmt.Errorf("Expected a function with no arguments, found %d", c.NumArgs())
	}

	result, _, err := c.CallWithArgs(context, []interface{}{}, nil)
	if err != nil {
		return nil, err
	}

	untyped, ok := result.(b6.UntypedCollection)
	if !ok {
		return nil, fmt.Errorf("Expected a collection, found %T", result)
	}
	collection := b6.AdaptCollection[any, any](untyped)

	keys, err := collection.AllKeys(nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to fill keys: %s", err)
	}

	values, err := collection.AllValues(nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to fill values: %s", err)
	}

	var expressions []*ingest.ExpressionFeature
	if expression, ok := c.Expression(); ok {
		expressions = append(expressions, &ingest.ExpressionFeature{
			ExpressionID: b6.ExpressionID{Namespace: id.Namespace, Value: id.Value},
			Expression:   expression,
		})
	}

	return &ingest.AddFeatures{
		Collections: []*ingest.CollectionFeature{
			{
				CollectionID: id,
				Keys:         keys,
				Values:       values,
			},
		},
		Expressions: expressions,
	}, nil
}
