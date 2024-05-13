package api

import (
	"context"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/ingest"
	pb "diagonal.works/b6/proto"
)

type AppliedChange struct {
	Change   ingest.Change
	Modified b6.Collection[b6.FeatureID, b6.FeatureID]
}

type Evaluator struct {
	Worlds          ingest.Worlds
	FunctionSymbols FunctionSymbols
	Adaptors        Adaptors
	Options         Options
	Lock            *sync.RWMutex
}

func (e *Evaluator) EvaluateProto(request *pb.EvaluateRequestProto) (interface{}, error) {
	var expression b6.Expression
	if err := expression.FromProto(request.Request); err != nil {
		return nil, err
	}
	root := b6.NewFeatureIDFromProto(request.Root)
	return e.EvaluateExpression(expression, root)
}

func (e *Evaluator) EvaluateString(expression string, root b6.FeatureID) (interface{}, error) {
	parsed, err := ParseExpression(expression)
	if err != nil {
		return nil, err
	}
	v, err := e.EvaluateExpression(parsed, root)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (e *Evaluator) EvaluateExpression(expression b6.Expression, root b6.FeatureID) (interface{}, error) {
	var world ingest.MutableWorld
	if root.IsValid() {
		world = e.Worlds.FindOrCreateWorld(root)
	} else {
		world = e.Worlds.FindOrCreateWorld(ingest.DefaultWorldFeatureID)
	}
	vmContext := Context{
		World:           world,
		Worlds:          e.Worlds,
		FunctionSymbols: e.FunctionSymbols,
		Adaptors:        e.Adaptors,
		Context:         context.Background(),
	}
	vmContext.FillFromOptions(&e.Options)

	simplified := Simplify(expression, e.FunctionSymbols)
	v, err := Evaluate(simplified, &vmContext)
	if err != nil {
		return b6.Literal{AnyLiteral: nil}, err
	}

	if change, ok := v.(ingest.Change); ok {
		e.Lock.RUnlock()
		e.Lock.Lock()
		var modified b6.Collection[b6.FeatureID, b6.FeatureID]
		modified, err = change.Apply(world)
		e.Lock.Unlock()
		e.Lock.RLock()
		return &AppliedChange{Change: change, Modified: modified}, nil
	}
	return v, err
}
