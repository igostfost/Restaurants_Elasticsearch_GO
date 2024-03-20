package db

import (
	"bytes"
	"context"
	"elasticTask/internal/csvreader"
	"elasticTask/pkg/types"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

func (es ElasticsearchStore) Indexeres(indexName string) {
	log.SetFlags(0)

	var (
		numWorkers      = runtime.NumCPU()
		flushBytes      = 5e+6
		data            []*types.Place
		countSuccessful uint64
		res             *esapi.Response
		err             error
	)

	log.Println(strings.Repeat("▁", 65))
	data = csvreader.CsvReader()

	// check index exsist
	existsReq := esapi.IndicesExistsRequest{
		Index: []string{indexName},
	}

	existsRes, err := existsReq.Do(context.Background(), es.client)
	if err != nil {
		log.Fatalf("Error checking index existence: %s", err)
	}
	defer existsRes.Body.Close()

	if existsRes.StatusCode == 200 {
		deleteReq := esapi.IndicesDeleteRequest{
			Index: []string{indexName},
		}

		deleteRes, err := deleteReq.Do(context.Background(), es.client)
		if err != nil {
			log.Fatalf("Error deleting index: %s", err)
		}
		defer deleteRes.Body.Close()

		if deleteRes.IsError() {
			log.Fatalf("Error deleting index: %s", deleteRes)
		}

		log.Printf("Index '%s' deleted successfully", indexName)
	}

	// Creating Index and Starting Mapping
	fmt.Println("Creating Index and Starting Mapping...")
	mapping := `{
    "properties": {
        "address": {"type": "text"},
        "phone": {"type": "text"},
        "name": {"type": "text"},
        "location": {"type": "geo_point"},
        "id": {"type": "long"}
    }
}`

	settings := `{
    "number_of_shards": 1,
    "number_of_replicas": 1,
    "max_result_window": 20000
}`

	req := esapi.IndicesCreateRequest{
		Index: indexName,
		Body:  strings.NewReader(fmt.Sprintf(`{"settings": %s}`, settings)),
	}

	res, err = req.Do(context.Background(), es.client)
	if err != nil {
		log.Fatalf("Error creating index: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error creating index: %s", res)
	}

	log.Printf("Index '%s' created successfully", indexName)

	reqMapping := esapi.IndicesPutMappingRequest{
		Index: []string{indexName},
		Body:  strings.NewReader(mapping),
	}

	resMapping, err := reqMapping.Do(context.Background(), es.client)
	if err != nil {
		log.Fatalf("Error setting mapping: %s", err)
	}
	defer resMapping.Body.Close()

	if resMapping.IsError() {
		log.Fatalf("Error setting mapping: %s", resMapping)
	}

	fmt.Printf("Mapping Index '%s' successfully\n", indexName)

	start := time.Now().UTC()

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	//
	// Create the BulkIndexer
	//
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         indexName,        // The default index name
		Client:        es.client,        // The Elasticsearch client
		NumWorkers:    numWorkers,       // The number of worker goroutines
		FlushBytes:    int(flushBytes),  // The flush threshold in bytes
		FlushInterval: 30 * time.Second, // The periodic flush interval
	})
	if err != nil {
		log.Fatalf("Error creating the indexer: %s", err)
	}
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	// Loop over the data
	//
	for _, a := range data {
		// Prepare the data payload: encode article to JSON
		//
		data, err := json.Marshal(a)
		if err != nil {
			log.Fatalf("Cannot encode article %d: %s", a.ID, err)
		}

		// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
		//
		// Add an item to the BulkIndexer
		//
		err = bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				// Action field configures the operation to perform (index, create, delete, update)
				Action: "index",

				// DocumentID is the (optional) document ID
				DocumentID: strconv.Itoa(a.ID),

				// Body is an `io.Reader` with the payload
				Body: bytes.NewReader(data),

				// OnSuccess is called for each successful operation
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					atomic.AddUint64(&countSuccessful, 1)
				},

				// OnFailure is called for each failed operation
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						log.Printf("ERROR: %s", err)
					} else {
						log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
					}
				},
			},
		)
		if err != nil {
			log.Fatalf("Unexpected error: %s", err)
		}
		// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
	}

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	// Close the indexer
	//
	if err := bi.Close(context.Background()); err != nil {
		log.Fatalf("Unexpected error: %s", err)
	}
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	biStats := bi.Stats()

	// Report the results: number of indexed docs, number of errors, duration, indexing rate
	//
	log.Println(strings.Repeat("▔", 65))

	dur := time.Since(start)

	if biStats.NumFailed > 0 {
		log.Fatalf(
			"Indexed [%s] documents with [%s] errors in %s (%s docs/sec)",
			humanize.Comma(int64(biStats.NumFlushed)),
			humanize.Comma(int64(biStats.NumFailed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),
		)
	} else {
		log.Printf(
			"Sucessfuly indexed [%s] documents in %s (%s docs/sec)",
			humanize.Comma(int64(biStats.NumFlushed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),
		)
	}
}
