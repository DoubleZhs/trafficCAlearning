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

func GetGenerateVehicleCount(timeOfDay int, dayDemandList []float64, randomDis float64) int {
	randomFactor := 1 + (rand.Float64()*2*randomDis - randomDis)
	baseDemand := dayDemandList[timeOfDay] * randomFactor
	baseN := math.Floor(baseDemand)

	var randomN float64
	randomDice := rand.Float64()
	if randomDice < baseDemand-baseN {
		randomN = 1
	} else {
		randomN = 0
	}

	// // test
	// switch {
	// case timeOfDay <= 57600/24*2:
	// 	baseN = 1
	// case timeOfDay <= 57600/24*4:
	// 	baseN = 2
	// case timeOfDay <= 57600/24*6:
	// 	baseN = 3
	// case timeOfDay <= 57600/24*8:
	// 	baseN = 4
	// case timeOfDay <= 57600/24*10:
	// 	baseN = 5
	// default:
	// 	baseN = 0
	// }
	// randomN = 0

	return int(baseN + randomN)
}

func readDemandCSV() []float64 {
	var filename string = "./resources/DemandTimeDistribution_Smoothed.csv"
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

	var demand []float64 = make([]float64, 0)
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
