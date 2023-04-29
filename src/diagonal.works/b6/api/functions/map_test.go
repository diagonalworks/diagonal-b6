package functions

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"diagonal.works/b6/api"
)

func TestMapWithDeadline(t *testing.T) {
	input := &api.ArrayIntIntCollection{Values: make([]int, 1031)}
	r := rand.New(rand.NewSource(42))
	max := 100000
	for i := range input.Values {
		input.Values[i] = r.Intn(max)
	}

	f := func(v interface{}, c *api.Context) (interface{}, error) {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		return v.(int) + 1, nil
	}

	seen := 0
	deadline, _ := context.WithTimeout(context.Background(), 2000*time.Microsecond)
	c, err := map_(input, f, &api.Context{Context: deadline})
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
	} else {
		i := c.Begin()
		for {
			var ok bool
			ok, err = i.Next()
			if !ok || err != nil {
				break
			}
			if i.Value().(int) != input.Values[seen]+1 {
				t.Errorf("Expected %d, found %d at position %d", input.Values[seen]+1, i.Value().(int), seen)
				return
			}
			seen++
		}
	}
	if err != context.DeadlineExceeded {
		t.Errorf("Expected %s, found %s", context.DeadlineExceeded, err)
	}
}

func TestMapParallelHappyPath(t *testing.T) {
	// Use a prime input length, to guarnatee it's not divisible by the
	// number of cores
	input := &api.ArrayIntIntCollection{Values: make([]int, 1031)}
	r := rand.New(rand.NewSource(42))
	for i := range input.Values {
		input.Values[i] = r.Intn(100000)
	}

	f := func(v interface{}, c *api.Context) (interface{}, error) {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		return v.(int) + 1, nil
	}

	seen := 0
	context := &api.Context{Cores: 8, Context: context.Background(), VM: &api.VM{}}
	c, err := mapParallel(input, f, context)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
	} else {
		i := c.Begin()
		for {
			ok, err := i.Next()
			if err != nil {
				t.Errorf("Expected no error, found: %s", err)
				break
			} else if !ok {
				break
			}
			if i.Value().(int) != input.Values[seen]+1 {
				t.Errorf("Expected %d, found %d at position %d", input.Values[seen]+1, i.Value().(int), seen)
				break
			}
			seen++
		}
	}
	if seen != len(input.Values) {
		t.Errorf("Expected %d values, found %d", len(input.Values), seen)
	}
}

func TestMapParallelWithFunctionReturningError(t *testing.T) {
	input := &api.ArrayIntIntCollection{Values: make([]int, 1031)}
	r := rand.New(rand.NewSource(42))
	max := 100000
	for i := range input.Values {
		input.Values[i] = r.Intn(max)
	}
	input.Values[479] = max // Choose to fail at an arbitrary point

	broken := errors.New("Broken")
	f := func(v interface{}, c *api.Context) (interface{}, error) {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		if v.(int) == max {
			return 0, broken
		}
		return v.(int) + 1, nil
	}

	seen := 0
	context := &api.Context{Cores: 8, Context: context.Background(), VM: &api.VM{}}
	c, err := mapParallel(input, f, context)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
	} else {
		i := c.Begin()
		for {
			var ok bool
			ok, err = i.Next()
			if !ok || err != nil {
				break
			}
			if i.Value().(int) != input.Values[seen]+1 {
				t.Errorf("Expected %d, found %d at position %d", input.Values[seen]+1, i.Value().(int), seen)
				return
			}
			seen++
		}
	}
	if err != broken {
		t.Errorf("Expected error %s, found %s", broken, err)
	}
}

type brokenCollection struct {
	c     api.Collection
	err   error
	after int

	i api.CollectionIterator
}

func (b *brokenCollection) Begin() api.CollectionIterator {
	return &brokenCollection{c: b.c, err: b.err, after: b.after, i: b.c.Begin()}
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

func TestMapParallelWithIteratorReturningError(t *testing.T) {
	values := &api.ArrayIntIntCollection{Values: make([]int, 1031)}
	r := rand.New(rand.NewSource(42))
	for i := range values.Values {
		values.Values[i] = r.Intn(100000)
	}

	f := func(v interface{}, c *api.Context) (interface{}, error) {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		return v.(int) + 1, nil
	}

	broken := errors.New("Broken")
	// Choose to fail at an arbitrary point
	input := &brokenCollection{c: values, after: 479, err: broken}

	seen := 0
	context := &api.Context{Cores: 8, Context: context.Background(), VM: &api.VM{}}
	c, err := mapParallel(input, f, context)
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
	} else {
		i := c.Begin()
		for {
			var ok bool
			ok, err = i.Next()
			if !ok || err != nil {
				break
			}
			if i.Value().(int) != values.Values[seen]+1 {
				t.Errorf("Expected %d, found %d at position %d", values.Values[seen]+1, i.Value().(int), seen)
				return
			}
			seen++
		}
	}
	if err != broken {
		t.Errorf("Expected error %s, found %s", broken, err)
	}
}

func TestMapParallelWithDeadline(t *testing.T) {
	input := &api.ArrayIntIntCollection{Values: make([]int, 1031)}
	r := rand.New(rand.NewSource(42))
	max := 100000
	for i := range input.Values {
		input.Values[i] = r.Intn(max)
	}

	f := func(v interface{}, c *api.Context) (interface{}, error) {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		return v.(int) + 1, nil
	}

	seen := 0
	cores := 8
	deadline, _ := context.WithTimeout(context.Background(), 200*time.Microsecond)
	c, err := mapParallel(input, f, &api.Context{Cores: cores, Context: deadline, VM: &api.VM{}})
	if err != nil {
		t.Errorf("Expected no error, found: %s", err)
	} else {
		i := c.Begin()
		for {
			var ok bool
			ok, err = i.Next()
			if !ok || err != nil {
				break
			}
			if i.Value().(int) != input.Values[seen]+1 {
				t.Errorf("Expected %d, found %d at position %d", input.Values[seen]+1, i.Value().(int), seen)
				return
			}
			seen++
		}
	}
	if err != context.DeadlineExceeded {
		t.Errorf("Expected %s, found %s", context.DeadlineExceeded, err)
	}
}
