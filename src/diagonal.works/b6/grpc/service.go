package grpc

import (
	"context"
	"sync"

	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
	"diagonal.works/b6/ingest"
	pb "diagonal.works/b6/proto"
	"google.golang.org/protobuf/proto"
)

type service struct {
	pb.UnimplementedB6Server
	world ingest.MutableWorld
	fs    api.FunctionSymbols
	fw    api.FunctionWrappers
	cores int
	lock  *sync.RWMutex
}

func (s *service) Evaluate(ctx context.Context, request *pb.EvaluateRequestProto) (*pb.EvaluateResponseProto, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	apply := func(change ingest.Change) (ingest.AppliedChange, error) {
		s.lock.RUnlock()
		s.lock.Lock()
		ids, err := change.Apply(s.world)
		s.lock.Unlock()
		s.lock.RLock()
		return ids, err
	}
	context := api.Context{
		World:            s.world,
		FunctionSymbols:  s.fs,
		FunctionWrappers: s.fw,
		Cores:            s.cores,
		Context:          ctx,
	}
	if v, err := api.Evaluate(request.Request, &context); err == nil {
		if change, ok := v.(ingest.Change); ok {
			v, err = apply(change)
			if err != nil {
				return nil, err
			}
		}
		if p, err := api.ToProto(v); err == nil {
			r := &pb.EvaluateResponseProto{
				Result: p,
			}
			if _, err := proto.Marshal(r); err != nil {
				panic(err)
			}
			return &pb.EvaluateResponseProto{
				Result: p,
			}, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func NewB6Service(w ingest.MutableWorld, cores int, lock *sync.RWMutex) pb.B6Server {
	return &service{
		world: w,
		fs:    functions.Functions(),
		fw:    functions.Wrappers(),
		cores: cores,
		lock:  lock,
	}
}
