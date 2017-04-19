package pusher

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

func createMapping() mapping.IndexMapping {
	mapping := bleve.NewIndexMapping()
	return mapping
}

func openIndex(path string) (index bleve.Index, err error) {
	if index, err = bleve.Open(path); err != nil {
		mapping := createMapping()
		if index, err = bleve.New(path, mapping); err != nil {
			return
		}
	}
	return
}
