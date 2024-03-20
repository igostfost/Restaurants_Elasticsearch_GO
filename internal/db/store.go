package db

import (
	"elasticTask/pkg/types"
	"github.com/elastic/go-elasticsearch/v8"

)

type Store interface {
	// returns a list of items, a total number of hits and (or) an error in case of one
	GetPlaces(limit int, offset int) ([]types.Place, int, error)
}

type ElasticsearchStore struct {
	client    *elasticsearch.Client
	indexName string
}

func NewElasticsearchStore(indexName string) (*ElasticsearchStore, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &ElasticsearchStore{
		client:    es,
		indexName: indexName,
	}, nil
}

