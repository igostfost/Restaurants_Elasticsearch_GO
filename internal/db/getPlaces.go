package db

import (
	"bytes"
	"context"
	"elasticTask/pkg/types"
	"encoding/json"
	"log"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// GetPlaces Находит места из Elasticsearch в нужном количестве для отображения в html
// limit - количесвто мест
// offset - смещение (начальная позиция) для запроса к Elasticsearch
func (es ElasticsearchStore) GetPlaces(limit int, offset int) ([]types.Place, int, error) {
	query := map[string]interface{}{
		"from": offset,
		"size": limit,
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, 0, err
	}

	res, err := es.client.Search(
		es.client.Search.WithContext(context.Background()),
		es.client.Search.WithIndex(es.indexName),
		es.client.Search.WithBody(&buf),
		es.client.Search.WithTrackTotalHits(true),
		es.client.Search.WithPretty(),
	)

	if err != nil {
		log.Fatalf("Elasticsearch Search() API ERROR: %v", err)
	}
	if res.IsError() {
		log.Fatalf("Elasticsearch Search() API ERROR: %v", res)
	}

	defer res.Body.Close()
	places, totalValue, err := ConvertResultsToPlaces(res)
	return places, totalValue, nil
}

// GetRecommendPlaces Находит самые близкие рекомендованные места по долшоте и широте.
// limit - количесвто мест
// lat - широта
// lon - долгота
func (es ElasticsearchStore) GetRecommendPlaces(limit int, lat, lon float64) ([]types.Place, int, error) {
	query := map[string]interface{}{
		"size": limit,
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"sort": []map[string]interface{}{
			{
				"_script": map[string]interface{}{
					"type": "number",
					"script": map[string]interface{}{
						"lang":   "painless",
						"source": "doc['location'].arcDistance(params.lat, params.lon)",
						"params": map[string]interface{}{
							"lat": lat,
							"lon": lon,
						},
					},
					"order": "asc",
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, 0, err
	}

	res, err := es.client.Search(
		es.client.Search.WithContext(context.Background()),
		es.client.Search.WithIndex(es.indexName),
		es.client.Search.WithBody(&buf),
		es.client.Search.WithTrackTotalHits(true),
		es.client.Search.WithPretty(),
	)

	if err != nil {
		log.Fatalf("Elasticsearch Search() API ERROR: %v", err)
	}
	if res.IsError() {
		log.Fatalf("Elasticsearch Search() API ERROR: %v", res)
	}

	defer res.Body.Close()

	places, totalValue, err := ConvertResultsToPlaces(res)
	return places, totalValue, nil
}

// ConvertResultsToPlaces преобразует ответ Elasticsearch в слайс мест и общее количество найденных мест.
// res - ответ Elasticsearch
func ConvertResultsToPlaces(res *esapi.Response) ([]types.Place, int, error) {

	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, 0, err
	}

	// Извлечение массива "hits" из секции "hits" в JSON-ответе Elasticsearch
	hits := response["hits"].(map[string]interface{})["hits"].([]interface{})

	totalValue := int(response["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))

	// Слайс для хранения мест
	var places []types.Place
	// Счетчик найденных мест
	count := 0

	for _, hit := range hits {
		count++
		// Преобразование интерфейса в map[string]interface{}
		hitMap := hit.(map[string]interface{})
		// Извлечение "_source" из текущего "hit"
		source := hitMap["_source"].(map[string]interface{})
		// Создание объекта типа Place на основе данных из "_source"
		place := types.Place{
			ID:      int(source["id"].(float64)),
			Name:    source["name"].(string),
			Address: source["address"].(string),
			Phone:   source["phone"].(string),
			Location: types.GeoJSON{
				Latitude:  source["location"].(map[string]interface{})["lat"].(float64),
				Longitude: source["location"].(map[string]interface{})["lon"].(float64),
			},
		}
		places = append(places, place)
	}
	return places, totalValue, nil
}
