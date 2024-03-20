package web

import (
	"context"
	"elasticTask/internal/db"
	"elasticTask/internal/utils"
	"elasticTask/pkg/types"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

func HtmlHandler(es *db.ElasticsearchStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			http.Error(w, "Invalid 'page' value: '"+pageStr+"'", http.StatusBadRequest)
			return
		}

		limit := 10
		offset := (page - 1) * limit

		places, total, err := es.GetPlaces(limit, offset)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		lastPage := total / limit
		hasPrev := page > 1
		hasNext := page < lastPage
		first := 1

		if page > lastPage {
			http.Error(w, "Invalid 'page' value: '"+pageStr+"'", http.StatusBadRequest)
			return
		}

		data := types.PageData{
			Total:     total,
			Places:    places,
			HasPrev:   hasPrev,
			PrevPage:  page - 1,
			HasNext:   hasNext,
			NextPage:  page + 1,
			LastPage:  lastPage,
			FirstPage: first,
		}
		tmpl, err := template.ParseFiles("web/template_style.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			//fmt.Println("hgfdcsfghjuhgsa")
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

func JsonHandler(es *db.ElasticsearchStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			http.Error(w, "Invalid 'page' value: '"+pageStr+"'", http.StatusBadRequest)
			return
		}

		limit := 10
		offset := (page - 1) * limit

		places, total, err := es.GetPlaces(limit, offset)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		lastPage := total / limit
		hasPrev := page > 1
		hasNext := page < lastPage

		if page > lastPage {
			http.Error(w, "Invalid 'page' value: '"+pageStr+"'", http.StatusBadRequest)
			return
		}

		data := types.PageData{
			Total:    total,
			Places:   places,
			HasPrev:  hasPrev,
			PrevPage: page - 1,
			HasNext:  hasNext,
			NextPage: page + 1,
			LastPage: lastPage,
		}

		// Преобразуем данные в формат JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Устанавливаем заголовок Content-Type на application/json
		w.Header().Set("Content-Type", "application/json")

		// Отправляем данные JSON в ответ
		w.Write(jsonData)
	}
}

func HtmlRecommendHandler(es *db.ElasticsearchStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получение параметров из URL
		latStr := r.URL.Query().Get("lat")
		lonStr := r.URL.Query().Get("lon")

		// Преобразование параметров в числа
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			http.Error(w, "Invalid 'lat' value", http.StatusBadRequest)
			return
		}

		lon, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			http.Error(w, "Invalid 'lon' value", http.StatusBadRequest)
			return
		}

		limit := 3
		// Выполнение запроса Elasticsearch
		places, total, err := es.GetRecommendPlaces(limit, lat, lon)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := types.Recommendation{
			Places: places,
			Total:  total,
		}

		tmpl, err := template.ParseFiles("web/rec_template_style.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			//fmt.Println("hgfdcsfghjuhgsa")
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

func JsonRecommendHandler(es *db.ElasticsearchStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получение параметров из URL
		latStr := r.URL.Query().Get("lat")
		lonStr := r.URL.Query().Get("lon")

		// Преобразование параметров в числа
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			http.Error(w, "Invalid 'lat' value", http.StatusBadRequest)
			return
		}

		lon, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			http.Error(w, "Invalid 'lon' value", http.StatusBadRequest)
			return
		}

		limit := 3
		// Выполнение запроса Elasticsearch
		places, total, err := es.GetRecommendPlaces(limit, lat, lon)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := types.Recommendation{
			Places: places,
			Total:  total,
		}

		// Преобразовываем данные в формат JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Устанавливаем заголовок Content-Type на application/json
		w.Header().Set("Content-Type", "application/json")

		// Отправляем данные JSON в ответ
		w.Write(jsonData)
	}
}

// Middleware для проверки токена перед доступом к защищенному ресурсу
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		// Проверка токена
		claims, err := utils.VerifyToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		fmt.Println("Authorization Succes")

		// Добавление информации о токене в контекст запроса
		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TokenHandler генерирует токен и возвращает его
func TokenHandler(es *db.ElasticsearchStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Генерация токена
		token, err := utils.GenerateToken()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Проверка, чтобы убедиться, что токен не пустой
		if token == "" {
			http.Error(w, "Token generation failed", http.StatusInternalServerError)
			return
		}

		// Формирование ответа
		response := map[string]string{"token": token}
		jsonResponse, _ := json.Marshal(response)

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
	}
}
