package gtfs

import (
	"fmt"
	"strings"
	"testing"

	"diagonal.works/b6"
	"diagonal.works/b6/graph"
	"diagonal.works/b6/ingest"
	"diagonal.works/b6/test"
)

const cores = 2
const operator = ""

func TestGTFSWorldMissingData(t *testing.T) {
	_, err := newWorldFromGTFSFiles(test.Data("gtfs-gala/"), operator, cores)
	if !strings.Contains(err.Error(), "routes.txt: no such file or directory") {
		t.Errorf("Expected missing files error, got %s", err.Error())
	}
}

func TestGTFSWorldPoint(t *testing.T) {
	w, err := newWorldFromGTFSFiles(test.Data("gtfs-manchester/"), operator, cores)
	if err != nil {
		t.Errorf("Failed to build world: %s", err.Error())
	}

	// 1800NB04091: "Manchester City Centre, Parsonage (Stop NC)" @ 53.48348,-2.24705
	point := w.FindFeatureByID(b6.FeatureID{Type: b6.FeatureTypePoint, Namespace: b6.NamespaceGTFS, Value: 18183387289522862582})
	location, _ := w.FindLocationByID(point.FeatureID())
	if location.String() != "[53.4834800, -2.2470500]" {
		t.Error("Expected location [53.4834800, -2.2470500], got : " + location.String())
	}
	tags := b6.Tags(point.AllTags())
	if len(tags) != 2 {
		t.Errorf(fmt.Sprintf("Expected 2 tags got %d", len(tags)))
	}

	if tags.Get("#gtfs").Value.String() != "stop" {
		t.Errorf("Expected #gtfs tag value stop got %s %s", tags[0].Key, tags[0].Value)
	}
}

func TestGTFSWorldPath(t *testing.T) {
	w, err := newWorldFromGTFSFiles(test.Data("gtfs-manchester/"), operator, cores)
	if err != nil {
		t.Errorf("Failed to build world: %s", err.Error())
	}

	// 1800NB04431: "Manchester City Centre, Victoria Stn Approach (Stop Ny)" @ 53.48716,-2.24285
	// 1800NB04091: "Manchester City Centre, Parsonage (Stop NC)" @ 53.48348,-2.24705
	path := w.FindFeatureByID(b6.FeatureID{Type: b6.FeatureTypePath, Namespace: b6.NamespaceGTFS, Value: 7371855906284259301})
	tags := path.AllTags()
	if len(tags) != 4 {
		t.Errorf(fmt.Sprintf("Expected only 4 tags got %d", len(tags)))
	}

	if tags[2].Key != graph.GTFSPeakTimeTag || tags[2].Value.String() != "180" {
		t.Errorf("Expected gtfs:peak tag value 180 got %s %s", tags[0].Key, tags[0].Value)
	}

	if tags[3].Key != graph.GTFSOffPeakTimeTag || tags[3].Value.String() != "120" {
		t.Errorf("Expected gtfs:off-peak tag value 120 got %s %s", tags[1].Key, tags[1].Value)
	}
}

func newWorldFromGTFSFiles(dir string, operator string, cores int) (b6.World, error) {
	source := TXTFilesGTFSSource{Directory: dir, Operator: operator, FailWhenNoFiles: true}
	return ingest.NewWorldFromSource(&source, &ingest.BuildOptions{Cores: cores})
}
