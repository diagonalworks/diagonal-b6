package functions

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"diagonal.works/b6/api"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

func TestTake(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	values := make([]float64, 1000)
	for i := range values {
		values[i] = r.Float64()
	}
	keys := make([]interface{}, len(values))
	for i := range keys {
		keys[i] = fmt.Sprintf("%d", i)
	}

	n := 100
	collection := &api.ArrayAnyFloatCollection{Keys: keys, Values: values}
	took, err := take(&api.Context{}, collection, n)
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
		if filled[keys[i]] != values[i] {
			t.Errorf("Expected %f at index %d, found %f", values[i], i, filled[keys[i]])
		}
	}
}

func TestTopFloat(t *testing.T) {
	values := make([]float64, 1000)
	for i := range values {
		values[i] = float64(i)
	}
	r := rand.New(rand.NewSource(42))
	r.Shuffle(len(values), func(i int, j int) {
		values[i], values[j] = values[j], values[i]
	})

	keys := make([]interface{}, len(values))
	for i := range keys {
		keys[i] = fmt.Sprintf("%d", i)
	}

	n := 100
	collection := &api.ArrayAnyFloatCollection{Keys: keys, Values: values}
	selected, err := top(&api.Context{}, collection, n)
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
		expected := float64(len(values) - i - 1)
		if v != expected {
			t.Errorf("expected %f, found %f at index %d", expected, v, i)
		}
	}
}

func TestTopInt(t *testing.T) {
	values := make([]int, 1000)
	for i := range values {
		values[i] = i
	}
	r := rand.New(rand.NewSource(42))
	r.Shuffle(len(values), func(i int, j int) {
		values[i], values[j] = values[j], values[i]
	})

	keys := make([]interface{}, len(values))
	for i := range keys {
		keys[i] = fmt.Sprintf("%d", i)
	}

	n := 100
	collection := &api.ArrayAnyIntCollection{Keys: keys, Values: values}
	selected, err := top(&api.Context{}, collection, n)
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
		expected := len(values) - i - 1
		if v != expected {
			t.Errorf("expected %d, found %d at index %d", expected, v, i)
		}
	}
}

func TestTopWithMixedValuesGivesAnError(t *testing.T) {
	collection := api.ArrayAnyCollection{
		Keys:   []interface{}{"0", "1"},
		Values: []interface{}{0, 1.0},
	}

	_, err := top(&api.Context{}, &collection, 1)
	if err == nil {
		t.Errorf("Expected an error, found none")
	}
}

func TestFilter(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	values := make([]float64, 1000)
	for i := range values {
		values[i] = r.Float64()
	}
	keys := make([]interface{}, len(values))
	for i := range keys {
		keys[i] = fmt.Sprintf("%d", i)
	}

	limit := 0.5
	f := func(_ *api.Context, v interface{}) (bool, error) { return v.(float64) > limit, nil }
	collection := &api.ArrayAnyFloatCollection{Keys: keys, Values: values}
	filtered, err := filter(&api.Context{}, collection, f)
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
	collection := &api.ArrayAnyIntCollection{
		Keys:   []interface{}{"population:total", "population:children", "population:total"},
		Values: []int{100, 50, 200},
	}

	byKey, err := sumByKey(&api.Context{}, collection)
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
	collection := &api.ArrayAnyIntCollection{
		Keys:   []interface{}{"epc:habitablerooms", "epc:habitablerooms", "epc:habitablerooms"},
		Values: []int{2, 3, 2},
	}

	counted, err := countValues(&api.Context{}, collection)
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
	c1 := api.ArrayStringStringCollection{
		Keys:   []string{"ka", "kb", "kc"},
		Values: []string{"va", "vb", "vc"},
	}
	c2 := api.ArrayStringStringCollection{
		Keys:   []string{"kd", "ke", "kf"},
		Values: []string{"vd", "ve", "vf"},
	}
	c := api.ArrayAnyCollection{
		Keys:   []interface{}{0, 1},
		Values: []interface{}{&c1, &c2},
	}

	flattened, err := flatten(&api.Context{}, &c)
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

func TestHistogram(t *testing.T) {
	any := api.ArrayAnyCollection{
		Keys:   []interface{}{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16"},
		Values: []interface{}{"heath", "fold", "heath", "fold", "epping", "fold", "epping", "briki", "epping", "briki", "fold", "unfold", "heath", "fold", "epping", "home"},
	}

	buckets, err := histogram(&api.Context{}, &any)
	if err != nil {
		t.Errorf("Expected no error, found %q", err)
	}

	expected := api.ArrayAnyIntCollection{
		Keys:   []interface{}{"fold", "epping", "heath", "briki", "other"},
		Values: []int{5, 4, 3, 2, 2},
	}

	if diff := cmp.Diff(&expected, buckets, cmp.AllowUnexported(api.ArrayAnyIntCollection{})); diff != "" {
		t.Errorf("Expected no error, found %q", err)
	}

	numbers := api.ArrayAnyCollection{
		Keys:   []interface{}{"1", "2", "3", "4", "5", "6", "7"},
		Values: []interface{}{"1", "1", "1", "1", "1", "1", "2"},
	}

	buckets, err = histogram(&api.Context{}, &numbers)
	if err != nil {
		t.Errorf("Expected an error, found none")
	}

	expected = api.ArrayAnyIntCollection{
		Keys:   []interface{}{"1", "2"},
		Values: []int{6, 1},
	}

	if diff := cmp.Diff(&expected, buckets, cmp.AllowUnexported(api.ArrayAnyIntCollection{})); diff != "" {
		t.Errorf("Found diff (-want, +got):\n%s", diff)
	}
}

func TestExport(t *testing.T) {
	filename := t.TempDir() + "export.yml"

	c := api.ArrayAnyCollection{
		Keys:   []interface{}{"1st", "2nd"},
		Values: []interface{}{"makkoli", "soju"},
	}

	if err := export(&api.Context{}, &c, filename); err != nil {
		t.Errorf("expected no error, found %s", err)
	}

	expected := api.CollectionData{
		Keys:   []interface{}{"1st", "2nd"},
		Values: []interface{}{"makkoli", "soju"},
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("expected no error, found %s", err)
	}

	var r api.CollectionData
	err = yaml.Unmarshal(content, &r)
	if err != nil {
		t.Errorf("expected no error, found %s", err)
	}

	if diff := cmp.Diff(expected, r); diff != "" {
		t.Errorf("Found diff (-want, +got):\n%s", diff)
	}
}
