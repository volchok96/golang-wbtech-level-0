package unmarshal

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"wb-nats-service/internal/models"
)

func ReadOrdersFromFiles(jsonFiles []string) []models.Order {
	var orders []models.Order

	for _, jsonFile := range jsonFiles {
		jsonData, err := os.ReadFile(jsonFile)
		if err != nil {
			log.Fatalf("Error reading JSON file: %v", err)
		}

		var order models.Order
		err = json.Unmarshal(jsonData, &order)
		if err != nil {
			log.Fatalf("Error unmarshalling JSON: %v", err)
		}

		orders = append(orders, order)
	}

	return orders
}

func ReadOrdersFromDirectory(dir string) []models.Order {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}

	var orders []models.Order
	var jsonFiles []string

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			jsonFiles = append(jsonFiles, filepath.Join(dir, file.Name()))
		}
	}

	orders = ReadOrdersFromFiles(jsonFiles)

	return orders
}