package simulator

import (
	"graphCA/element"
	"graphCA/recorder"
	"graphCA/utils"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func VehicleProcess(simTime int, simpleGraph, simulationGraph *simple.DirectedGraph, allowedDestination []graph.Node) {
	checkCompletedVehicle(simTime, simpleGraph, simulationGraph, allowedDestination)
	updateVehicleActiveStatus()
	updateVehiclePosition(simTime)
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
			for {
				newDestination = allowedDestination[rand.Intn(len(allowedDestination)-1)]
				if newOrigin.ID() != newDestination.ID() {
					break
				}
			}
			newVehicle := element.NewVehicle(id, velocity, acceleration, 1.0, slowingProb, true)
			newVehicle.SetOD(simulationGraph, newOrigin, newDestination)
			shortestPath, _, err := utils.ShortestPath(simpleGraph, newOrigin, newDestination)
			if err != nil {
				panic(err)
			}

			newVehicle.SetPath(shortestPath)

			newVehicle.BufferIn(simTime)
			//waitingVehiclesMutex.Lock()
			waitingVehicles[newVehicle] = struct{}{}
			numVehiclesWaiting++
			//waitingVehiclesMutex.Unlock()
		}
	}

	completedVehicles = make(map[*element.Vehicle]struct{})
}

func updateVehicleActiveStatus() {
	recordActivatedVehicle := make(map[*element.Vehicle]struct{})

	for vehicle := range waitingVehicles {
		ok := vehicle.UpdateActiveState()
		if ok {
			recordActivatedVehicle[vehicle] = struct{}{}
		}
	}

	for vehicle := range recordActivatedVehicle {
		delete(waitingVehicles, vehicle)
		numVehiclesWaiting--
		activeVehicles[vehicle] = struct{}{}
		numVehiclesActive++
	}
}

func updateVehiclePosition(simTime int) {
	if len(activeVehicles) == 0 {
		return
	}

	for vehicle := range activeVehicles {
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
}
