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

func NewJoinTagsFromFile(filename string, c context.Context) (JoinTags, error) {
	j := make(JoinTags)
	return j, j.FillFromFile(filename, c)
}

func NewJoinTagsFromPatterns(patterns []string, c context.Context) (JoinTags, error) {
	j := make(JoinTags)
	for _, pattern := range patterns {
		fs, err := filesystem.New(c, pattern)
		if err == nil {
			defer fs.Close()
			matches, err := fs.List(c, pattern)
			if err == nil {
				for _, match := range matches {
					if err = j.fillFromFile(match, fs, c); err != nil {
						break
					}
				}
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return j, nil
}

func (j JoinTags) FillFromFile(filename string, c context.Context) error {
	fs, err := filesystem.New(c, filename)
	if err != nil {
		return err
	}
	defer fs.Close()
	return j.fillFromFile(filename, fs, c)
}

func (j JoinTags) fillFromFile(filename string, fs filesystem.Interface, c context.Context) error {
	f, err := fs.OpenRead(c, filename)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	header, err := r.Read()
	if err != nil {
		return err
	}
	if len(header) < 2 {
		return fmt.Errorf("Expected at least 2 columns in %s, found %s", filename, header)
	}

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if len(row) == len(header) {
			tags := j[row[0]]
			for i := 1; i < len(row); i++ {
				tags = append(tags, b6.Tag{Key: header[i], Value: row[i]})
			}
			j[row[0]] = tags
		}
	}
	return nil
}

func (j JoinTags) AddTags(id string, f Feature) {
	if j != nil {
		for _, t := range j[id] {
			f.AddTag(t)
		}
	}
}
