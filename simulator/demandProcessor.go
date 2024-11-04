package simulator

import (
	"encoding/csv"
	"log"
	"math"
	"os"
	"strconv"

	"golang.org/x/exp/rand"
)

var rawDemand []float64 = readDemandCSV()

func AdjustDemand(A, B float64) []float64 {
	adjustedDemand := make([]float64, len(rawDemand))
	for i, d := range rawDemand {
		adjustedDemand[i] = d*A + B
	}
	return adjustedDemand
}

func GetGenerateVehicleCount(timeOfDay int, dayDemandList []float64) int {
	baseDemand := dayDemandList[timeOfDay]
	baseN := math.Floor(baseDemand)

	var randomN float64
	randomDice := rand.Float64()
	if randomDice < baseDemand-baseN {
		randomN = 1
	} else {
		randomN = 0
	}

	return int(baseN + randomN)
}

func readDemandCSV() []float64 {
	var filename string = "./resources/DemandTimeDistribution.csv"
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file: %s", err)
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read csv: %s", err)
		panic(err)
	}

	var demand []float64 = make([]float64, 57600)
	for _, record := range records[1:] {
		pro, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Printf("Failed to parse probability: %s", err)
			panic(err)
		}
		demand = append(demand, pro)
	}

	return demand
}
