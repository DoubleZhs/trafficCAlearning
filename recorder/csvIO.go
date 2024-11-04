package recorder

import (
	"encoding/csv"
	"log"
	"os"
)

func initializeCSV(filename string, header []string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create file: %s", err)
		// 创建失败直接退出
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write(header)
}

func appendToCSV(filename string, data [][]string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open file: %s", err)
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.WriteAll(data); err != nil {
		log.Fatalf("Failed to write to file: %s", err)
		panic(err)
	}
}
