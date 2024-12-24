package simulator

import (
	"encoding/csv"
	"fmt"
	"graphCA/element"
	"os"
	"strconv"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func CreateAnaheimSimpleGraph() (*simple.DirectedGraph, map[int64]graph.Node, map[int64]graph.Edge) {
	simpleGraph := simple.NewDirectedGraph()
	simpleNodes := make(map[int64]graph.Node)
	simpleEdges := make(map[int64]graph.Edge)

	nodeFilePath := "./resources/Anaheim_Nodes.csv"

	nodeFile, err := os.Open(nodeFilePath)
	if err != nil {
		panic(fmt.Sprintf("failed to open node file: %v", err))
	}
	defer nodeFile.Close()

	nodeReader := csv.NewReader(nodeFile)
	nodeRecords, err := nodeReader.ReadAll()
	if err != nil {
		panic(fmt.Sprintf("failed to read node file: %v", err))
	}

	for _, record := range nodeRecords[1:] {
		id, _ := strconv.ParseInt(record[0], 10, 64)
		speed, _ := strconv.Atoi(record[1])

		capacity := 2.0
		if speed > 5 {
			capacity = 4.0
		}

		commonCell := element.NewCommonCell(id, speed, capacity)
		simpleNodes[id] = commonCell
		simpleGraph.AddNode(commonCell)
	}

	edgeFilePath := "./resources/Anaheim_Edges.csv"

	edgeFile, err := os.Open(edgeFilePath)
	if err != nil {
		panic(fmt.Sprintf("failed to open edge file: %v", err))
	}
	defer edgeFile.Close()

	edgeReader := csv.NewReader(edgeFile)
	edgeRecords, err := edgeReader.ReadAll()
	if err != nil {
		panic(fmt.Sprintf("failed to read edge file: %v", err))
	}

	for _, record := range edgeRecords[1:] {
		fromId, _ := strconv.ParseInt(record[0], 10, 64)
		toId, _ := strconv.ParseInt(record[1], 10, 64)
		length, _ := strconv.Atoi(record[2])
		speed, _ := strconv.Atoi(record[3])
		designCapacity, _ := strconv.ParseFloat(record[4], 64)

		capacity := 2.0
		if speed > 5 {
			capacity = 4.0
		}

		link := element.NewLink(fromId*1000+toId, simpleNodes[fromId], simpleNodes[toId], length, speed, capacity, designCapacity)
		simpleEdges[link.ID()] = link
		simpleGraph.SetEdge(link)
	}

	return simpleGraph, simpleNodes, simpleEdges
}

func CreateAnaheimSimulationGraph(simpleG *simple.DirectedGraph, simpleNodes map[int64]graph.Node, simpleEdges map[int64]graph.Edge, inDegreeThreshold, initPhaseInterval int) (*simple.DirectedGraph, map[int64]graph.Node, map[int64][]*element.TrafficLightCell) {
	simulationGraph := simple.NewDirectedGraph()
	simulationNodes := make(map[int64]graph.Node)

	for id, node := range simpleNodes {
		switch n := node.(type) {
		case *element.CommonCell:
			simulationGraph.AddNode(n)
			simulationNodes[id] = n
		default:
			panic("Node is not a common cell")
		}
	}

	for _, link := range simpleEdges {
		switch n := link.(type) {
		case *element.Link:
			n.AddToSimGraph(simulationGraph)
			for _, cell := range n.Flat() {
				simulationNodes[cell.ID()] = cell
			}
		default:
			panic("Node is not a link")
		}
	}

	// 按入度识别需要循环的红绿灯
	var lightGroups map[int64][]*element.TrafficLightCell = make(map[int64][]*element.TrafficLightCell)
	for _, node := range simulationNodes {
		if _, exists := simpleNodes[node.ID()]; exists && simulationGraph.To(node.ID()).Len() >= inDegreeThreshold {
			n := simulationGraph.To(node.ID())
			lightNodes := make([]*element.TrafficLightCell, 0)
			for n.Next() {
				ln := n.Node()
				lightNode, ok := ln.(*element.TrafficLightCell)
				if !ok {
					panic("Node is not a traffic light cell")
				}
				lightNodes = append(lightNodes, lightNode)
			}
			lightGroups[node.ID()] = lightNodes
		}
	}

	for _, lightGroup := range lightGroups {
		initLightGroupInterval(lightGroup, initPhaseInterval)
	}

	return simulationGraph, simulationNodes, lightGroups
}

func initLightGroupInterval(lightGroup []*element.TrafficLightCell, initPhaseInterval int) {
	for i, light := range lightGroup {
		light.SetInterval(initPhaseInterval*len(lightGroup), [2]int{initPhaseInterval * i, initPhaseInterval * (i + 1)})
	}
}
