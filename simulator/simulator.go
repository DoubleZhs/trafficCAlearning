package simulator

import (
	"graphCA/element"
	"sync"
)

var (
	activeVehicles    map[*element.Vehicle]struct{} = make(map[*element.Vehicle]struct{})
	waitingVehicles   map[*element.Vehicle]struct{} = make(map[*element.Vehicle]struct{})
	completedVehicles map[*element.Vehicle]struct{} = make(map[*element.Vehicle]struct{})

	activeVehiclesMutex    *sync.Mutex = &sync.Mutex{}
	waitingVehiclesMutex   *sync.Mutex = &sync.Mutex{}
	completedVehiclesMutex *sync.Mutex = &sync.Mutex{}

	numVehicleGenerated int64
	numVehiclesActive   int64
	numVehiclesWaiting  int64
	numVehicleCompleted int64
)

const (
	kPathsNum             int   = 3
	maxNumVehiclesWaiting int64 = 10000
)

func GetVehiclesNum() (int64, int64, int64, int64) {
	activeVehiclesMutex.Lock()
	defer activeVehiclesMutex.Unlock()
	waitingVehiclesMutex.Lock()
	defer waitingVehiclesMutex.Unlock()
	completedVehiclesMutex.Lock()
	defer completedVehiclesMutex.Unlock()

	return numVehicleGenerated, numVehiclesActive, numVehiclesWaiting, numVehicleCompleted
}

func GetVehiclesOnRoad() map[*element.Vehicle]struct{} {
	activeVehiclesMutex.Lock()
	defer activeVehiclesMutex.Unlock()

	vehiclesOnRaod := make(map[*element.Vehicle]struct{})
	for vehicle := range activeVehicles {
		if vehicle.State() == 4 {
			vehiclesOnRaod[vehicle] = struct{}{}
		}
	}

	return vehiclesOnRaod
}

func GetAverageSpeed_Density(vehiclesOnRaod map[*element.Vehicle]struct{}, numNodes int) (float64, float64) {
	totalSpeed := 0.0
	for vehicle := range vehiclesOnRaod {
		totalSpeed += float64(vehicle.Velocity())
	}
	averageSpeed := totalSpeed / float64(len(vehiclesOnRaod))

	density := float64(len(vehiclesOnRaod)) / float64(numNodes)

	return averageSpeed, density
}
