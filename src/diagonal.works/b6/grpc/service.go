package grpc

import (
	"context"
	"fmt"
	"sync"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/api/functions"
	"diagonal.works/b6/ingest"
	pb "diagonal.works/b6/proto"
	"golang.org/x/mod/semver"
	"google.golang.org/protobuf/proto"
)

type service struct {
	pb.UnimplementedB6Server
	worlds  ingest.Worlds
	fs      api.FunctionSymbols
	a       api.Adaptors
	options api.Options
	lock    *sync.RWMutex
}

func (s *service) Evaluate(ctx context.Context, request *pb.EvaluateRequestProto) (*pb.EvaluateResponseProto, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	w := s.worlds.FindOrCreateWorld(b6.NewFeatureIDFromProto(request.Root))

	apply := func(change ingest.Change) (b6.Collection[b6.FeatureID, b6.FeatureID], error) {
		ids, err := change.Apply(w)
		return ids, err
	}

	if !semver.IsValid("v" + request.Version) {
		return nil, fmt.Errorf("client version %q is not a valid version", request.Version)
	} else if semver.Major("v"+request.Version) != semver.Major("v"+b6.ApiVersion) {
		return nil, fmt.Errorf("client version %s is not compatible with b6 version %s", request.Version, b6.ApiVersion)
	}

	context := api.Context{
		World:           w,
		Worlds:          s.worlds,
		FunctionSymbols: s.fs,
		Adaptors:        s.a,
		Context:         ctx,
	}
	context.FillFromOptions(&s.options)
	expression, err := b6.ExpressionFromProto(request.Request)
	if err != nil {
		return nil, err
	}
	simplified := api.Simplify(expression, context.FunctionSymbols)
	v, err := api.Evaluate(simplified, &context)
	if err != nil {
		return nil, err
	}

	if change, ok := v.(ingest.Change); ok {
		s.lock.RUnlock()
		s.lock.Lock()
		v, err = apply(change)
		s.lock.Unlock()
		s.lock.RLock()
		if err != nil {
			return nil, err
		}
	}
	ve, err := b6.FromLiteral(v)
	if err != nil {
		return nil, err
	}

	pe, err := ve.ToProto()
	if err != nil {
		return nil, err
	}

	r := &pb.EvaluateResponseProto{
		Result: pe,
	}
	if _, err := proto.Marshal(r); err != nil {
		panic(err)
	}
	return &pb.EvaluateResponseProto{
		Result: pe,
	}, nil
}

func (s *service) ListWorlds(ctx context.Context, request *pb.ListWorldsRequestProto) (*pb.ListWorldsResponseProto, error) {
	ids := s.worlds.ListWorlds()
	response := &pb.ListWorldsResponseProto{
		Ids: make([]*pb.FeatureIDProto, len(ids)),
	}
	for i, id := range ids {
		response.Ids[i] = b6.NewProtoFromFeatureID(id)
	}
	return response, nil
}

func (s *service) DeleteWorld(ctx context.Context, request *pb.DeleteWorldRequestProto) (*pb.DeleteWorldResponseProto, error) {
	s.worlds.DeleteWorld(b6.NewFeatureIDFromProto(request.Id))
	return &pb.DeleteWorldResponseProto{}, nil
}

func NewB6Service(worlds ingest.Worlds, options api.Options, lock *sync.RWMutex) pb.B6Server {
	return &service{
		worlds:  worlds,
		fs:      functions.Functions(),
		a:       functions.Adaptors(),
		options: options,
		lock:    lock,
	}
}
