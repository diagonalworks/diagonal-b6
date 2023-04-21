package ingest

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"diagonal.works/b6"
	"github.com/apache/beam/sdks/go/pkg/beam/io/filesystem"
)

type JoinTags map[string][]b6.Tag

func NewJoinTagsFromCSV(filename string) (JoinTags, error) {
	fs, err := filesystem.New(context.Background(), filename)
	if err != nil {
		return nil, err
	}
	defer fs.Close()

	f, err := fs.OpenRead(context.Background(), filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	if len(header) < 2 {
		return nil, fmt.Errorf("Expected at least 2 columns in %s, found %s", filename, header)
	}

	join := make(JoinTags)
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if len(row) == len(header) {
			tags := join[row[0]]
			for i := 1; i < len(row); i++ {
				tags = append(tags, b6.Tag{Key: header[i], Value: row[i]})
			}
			join[row[0]] = tags
		}
	}
	return join, nil
}

func (j JoinTags) AddTags(id string, f Feature) {
	if j != nil {
		for _, t := range j[id] {
			f.AddTag(t)
		}
	}
}
