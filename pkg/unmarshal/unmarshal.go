package unmarshal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"wb-kafka-service/internal/models"
	"wb-kafka-service/pkg/logger"
)

func ReadOrdersFromFiles(log *logger.Logger, jsonFiles []string) []models.Order {
	var orders []models.Order

	for _, jsonFile := range jsonFiles {
		jsonData, err := os.ReadFile(jsonFile)
		if err != nil {
			log.Error("Error reading JSON file", err)
			continue 
		}

		var order models.Order
		err = json.Unmarshal(jsonData, &order)
		if err != nil {
			log.Error("Error unmarshalling JSON", err)
			continue 
		}

		orders = append(orders, order)
	}

	return orders
}

func ReadOrdersFromDirectory(log *logger.Logger, dir string) []models.Order {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Error("Error reading directory", err)
		return nil 
	}

	var orders []models.Order
	var jsonFiles []string

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			jsonFiles = append(jsonFiles, filepath.Join(dir, file.Name()))
		}
	}

	orders = ReadOrdersFromFiles(log, jsonFiles)

	return orders
}
