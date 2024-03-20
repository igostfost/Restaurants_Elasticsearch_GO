package main

import (
	"elasticTask/internal/db"
	"elasticTask/web"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	var indexName = "places"
	log.SetFlags(0)

	store, err := db.NewElasticsearchStore(indexName)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	store.Indexeres(indexName)
	fmt.Println("Server started...")

	time.Sleep(1 * time.Second)

	http.HandleFunc("/web/places", web.HtmlHandler(store))
	http.HandleFunc("/web/recommend", web.HtmlRecommendHandler(store))
	http.HandleFunc("/api/places", web.JsonHandler(store))
	//http.HandleFunc("/api/recommend", web.JsonRecommendHandler(store))
	http.HandleFunc("/api/get_token", web.TokenHandler(store))
	http.Handle("/api/recommend", web.AuthMiddleware(http.HandlerFunc(web.JsonRecommendHandler(store))))

	http.ListenAndServe(":8888", nil)
}
