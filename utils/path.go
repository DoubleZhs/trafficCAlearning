package utils

import (
	"errors"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

var (
	cache = newPathCache()
)

func ShortestPath(g *simple.DirectedGraph, origin, destination graph.Node) ([]graph.Node, float64, error) {
	// 检查缓存中是否已经计算过路径
	if cachedPath, ok := cache.Get(origin, destination); ok {
		return cachedPath, float64(len(cachedPath) - 1), nil
	}

	// 如果缓存中没有，则计算路径
	shortestPath, length := path.DijkstraFrom(origin, g).To(destination.ID())
	if length == 0 {
		return nil, -1, errors.New("no path found")
	}

	// 写入缓存
	cache.Set(origin, destination, shortestPath)

	return shortestPath, length, nil
}

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

	// Reverse the path to get the correct order
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path, float64(len(path) - 1), nil
}
