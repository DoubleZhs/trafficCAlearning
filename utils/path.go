package utils

import (
	"container/heap"
	"errors"
	"graphCA/element"
	"math"
	"math/rand/v2"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

// NodeHeap is a priority queue for graph nodes
type NodeHeap []nodeDist

type nodeDist struct {
	node graph.Node
	dist float64
}

func (h NodeHeap) Len() int           { return len(h) }
func (h NodeHeap) Less(i, j int) bool { return h[i].dist < h[j].dist }
func (h NodeHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *NodeHeap) Push(x interface{}) {
	*h = append(*h, x.(nodeDist))
}

func (h *NodeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

func ShortestPath(g *simple.DirectedGraph, origin, destination graph.Node) ([]graph.Node, float64, error) {
	dist := make(map[int64]float64)
	prev := make(map[int64]graph.Node)
	nodeHeap := &NodeHeap{}
	heap.Init(nodeHeap)

	nodes := g.Nodes()
	for nodes.Next() {
		node := nodes.Node()
		dist[node.ID()] = math.Inf(1)
	}
	dist[origin.ID()] = 0
	heap.Push(nodeHeap, nodeDist{node: origin, dist: 0})

	for nodeHeap.Len() > 0 {
		u := heap.Pop(nodeHeap).(nodeDist).node

		if u.ID() == destination.ID() {
			break
		}

		for neighbors := g.From(u.ID()); neighbors.Next(); {
			v := neighbors.Node()
			edge := g.Edge(u.ID(), v.ID()).(*element.Link)
			alt := dist[u.ID()] + edge.Weight()

			if alt < dist[v.ID()] {
				dist[v.ID()] = alt
				prev[v.ID()] = u
				heap.Push(nodeHeap, nodeDist{node: v, dist: alt})
			}
		}
	}

	if dist[destination.ID()] == math.Inf(1) {
		return nil, -1, errors.New("no path found")
	}

	path := []graph.Node{}
	for u := destination; u != nil; {
		path = append([]graph.Node{u}, path...)
		if prevNode, ok := prev[u.ID()]; ok {
			u = prevNode
		} else {
			break
		}
	}

	return path, dist[destination.ID()], nil
}

// func ShortestPath(g *simple.DirectedGraph, origin, destination graph.Node) ([]graph.Node, float64, error) {
// 	// 检查缓存中是否已经计算过路径
// 	if cachedPath, ok := pCache.Get(origin, destination); ok {
// 		return cachedPath, float64(len(cachedPath) - 1), nil
// 	}

// 	// 如果缓存中没有，则计算路径
// 	shortestPath, length := path.DijkstraFrom(origin, g).To(destination.ID())
// 	if length == 0 {
// 		return nil, -1, errors.New("no path found")
// 	}

// 	// 写入缓存
// 	pCache.Set(origin, destination, shortestPath, length)

// 	return shortestPath, length, nil
// }

func RandomPath(g *simple.DirectedGraph, origin, destination graph.Node) ([]graph.Node, float64, error) {
	visited := make(map[int64]bool)
	var path []graph.Node

	var dfs func(node graph.Node) bool
	dfs = func(node graph.Node) bool {
		if node.ID() == destination.ID() {
			path = append(path, node)
			return true
		}

		visited[node.ID()] = true
		neighbors := g.From(node.ID())
		neighborList := make([]graph.Node, 0, neighbors.Len())
		for neighbors.Next() {
			neighborList = append(neighborList, neighbors.Node())
		}

		// Shuffle neighbors to introduce randomness
		rand.Shuffle(len(neighborList), func(i, j int) {
			neighborList[i], neighborList[j] = neighborList[j], neighborList[i]
		})

		for _, neighbor := range neighborList {
			if !visited[neighbor.ID()] {
				if dfs(neighbor) {
					path = append(path, node)
					return true
				}
			}
		}

		return false
	}

	if !dfs(origin) {
		return nil, -1, errors.New("no path found")
	}

	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path, float64(len(path) - 1), nil
}

func KShortestPaths(g *simple.DirectedGraph, origin, destination graph.Node, k int, linksMap map[[2]int64]*element.Link) ([][]graph.Node, error) {
	if k <= 0 {
		return nil, errors.New("k must be greater than 0")
	}

	type pathWithCost struct {
		path []graph.Node
		cost float64
	}

	// Find the shortest path using Dijkstra's algorithm
	shortestPath, cost, err := ShortestPath(g, origin, destination)
	if err != nil {
		return nil, err
	}

	// Ensure the shortest path ends at the destination
	if len(shortestPath) == 0 || shortestPath[len(shortestPath)-1].ID() != destination.ID() {
		return nil, errors.New("shortest path does not end at destination")
	}

	paths := []pathWithCost{{path: shortestPath, cost: cost}}
	resultPaths := [][]graph.Node{shortestPath}

	for i := 1; i < k; i++ {
		var bestCandidate pathWithCost
		foundCandidate := false

		for j := 0; j < len(paths[i-1].path)-1; j++ {
			spurNode := paths[i-1].path[j]
			rootPath := paths[i-1].path[:j+1]

			// Create a copy of the graph to avoid modifying the original graph
			gCopy := simple.NewDirectedGraph()
			nodes := g.Nodes()
			for nodes.Next() {
				node := nodes.Node()
				gCopy.AddNode(node)
			}
			edges := g.Edges()
			for edges.Next() {
				edge := edges.Edge()
				gCopy.SetEdge(edge)
			}

			// Remove edges that are part of the rootPath
			for _, p := range resultPaths {
				if len(p) > j && equalPaths(rootPath, p[:j+1]) {
					gCopy.RemoveEdge(p[j].ID(), p[j+1].ID())
				}
			}

			spurPath, _, err := ShortestPath(gCopy, spurNode, destination)
			if err == nil {
				totalPath := append(rootPath[:len(rootPath)-1], spurPath...)
				totalCost := pathCost(totalPath, linksMap)
				candidate := pathWithCost{path: totalPath, cost: totalCost}

				// Ensure the candidate path ends at the destination
				if len(candidate.path) == 0 || candidate.path[len(candidate.path)-1].ID() != destination.ID() {
					continue
				}

				if !foundCandidate || candidate.cost < bestCandidate.cost {
					bestCandidate = candidate
					foundCandidate = true
				}
			}
		}

		if !foundCandidate {
			break
		}

		paths = append(paths, bestCandidate)
		resultPaths = append(resultPaths, bestCandidate.path)
	}

	return resultPaths, nil
}

func equalPaths(p1, p2 []graph.Node) bool {
	if len(p1) != len(p2) {
		return false
	}
	for i := range p1 {
		if p1[i].ID() != p2[i].ID() {
			return false
		}
	}
	return true
}
