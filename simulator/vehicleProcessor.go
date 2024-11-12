package simulator

import (
	"graphCA/element"
	"graphCA/recorder"
	"graphCA/utils"
	"sync"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func VehicleProcess(numWorkers, simTime int, simpleGraph, simulationGraph *simple.DirectedGraph, allowedDestination []graph.Node) {
	checkCompletedVehicle(simTime, simpleGraph, simulationGraph, allowedDestination)
	updateVehicleActiveStatus(numWorkers)
	updateVehiclePosition(numWorkers, simTime)
}

func checkCompletedVehicle(simTime int, simpleGraph, simulationGraph *simple.DirectedGraph, allowedDestination []graph.Node) {
	if len(completedVehicles) == 0 {
		return
	}

	for vehicle := range completedVehicles {
		recorder.RecordVehicleData(vehicle)
		recorder.RecordTraceData(vehicle)

		// closedVehicle 从当前位置重新进入系统
		if vehicle.Flag() {
			id, velocity, acceleration, slowingProb := vehicle.Index(), vehicle.Velocity(), vehicle.Acceleration(), vehicle.SlowingProb()
			newOrigin := vehicle.Destination()

			var newDestination graph.Node
			// for {
			// 	lim := TripDistanceLim()
			// 	newDestinationInRange := utils.AccessibleNodesWithinLim(simpleGraph, newOrigin, lim)
			// 	if len(newDestinationInRange) == 0 {
			// 		continue
			// 	}
			// 	newDestination = newDestinationInRange[rand.Intn(len(newDestinationInRange)-1)]
			// 	if newOrigin.ID() != newDestination.ID() {
			// 		break
			// 	}
			// }
			for {
				newDestination = allowedDestination[rand.Intn(len(allowedDestination)-1)]
				if newOrigin.ID() != newDestination.ID() {
					break
				}
			}
			newVehicle := element.NewVehicle(id, velocity, acceleration, 1.0, slowingProb, true)
			newVehicle.SetOD(simulationGraph, newOrigin, newDestination)

			// path, _, err := utils.ShortestPath(simpleGraph, newOrigin, newDestination)
			// path, _, err := utils.RandomPath(simpleGraph, newOrigin, newDestination)
			paths, err := utils.KShortestPaths(simpleGraph, newOrigin, newDestination, kPathsNum)
			if err != nil {
				panic(err)
			}
			randomDice := rand.Intn(len(paths))
			newVehicle.SetPath(paths[randomDice])

			newVehicle.BufferIn(simTime)
			waitingVehiclesMutex.Lock()
			waitingVehicles[newVehicle] = struct{}{}
			numVehiclesWaiting++
			waitingVehiclesMutex.Unlock()
		}
	}

	completedVehicles = make(map[*element.Vehicle]struct{})
}

func updateVehicleActiveStatus(numWorkers int) {
	if len(waitingVehicles) == 0 {
		return
	}

	vehicles := make([]*element.Vehicle, 0, len(waitingVehicles))
	for vehicle := range waitingVehicles {
		vehicles = append(vehicles, vehicle)
	}

	if len(vehicles) < numWorkers {
		numWorkers = len(vehicles)
	}

	vehiclesPerThread := len(vehicles) / numWorkers
	extraVehicles := len(vehicles) % numWorkers

	var wg sync.WaitGroup
	recordActivatedVehicle := make(chan *element.Vehicle, len(vehicles))

	start := 0
	for i := 0; i < numWorkers; i++ {
		end := start + vehiclesPerThread
		if i < extraVehicles {
			end++
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				vehicle := vehicles[j]
				ok := vehicle.UpdateActiveState()
				if ok {
					recordActivatedVehicle <- vehicle
				}
			}
		}(start, end)

		start = end
	}

	wg.Wait()
	close(recordActivatedVehicle)

	for vehicle := range recordActivatedVehicle {
		delete(waitingVehicles, vehicle)
		numVehiclesWaiting--
		activeVehicles[vehicle] = struct{}{}
		numVehiclesActive++
	}
}

func updateVehiclePosition(numWorkers, simTime int) {
	if len(activeVehicles) == 0 {
		return
	}

	vehicles := make([]*element.Vehicle, 0, len(activeVehicles))
	for vehicle := range activeVehicles {
		vehicles = append(vehicles, vehicle)
	}

	// 打乱vehicle顺序
	for i := len(vehicles) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		vehicles[i], vehicles[j] = vehicles[j], vehicles[i]
	}

	if len(vehicles) < numWorkers {
		numWorkers = len(vehicles)
	}

	vehiclesPerThread := len(vehicles) / numWorkers
	extraVehicles := len(vehicles) % numWorkers

	var wg sync.WaitGroup

	start := 0
	for i := 0; i < numWorkers; i++ {
		end := start + vehiclesPerThread
		if i < extraVehicles {
			end++
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				vehicle := vehicles[j]
				if vehicle.State() == 3 {
					vehicle.SystemIn()
				}
				if vehicle.State() == 4 {
					completed := vehicle.Move(simTime)
					if completed {
						activeVehiclesMutex.Lock()
						delete(activeVehicles, vehicle)
						numVehiclesActive--
						activeVehiclesMutex.Unlock()

						completedVehiclesMutex.Lock()
						completedVehicles[vehicle] = struct{}{}
						numVehicleCompleted++
						completedVehiclesMutex.Unlock()
					}
				}
			}
		}(start, end)

		start = end
	}

	wg.Wait()
}
