package utils

import (
	"errors"

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
