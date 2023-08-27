package grpc

import (
	"context"
	"sync"
	"testing"
	"time"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"diagonal.works/b6/ingest"
	pb "diagonal.works/b6/proto"
	"diagonal.works/b6/test/camden"
)

func findInTagsProto(tags []*pb.TagProto, key string) (string, bool) {
	for _, tag := range tags {
		if tag.Key == key {
			return tag.Value, true
		}
	}
	return "", false
}

func TestGRPC(t *testing.T) {
	tests := []struct {
		name string
		f    func(pb.B6Server, b6.World, *testing.T)
	}{
		{"Evaluate", ValidateEvaluate},
		{"ConcurrentReadAndWrite", ValidateConcurrentReadAndWrite},
		{"RejectRequestsWithDifferentMajorVersion", ValidateRejectRequestsWithDifferentMajorVersion},
	}

	base := camden.BuildGranarySquareForTests(t)

	w := ingest.NewMutableOverlayWorld(base)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.f(NewB6Service(w, 1, &sync.RWMutex{}), w, t)
		})
	}
}

func ValidateEvaluate(service pb.B6Server, w b6.World, t *testing.T) {
	e := `find [#building] | map {b -> get b "building:levels"}`
	root, err := api.ParseExpression(e)
	if err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}
	request := &pb.EvaluateRequestProto{
		Request: root,
		Version: b6.ApiVersion,
	}
	response, err := service.Evaluate(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	node := response.GetResult()
	if node == nil {
		t.Fatal("Expected a node")
	}
	literal := node.GetLiteral()
	if literal == nil {
		t.Fatal("Expected a literal")
	}
	collection := literal.GetCollectionValue()
	if collection == nil {
		t.Fatal("Expected a CollectionValue")
	}
	expected := camden.BuildingsInGranarySquare
	if len(collection.Values) != expected {
		t.Errorf("Expected %d values, found %d", expected, len(collection.Values))
	}
}

func ValidateConcurrentReadAndWrite(service pb.B6Server, w b6.World, t *testing.T) {
	// Read from, and write to, the world in two different goroutines. Although in theory
	// the test in non-deterministic, as the reads and writes may be accidentally entirely
	// out of sync, in practice the loops are tight enough and the time delay long enough
	// that this isn't an issue, and artificially slowing an operation would be invasive.
	end := time.Now().Add(1 * time.Second)
	var wg sync.WaitGroup
	wg.Add(2)

	e := `find [#building] | map {b -> tag "diagonal-fill-colour" (get b "building:levels")}`
	write, err := api.ParseExpression(e)
	if err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

	e = `find [#building] | map {b -> get b "building:levels"}`
	read, err := api.ParseExpression(e)
	if err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

	// Writer
	go func() {
		defer wg.Done()
		for time.Now().Before(end) {
			request := &pb.EvaluateRequestProto{
				Request: write,
				Version: b6.ApiVersion,
			}
			if _, err := service.Evaluate(context.Background(), request); err != nil {
				t.Error(err)
				return
			}
		}
	}()

	// Reader
	go func() {
		defer wg.Done()
		for time.Now().Before(end) {
			request := &pb.EvaluateRequestProto{
				Request: read,
				Version: b6.ApiVersion,
			}
			if _, err := service.Evaluate(context.Background(), request); err != nil {
				t.Error(err)
				return
			}
		}
	}()
	wg.Wait()
}

func ValidateRejectRequestsWithDifferentMajorVersion(service pb.B6Server, w b6.World, t *testing.T) {
	request := &pb.EvaluateRequestProto{
		Request: &pb.NodeProto{
			Node: &pb.NodeProto_Literal{
				Literal: &pb.LiteralNodeProto{
					Value: &pb.LiteralNodeProto_IntValue{
						IntValue: 42,
					},
				},
			},
		},
	}
	if _, err := service.Evaluate(context.Background(), request); err == nil {
		t.Error("Expected error when missing version, found none")
	}

	request.Version = "36.0.0" // Will need to change when ApiVersion passes 36
	if _, err := service.Evaluate(context.Background(), request); err == nil {
		t.Error("Expected error with different major version, found none")
	}
}
