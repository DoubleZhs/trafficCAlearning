package main

import (
	"fmt"
	"graphCA/element"
	"graphCA/recorder"
	"graphCA/simulator"

	"gonum.org/v1/gonum/graph"
)

const numClosedVehicle int = 10
const simDay int = 1

func init() {
	recorder.InitLinkDataCSV("linkData.csv")
	recorder.InitVehicleDataCSV("vehicleData.csv")
	recorder.InitTraceDataCSV("traceData.csv")
}

func main() {
	simpleG, simpleNodes := simulator.CreateAnaheimSimpleGraph()
	simulationG, _ := simulator.CreateAnaheimSimulationGraph(simpleG, simpleNodes)

	keyNodes := make(map[int64]graph.Node)
	linkNodes := make(map[int64]graph.Node)
	for i, node := range simpleNodes {
		if i <= 416 {
			keyNodes[i] = node
		} else {
			linkNodes[i] = node
		}
	}

	var allowedOrigin, allowedDestination []graph.Node
	for i, node := range simpleNodes {
		if i <= 416 {
			allowedOrigin = append(allowedOrigin, node)
			allowedDestination = append(allowedDestination, node)
		}
	}

	fmt.Println(len(keyNodes))
	fmt.Println(len(linkNodes))

	simulator.InitFixedVehicle(numClosedVehicle, simpleG, simulationG, allowedOrigin, allowedDestination)

	var demand []float64
	for timeStep := 0; timeStep < simDay*57600; timeStep++ {
		timeOfDay := timeStep % 57600
		//currentDay := timeStep/57600 + 1

		if timeOfDay == 0 {
			demand = simulator.AdjustDemand(100, 1)
		}

		generateNum := simulator.GetGenerateVehicleCount(timeOfDay, demand)
		simulator.GenerateScheduleVehicle(timeStep, generateNum, simpleG, simulationG, allowedOrigin, allowedDestination)

		simulator.VehicleProcess(timeStep, simpleG, simulationG, allowedDestination)

		numVehicleGenerated, numVehiclesActive, numVehiclesWaiting, numVehicleCompleted := simulator.GetVehiclesNum()

		fmt.Println("Time step:", timeStep, "Generated:", numVehicleGenerated, "Active:", numVehiclesActive, "Waiting:", numVehiclesWaiting, "Completed:", numVehicleCompleted)

		links := make([]*element.Link, 0)
		for _, node := range linkNodes {
			assertlink, ok := node.(*element.Link)
			if !ok {
				fmt.Println("Link ID:", assertlink.ID())
				panic("Node is not a link")
			}
			links = append(links, assertlink)
		}
		recorder.RecordLinkData(timeStep, links)

		if timeOfDay%1200 == 0 {
			recorder.WriteToLinkDataCSV("linkData.csv", timeStep, links)
			recorder.WriteToVehicleDataCSV("vehicleData.csv")
			recorder.WriteToTraceDataCSV("traceData.csv")
		}
	}
}
