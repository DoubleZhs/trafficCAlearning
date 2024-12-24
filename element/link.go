package element

import (
	"sync/atomic"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

var cellIndex int64 = 10000000000

func getNextCellID() int64 {
	return atomic.AddInt64(&cellIndex, 1)
}

type Link struct {
	id              int64
	cells           []graph.Node
	from, to        graph.Node
	numCells        int
	speedLimit      int
	capacity        float64
	design_capacity float64
}

func NewLink(id int64, from, to graph.Node, numCells, speed int, capacity, design_capacity float64) *Link {
	if numCells < 2 {
		panic("numCells must be at least 2")
	}
	cells := make([]graph.Node, numCells)
	for i := 0; i < numCells; i++ {
		if i < numCells-1 {
			cells[i] = NewCommonCell(getNextCellID(), speed, float64(capacity))
		} else {
			cells[i] = NewTrafficLightCell(getNextCellID(), speed, float64(capacity), 0, [2]int{0, 1})
		}
	}
	return &Link{
		id:              id,
		cells:           cells,
		from:            from,
		to:              to,
		numCells:        numCells,
		speedLimit:      speed,
		capacity:        capacity,
		design_capacity: design_capacity,
	}
}

func (l *Link) ID() int64 {
	return l.id
}

func (l *Link) Flat() []graph.Node {
	return l.cells
}

func (l *Link) From() graph.Node {
	return l.from
}

func (l *Link) To() graph.Node {
	return l.to
}

func (l *Link) ReversedEdge() graph.Edge {
	reversedCells := make([]graph.Node, len(l.cells))
	for i, j := 0, len(l.cells)-1; i <= j; i, j = i+1, j-1 {
		reversedCells[i], reversedCells[j] = l.cells[j], l.cells[i]
	}
	return &Link{
		id:              l.id,
		cells:           reversedCells,
		from:            l.to,
		to:              l.from,
		numCells:        l.numCells,
		speedLimit:      l.speedLimit,
		capacity:        l.capacity,
		design_capacity: l.design_capacity,
	}
}

func (l *Link) Weight() float64 {
	averageSpeed := l.AverageSpeed()
	density := l.Density()
	// eTime := (float64(l.numCells) / float64(l.speedLimit)) * (1 + 0.15*math.Pow((averageSpeed*density/l.design_capacity), 4.0))
	if averageSpeed == 0 {
		if density <= 0.7 {
			averageSpeed = float64(l.speedLimit)
		} else {
			averageSpeed = 0.00001
		}

	}
	eTime := float64(l.numCells) / averageSpeed
	return eTime
}

func (l *Link) AddToSimGraph(g *simple.DirectedGraph) {
	for _, cell := range l.cells {
		g.AddNode(cell)
	}
	for i := 0; i < len(l.cells)-1; i++ {
		g.SetEdge(simple.Edge{F: l.cells[i], T: l.cells[i+1]})
	}
	g.SetEdge(simple.Edge{F: l.from, T: l.cells[0]})
	g.SetEdge(simple.Edge{F: l.cells[len(l.cells)-1], T: l.to})
}

func (l *Link) Length() int {
	return l.numCells
}

func (l *Link) Vehicles() []*Vehicle {
	allVehicle := make([]*Vehicle, 0)
	for _, cell := range l.cells {
		c := cell.(Cell)
		vehicles := c.ListContainer()
		allVehicle = append(allVehicle, vehicles...)
	}
	return allVehicle
}

func (l *Link) NumVehicles() int {
	allVehicle := l.Vehicles()
	return len(allVehicle)
}

func (l *Link) Density() float64 {
	numVehicle := l.NumVehicles()
	return float64(numVehicle) / (float64(l.numCells) * l.capacity)
}

func (l *Link) AverageSpeed() float64 {
	allVehicle := l.Vehicles()
	var totalSpeed float64 = 0
	for _, v := range allVehicle {
		totalSpeed += float64(v.velocity)
	}
	numVehicle := len(allVehicle)
	var averageSpeed float64 = 0
	if numVehicle == 0 {
		averageSpeed = 0
	} else {
		averageSpeed = totalSpeed / float64(numVehicle)
	}
	return averageSpeed
}

func (l *Link) Report() (int, int, float64, int, float64) {
	numVehicle := l.NumVehicles()
	averageSpeed := l.AverageSpeed()
	return l.numCells, l.speedLimit, l.capacity, numVehicle, averageSpeed
}
