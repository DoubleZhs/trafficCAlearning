package element

type TrafficLightCell struct {
	CommonCell

	phase    bool
	interval [2]int // index 0: true, index 1: false
	count    int
}

func NewTrafficLightCell(id int64, speed int, capacity float64, initPhase bool, interval [2]int, initCount int) *TrafficLightCell {
	return &TrafficLightCell{
		CommonCell: *NewCommonCell(id, speed, capacity),
		phase:      initPhase,
		interval:   interval,
		count:      initCount,
	}
}

func (light *TrafficLightCell) changePhase() {
	light.phase = !light.phase
}

func (light *TrafficLightCell) Cycle() {
	light.count++
	if light.phase && light.count == light.interval[0] {
		light.changePhase()
		light.count = 1
	} else if !light.phase && light.count == light.interval[1] {
		light.changePhase()
		light.count = 1
	}
}

func (light *TrafficLightCell) Loadable(vehicle *Vehicle) bool {
	light.containerMux.Lock()
	defer light.containerMux.Unlock()

	return light.phase && (light.occupation+vehicle.occupy <= light.capacity)
}
