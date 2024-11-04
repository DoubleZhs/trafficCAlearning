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

func CreateAnaheimSimpleGraph() (*simple.DirectedGraph, map[int64]graph.Node) {
	simpleGraph := simple.NewDirectedGraph()

	edgeFilePath := "./resources/Anaheim_Edges.csv"
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

	simpleNodes := make(map[int64]graph.Node)

	for _, record := range nodeRecords[1:] {
		nodeID, _ := strconv.ParseInt(record[0], 10, 64)
		length, _ := strconv.Atoi(record[1])
		speed, _ := strconv.Atoi(record[2])

		capacity := 2.0
		if speed > 5 {
			capacity = 4.0
		}

		if length == -1 {
			commonCell := element.NewCommonCell(nodeID, speed, capacity)
			simpleNodes[nodeID] = commonCell
			simpleGraph.AddNode(commonCell)
		} else {
			link := element.NewLink(nodeID, length, speed, capacity)
			simpleNodes[nodeID] = link
			simpleGraph.AddNode(link)
		}
	}

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
		initNodeID, _ := strconv.ParseInt(record[0], 10, 64)
		termNodeID, _ := strconv.ParseInt(record[1], 10, 64)

		initNode := simpleNodes[initNodeID]
		termNode := simpleNodes[termNodeID]

		simpleGraph.SetEdge(simple.Edge{F: initNode, T: termNode})
	}

	return simpleGraph, simpleNodes
}

func CreateAnaheimSimulationGraph(simpleG *simple.DirectedGraph, nodes map[int64]graph.Node) (*simple.DirectedGraph, map[int64]graph.Node) {
	simulationGraph := simple.NewDirectedGraph()
	simulationNodes := make(map[int64]graph.Node)

	// Add all nodes to the simulation graph
	for id, node := range nodes {
		switch n := node.(type) {
		case *element.CommonCell:
			simulationGraph.AddNode(n)
			simulationNodes[id] = n
		case *element.Link:
			n.AddToGraph(simulationGraph)
			for _, cell := range n.Flat() {
				simulationNodes[cell.ID()] = cell
			}
		}
	}

	// Read edges from the edge file
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
		initNodeID, _ := strconv.ParseInt(record[0], 10, 64)
		termNodeID, _ := strconv.ParseInt(record[1], 10, 64)

		initNode := nodes[initNodeID]
		termNode := nodes[termNodeID]

		switch initNode := initNode.(type) {
		case *element.CommonCell:
			switch termNode := termNode.(type) {
			case *element.CommonCell:
				simulationGraph.SetEdge(simple.Edge{F: initNode, T: termNode})
			case *element.Link:
				termNode.AddFromNode(simulationGraph, initNode)
			}
		case *element.Link:
			switch termNode := termNode.(type) {
			case *element.CommonCell:
				initNode.AddToNode(simulationGraph, termNode)
			case *element.Link:
				initNode.AddToNode(simulationGraph, termNode)
				termNode.AddFromNode(simulationGraph, initNode)
			}
		}
	}

	return simulationGraph, simulationNodes
}
