package utils

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/traverse"
)

func Accessible(g graph.Graph, from, to graph.Node) bool {
	dfs := traverse.DepthFirst{}
	found := false

	dfs.Walk(g, from, func(n graph.Node) bool {
		if n.ID() == to.ID() {
			found = true
			return false // Stop the traversal
		}
		return true // Continue the traversal
	})

	return found
}
