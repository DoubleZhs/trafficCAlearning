package utils

import (
	"sync"

	"gonum.org/v1/gonum/graph"
)

type pathCache struct {
	cache      map[[2]graph.Node][]graph.Node
	cacheMutex sync.RWMutex
}

func newPathCache() *pathCache {
	return &pathCache{
		cache: make(map[[2]graph.Node][]graph.Node),
	}
}

func (c *pathCache) Get(from, to graph.Node) ([]graph.Node, bool) {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()
	if nodes, ok := c.cache[[2]graph.Node{from, to}]; ok {
		return nodes, true
	}
	return nil, false
}

func (c *pathCache) Set(from, to graph.Node, nodes []graph.Node) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.cache[[2]graph.Node{from, to}] = nodes
}
