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

func GetVehiclesNum() (int64, int64, int64, int64) {
	activeVehiclesMutex.Lock()
	defer activeVehiclesMutex.Unlock()
	waitingVehiclesMutex.Lock()
	defer waitingVehiclesMutex.Unlock()
	completedVehiclesMutex.Lock()
	defer completedVehiclesMutex.Unlock()

	return numVehicleGenerated, numVehiclesActive, numVehiclesWaiting, numVehicleCompleted
}
