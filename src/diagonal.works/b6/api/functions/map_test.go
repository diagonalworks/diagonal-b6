package functions

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
)

func TestMapWithDeadline(t *testing.T) {
	input := b6.ArrayValuesCollection[int](make([]int, 1031))
	r := rand.New(rand.NewSource(42))
	max := 100000
	for i := range input {
		input[i] = r.Intn(max)
	}

	f := func(c *api.Context, v interface{}) (interface{}, error) {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		return v.(int) + 1, nil
	}

	seen := 0
	deadline, _ := context.WithTimeout(context.Background(), 2000*time.Microsecond)
	c, err := map_(&api.Context{Context: deadline}, input.Collection(), f)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	i := c.Begin()
	for {
		var ok bool
		ok, err = i.Next()
		if !ok || err != nil {
			break
		}
		if i.Value().(int) != input[seen]+1 {
			t.Fatalf("Expected %d, found %d at position %d", input[seen]+1, i.Value().(int), seen)
		}
		seen++
	}
	if err != context.DeadlineExceeded {
		t.Errorf("Expected %s, found %s", context.DeadlineExceeded, err)
	}
}

func TestMapParallelHappyPath(t *testing.T) {
	// Use a prime input length, to guarantee it's not divisible by the
	// number of cores
	input := b6.ArrayValuesCollection[int](make([]int, 1031))
	r := rand.New(rand.NewSource(42))
	for i := range input {
		input[i] = r.Intn(100000)
	}

	f := func(c *api.Context, v interface{}) (interface{}, error) {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		return v.(int) + 1, nil
	}

	seen := 0
	context := &api.Context{Cores: 8, Context: context.Background(), VM: &api.VM{}}
	c, err := mapParallel(context, input.Collection(), f)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	i := c.Begin()
	for {
		ok, err := i.Next()
		if err != nil {
			t.Fatalf("Expected no error, found: %s", err)
		}
		if !ok {
			break
		}
		if i.Value().(int) != input[seen]+1 {
			t.Fatalf("Expected %d, found %d at position %d", input[seen]+1, i.Value().(int), seen)
		}
		seen++
	}
	if seen != len(input) {
		t.Errorf("Expected %d values, found %d", len(input), seen)
	}
}

func TestMapParallelWithFunctionReturningError(t *testing.T) {
	input := b6.ArrayValuesCollection[int](make([]int, 1031))
	r := rand.New(rand.NewSource(42))
	max := 100000
	for i := range input {
		input[i] = r.Intn(max)
	}
	input[479] = max // Choose to fail at an arbitrary point

	broken := errors.New("Broken")
	f := func(c *api.Context, v interface{}) (interface{}, error) {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		if v.(int) == max {
			return 0, broken
		}
		return v.(int) + 1, nil
	}

	seen := 0
	context := &api.Context{Cores: 8, Context: context.Background(), VM: &api.VM{}}
	c, err := mapParallel(context, input.Collection(), f)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	i := c.Begin()
	for {
		var ok bool
		ok, err = i.Next()
		if !ok || err != nil {
			break
		}
		if i.Value().(int) != input[seen]+1 {
			t.Fatalf("Expected %d, found %d at position %d", input[seen]+1, i.Value().(int), seen)
		}
		seen++
	}
	if err != broken {
		t.Errorf("Expected error %s, found %s", broken, err)
	}
}

type brokenCollection struct {
	c     b6.UntypedCollection
	err   error
	after int

	i b6.Iterator[any, any]
}

func (b *brokenCollection) Begin() b6.Iterator[any, any] {
	return &brokenCollection{c: b.c, err: b.err, after: b.after, i: b.c.BeginUntyped()}
}

func (b *brokenCollection) Next() (bool, error) {
	if b.after <= 0 {
		return false, b.err
	}
	b.after--
	return b.i.Next()
}

func (b *brokenCollection) Key() interface{} {
	return b.i.Key()
}

func (b *brokenCollection) Value() interface{} {
	return b.i.Value()
}

func (b *brokenCollection) Count() (int, bool) {
	return b.c.Count()
}

func TestMapParallelWithIteratorReturningError(t *testing.T) {
	values := b6.ArrayValuesCollection[int](make([]int, 1031))
	r := rand.New(rand.NewSource(42))
	for i := range values {
		values[i] = r.Intn(100000)
	}

	f := func(c *api.Context, v interface{}) (interface{}, error) {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		return v.(int) + 1, nil
	}

	broken := errors.New("Broken")
	// Choose to fail at an arbitrary point
	input := b6.Collection[any, any]{
		AnyCollection: &brokenCollection{c: values.Collection(), after: 479, err: broken},
	}

	seen := 0
	context := &api.Context{Cores: 8, Context: context.Background(), VM: &api.VM{}}
	mapped, err := mapParallel(context, input, f)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	i := mapped.Begin()
	for {
		var ok bool
		ok, err = i.Next()
		if !ok || err != nil {
			break
		}
		if i.Value().(int) != values[seen]+1 {
			t.Fatalf("Expected %d, found %d at position %d", values[seen]+1, i.Value().(int), seen)
		}
		seen++
	}
	if err != broken {
		t.Errorf("Expected error %s, found %s", broken, err)
	}
}

func TestMapParallelWithDeadline(t *testing.T) {
	input := b6.ArrayValuesCollection[int](make([]int, 1031))
	r := rand.New(rand.NewSource(42))
	max := 100000
	for i := range input {
		input[i] = r.Intn(max)
	}

	f := func(c *api.Context, v interface{}) (interface{}, error) {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		return v.(int) + 1, nil
	}

	seen := 0
	cores := 8
	deadline, _ := context.WithTimeout(context.Background(), 200*time.Microsecond)
	ctx := &api.Context{Cores: cores, Context: deadline, VM: &api.VM{}}
	mapped, err := mapParallel(ctx, input.Collection(), f)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}
	i := mapped.Begin()
	for {
		var ok bool
		ok, err = i.Next()
		if !ok || err != nil {
			break
		}
		if i.Value().(int) != input[seen]+1 {
			t.Fatalf("Expected %d, found %d at position %d", input[seen]+1, i.Value().(int), seen)
		}
		seen++
	}
	if err != context.DeadlineExceeded {
		t.Errorf("Expected %s, found %s", context.DeadlineExceeded, err)
	}
}
