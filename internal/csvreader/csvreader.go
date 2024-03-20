package csvreader

import (
	"elasticTask/pkg/types"
	"encoding/csv"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func CsvReader() []*types.Place {

	csvFilePath, err := filepath.Abs("data/data.csv")

	// Открываем CSV файл
	file, err := os.Open(csvFilePath)
	if err != nil {
		log.Fatalf("Error opening CSV file: %s", err)
	}
	defer file.Close()

	// Создаем читатель CSV
	reader := csv.NewReader(file)
	reader.Comma = '\t' // Установка разделителя табуляции
	records, err := reader.ReadAll()

	// Преобразуем записи в объекты Place
	var places []*types.Place
	for i, record := range records {
		if i == 0 {
			// Пропускаем первую строку (заголовки)
			continue
		}

		// Преобразуем строки в соответствующие типы данных
		id, _ := strconv.Atoi(record[0])
		lat, _ := strconv.ParseFloat(record[5], 64)
		lon, _ := strconv.ParseFloat(record[4], 64)

		place := types.Place{
			ID:      id,
			Name:    record[1],
			Address: record[2],
			Phone:   record[3],
			Location: types.GeoJSON{
				Latitude:  lat,
				Longitude: lon,
			},
		}

		places = append(places, &place)

	}
	return places
}
