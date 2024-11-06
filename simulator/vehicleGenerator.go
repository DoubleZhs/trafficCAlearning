package simulator

import (
	"graphCA/element"
	"graphCA/utils"
	"sync"
	"sync/atomic"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func getNextVehicleID() int64 {
	return atomic.AddInt64(&numVehicleGenerated, 1)
}

func randomVelocity() int {
	return 1 + rand.Intn(2)
}

func randomAcceleration() int {
	return 1 + rand.Intn(3)
}

func randomSlowingProbability() float64 {
	return rand.Float64() / 5.0
}

func InitFixedVehicle(n int, simpleGraph, simulationGraph *simple.DirectedGraph, allowedOrigin, allowedDestination []graph.Node) {
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()

			vehicle := element.NewVehicle(getNextVehicleID(), randomVelocity(), randomAcceleration(), 1.0, randomSlowingProbability(), true) // ClosedVehicle = true

			// OD生成
			oCell := allowedOrigin[rand.Intn(len(allowedOrigin)-1)]

			var dCell graph.Node
			for {
				lim := TripDistanceLim()
				dCellsInRange := utils.AccessibleNodesWithinLim(simpleGraph, oCell, lim)
				if len(dCellsInRange) == 0 {
					continue
				}
				dCell = dCellsInRange[rand.Intn(len(dCellsInRange)-1)]
				if oCell.ID() != dCell.ID() {
					break
				}
			}
			vehicle.SetOD(simulationGraph, oCell, dCell)

			// 路径
			path, _, err := utils.ShortestPath(simpleGraph, oCell, dCell)
			// path, _, err := utils.RandomPath(simpleGraph, oCell, dCell)
			if err != nil {
				panic(err)
			}
			vehicle.SetPath(path)

			// 进入缓冲区
			vehicle.BufferIn(0)

			waitingVehiclesMutex.Lock()
			waitingVehicles[vehicle] = struct{}{}
			numVehiclesWaiting++
			waitingVehiclesMutex.Unlock()
		}()
	}
	wg.Wait()
}

func GenerateScheduleVehicle(simTime, n int, simpleGraph, simulationGraph *simple.DirectedGraph, allowedOrigin, allowedDestination []graph.Node) {
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()

			vehicle := element.NewVehicle(getNextVehicleID(), randomVelocity(), randomAcceleration(), 1.0, randomSlowingProbability(), false)

			// OD生成
			oCell := allowedOrigin[rand.Intn(len(allowedOrigin)-1)]

			var dCell graph.Node
			for {
				lim := TripDistanceLim()
				dCellsInRange := utils.AccessibleNodesWithinLim(simpleGraph, oCell, lim)
				if len(dCellsInRange) == 0 {
					continue
				}
				dCell = dCellsInRange[rand.Intn(len(dCellsInRange)-1)]
				if oCell.ID() != dCell.ID() {
					break
				}
			}
			vehicle.SetOD(simulationGraph, oCell, dCell)

			// 路径
			path, _, err := utils.ShortestPath(simpleGraph, oCell, dCell)
			// path, _, err := utils.RandomPath(simpleGraph, oCell, dCell)
			if err != nil {
				panic(err)
			}
			vehicle.SetPath(path)

			vehicle.BufferIn(simTime)

			waitingVehiclesMutex.Lock()
			waitingVehicles[vehicle] = struct{}{}
			numVehiclesWaiting++
			waitingVehiclesMutex.Unlock()
		}()
	}
	wg.Wait()
}
