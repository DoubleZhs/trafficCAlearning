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

func InitFixedVehicle(n int, simpleGraph, simulationGraph *simple.DirectedGraph, linksMap map[[2]int64]*element.Link, allowedODNodes [][2]graph.Node) {
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()

			vehicle := element.NewVehicle(getNextVehicleID(), randomVelocity(), randomAcceleration(), 1.0, randomSlowingProbability(), true) // ClosedVehicle = true

			// OD生成
			od := allowedODNodes[rand.Intn(len(allowedODNodes)-1)]
			oCell := od[0]
			dCell := od[1]

			vehicle.SetOD(simulationGraph, oCell, dCell)

			// 路径
			path, _, err := utils.ShortestPath(simpleGraph, oCell, dCell)
			if err != nil {
				panic(err)
			}
			vehicle.SetPath(path, linksMap)

			// paths, err := utils.KShortestPaths(simpleGraph, oCell, dCell, 3, linksMap)
			// if err != nil {
			// 	panic(err)
			// }
			// path := utils.LogitChoose(paths, 0.5, linksMap)
			// _, err = vehicle.SetPath(path, linksMap)
			// if err != nil {
			// 	panic(err)
			// }

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

func GenerateScheduleVehicle(simTime, n int, simpleGraph, simulationGraph *simple.DirectedGraph, linksMap map[[2]int64]*element.Link, allowedODNodes [][2]graph.Node) {
	if numVehiclesWaiting > maxNumVehiclesWaiting {
		return
	}
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()

			vehicle := element.NewVehicle(getNextVehicleID(), randomVelocity(), randomAcceleration(), 1.0, randomSlowingProbability(), false)

			// OD生成
			od := allowedODNodes[rand.Intn(len(allowedODNodes)-1)]
			oCell := od[0]
			dCell := od[1]
			vehicle.SetOD(simulationGraph, oCell, dCell)

			// 路径
			path, _, err := utils.ShortestPath(simpleGraph, oCell, dCell)
			if err != nil {
				panic(err)
			}
			vehicle.SetPath(path, linksMap)

			// paths, err := utils.KShortestPaths(simpleGraph, oCell, dCell, 5, linksMap)
			// if err != nil {
			// 	panic(err)
			// }
			// path := utils.LogitChoose(paths, 0.1, linksMap)
			// _, err = vehicle.SetPath(path, linksMap)
			// if err != nil {
			// 	panic(err)
			// }

			vehicle.BufferIn(simTime)

			waitingVehiclesMutex.Lock()
			waitingVehicles[vehicle] = struct{}{}
			numVehiclesWaiting++
			waitingVehiclesMutex.Unlock()
		}()
	}
	wg.Wait()
}
