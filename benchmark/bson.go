package benchmark

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	perfDataDir  = "perf"
	bsonDataDir  = "extended_bson"
	flatBSONData = "flat_bson.json"
	deepBSONData = "deep_bson.json"
	fullBSONData = "full_bson.json"
)

// utility functions for the bson benchmarks

func loadSourceDocument(pathParts ...string) (*bson.Document, error) {
	data, err := ioutil.ReadFile(filepath.Join(pathParts...))
	if err != nil {
		return nil, err
	}
	doc := bson.NewDocument()
	err = bson.UnmarshalExtJSON(data, true, &doc)
	if err != nil {
		return nil, err
	}

	if doc.Len() == 0 {
		return nil, errors.New("empty bson document")
	}

	return doc, nil
}

func loadSourceRaw(pathParts ...string) (bson.Raw, error) {
	doc, err := loadSourceDocument(pathParts...)
	if err != nil {
		return nil, err
	}
	raw, err := doc.MarshalBSON()
	if err != nil {
		return nil, err
	}

	return bson.Raw(raw), nil
}
