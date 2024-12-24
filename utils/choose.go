package utils

import (
	"graphCA/element"
	"math"
	"math/rand/v2"

	"gonum.org/v1/gonum/graph"
)

func LogitChoose(paths [][]graph.Node, beta float64, linksMap map[[2]int64]*element.Link) []graph.Node {
	costs := make([]float64, len(paths))
	for i, path := range paths {
		costs[i] = pathCost(path, linksMap)
	}

	probs := make([]float64, len(paths))
	total := 0.0
	for _, cost := range costs {
		total += math.Exp(-beta * cost)
	}
	for i, cost := range costs {
		probs[i] = math.Exp(-beta*cost) / total
	}

	cumulate := make([]float64, len(probs))
	for i, prob := range probs {
		if i == 0 {
			cumulate[i] = prob
		} else {
			cumulate[i] = cumulate[i-1] + prob
		}
	}

	randDice := rand.Float64()
	for i, cum := range cumulate {
		if randDice <= cum {
			return paths[i]
		}
	}

	return nil
}

func pathCost(path []graph.Node, linksMap map[[2]int64]*element.Link) float64 {
	totalCost := 0.0
	for i := 0; i < len(path)-1; i++ {
		link := linksMap[[2]int64{path[i].ID(), path[i+1].ID()}]
		if link == nil {
			// Handle the case where the link is nil
			return math.Inf(1) // or any other appropriate error handling
		}
		totalCost += link.Weight()
	}
	return totalCost
}
