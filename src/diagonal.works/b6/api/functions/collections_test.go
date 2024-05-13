package functions

import (
	"context"
	"math/rand"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/api"
	"github.com/google/go-cmp/cmp"
)

func TestTake(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	collection := b6.ArrayValuesCollection[float64]{}
	for i := 0; i < 1000; i++ {
		collection = append(collection, r.Float64())

	}
	n := 100
	took, err := take(&api.Context{}, collection.Collection(), n)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	filled := make(map[interface{}]float64)
	if err := api.FillMap(took, filled); err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

	if n != len(filled) {
		t.Errorf("Expected %d values, found %d", n, len(filled))
	}

	for i := 0; i < n; i++ {
		if filled[i] != collection[i] {
			t.Errorf("Expected %f at index %d, found %f", collection[i], i, filled[i])
		}
	}
}

func TestTopFloat(t *testing.T) {
	collection := b6.ArrayValuesCollection[float64]{}
	for i := 0; i < 1000; i++ {
		collection = append(collection, float64(i))
	}
	r := rand.New(rand.NewSource(42))
	r.Shuffle(len(collection), func(i int, j int) {
		collection[i], collection[j] = collection[j], collection[i]
	})

	n := 100
	selected, err := top(&api.Context{}, collection.Collection(), n)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	filled := make([]float64, 0, n)
	if err := api.FillSliceFromValues(selected, &filled); err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

	if n != len(filled) {
		t.Errorf("Expected %d values, found %d", n, len(filled))
	}

	for i, v := range filled {
		expected := float64(len(collection) - i - 1)
		if v != expected {
			t.Errorf("expected %f, found %f at index %d", expected, v, i)
		}
	}
}

func TestTopInt(t *testing.T) {
	collection := b6.ArrayValuesCollection[int]{}
	for i := 0; i < 1000; i++ {
		collection = append(collection, i)
	}
	r := rand.New(rand.NewSource(42))
	r.Shuffle(len(collection), func(i int, j int) {
		collection[i], collection[j] = collection[j], collection[i]
	})

	n := 100
	selected, err := top(&api.Context{}, collection.Collection(), n)
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	filled := make([]int, 0, n)
	if err := api.FillSliceFromValues(selected, &filled); err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

	if n != len(filled) {
		t.Errorf("Expected %d values, found %d", n, len(filled))
	}

	for i, v := range filled {
		expected := len(collection) - i - 1
		if v != expected {
			t.Errorf("expected %d, found %d at index %d", expected, v, i)
		}
	}
}

func TestTopWithMixedValuesGivesAnError(t *testing.T) {
	collection := &b6.ArrayCollection[interface{}, interface{}]{
		Keys:   []interface{}{"0", "1"},
		Values: []interface{}{0, 1.0},
	}

	_, err := top(&api.Context{}, collection.Collection(), 1)
	if err == nil {
		t.Errorf("Expected an error, found none")
	}
}

func TestFilter(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	collection := b6.ArrayValuesCollection[float64]{}
	for i := 0; i < 1000; i++ {
		collection = append(collection, r.Float64())

	}

	limit := 0.5
	f := func(_ *api.Context, v interface{}) (bool, error) { return v.(float64) > limit, nil }
	c := &api.Context{Cores: 8, Context: context.Background(), VM: &api.VM{}}
	filtered, err := filter(c, collection.Collection(), api.NewNativeFunction1(f, b6.NewSymbolExpression("x-native")))
	if err != nil {
		t.Fatalf("Expected no error, found: %s", err)
	}

	filled := make(map[interface{}]float64)
	if err := api.FillMap(filtered, filled); err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}

	if len(filled) == 0 {
		t.Fatalf("Expected at least 1 value")
	} else {
		for _, f := range filled {
			if f <= limit {
				t.Errorf("Expected %f to be below %f", f, limit)
			}
		}
	}
}

func TestSumByKey(t *testing.T) {
	collection := b6.ArrayCollection[string, int]{
		Keys:   []string{"population:total", "population:children", "population:total"},
		Values: []int{100, 50, 200},
	}

	byKey, err := sumByKey(&api.Context{}, collection.Collection().Values())
	if err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}
	filled := make(map[string]int)
	if err := api.FillMap(byKey, filled); err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}
	expected := map[string]int{
		"population:total":    300,
		"population:children": 50,
	}
	if diff := cmp.Diff(expected, filled); diff != "" {
		t.Errorf("Found diff (-want, +got):\n%s", diff)
	}
}

func TestCountValues(t *testing.T) {
	collection := b6.ArrayCollection[string, int]{
		Keys:   []string{"epc:habitablerooms", "epc:habitablerooms", "epc:habitablerooms"},
		Values: []int{2, 3, 2},
	}

	c := b6.AdaptCollection[any, any](collection.Collection())
	counted, err := countValues(&api.Context{}, c)
	if err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}
	filled := make(map[int]int)
	if err := api.FillMap(counted, filled); err != nil {
		t.Fatalf("Expected no error, found %s", err)
	}
	expected := map[int]int{
		2: 2,
		3: 1,
	}
	if diff := cmp.Diff(expected, filled); diff != "" {
		t.Errorf("Found diff (-want, +got):\n%s", diff)
	}
}

func TestFlatten(t *testing.T) {
	c1 := b6.ArrayCollection[string, string]{
		Keys:   []string{"ka", "kb", "kc"},
		Values: []string{"va", "vb", "vc"},
	}
	c2 := b6.ArrayCollection[string, string]{
		Keys:   []string{"kd", "ke", "kf"},
		Values: []string{"vd", "ve", "vf"},
	}
	c := b6.ArrayCollection[any, b6.UntypedCollection]{
		Keys:   []any{0, 1},
		Values: []b6.UntypedCollection{c1.Collection(), c2.Collection()},
	}

	flattened, err := flatten(&api.Context{}, c.Collection())
	if err != nil {
		t.Fatalf("Expected no error, found %q", err)
	}

	filled := make(map[string]string)
	if err := api.FillMap(flattened, filled); err != nil {
		t.Fatalf("Expected no error, found %q", err)
	}

	expected := map[string]string{
		"ka": "va",
		"kb": "vb",
		"kc": "vc",
		"kd": "vd",
		"ke": "ve",
		"kf": "vf",
	}
	if diff := cmp.Diff(expected, filled); diff != "" {
		t.Errorf("Found diff (-want, +got):\n%s", diff)
	}
}
