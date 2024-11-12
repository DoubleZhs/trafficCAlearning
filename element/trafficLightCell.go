package element

type TrafficLightCell struct {
	CommonCell

	// 一个周期完成一轮红绿灯切换，truePhaseInterval规定计数器属于该范围内时相位为true
	phase             bool
	truePhaseInterval [2]int
	interval          int
	count             int
}

func NewTrafficLightCell(id int64, speed int, capacity float64, interval int, truePhaseInterval [2]int) *TrafficLightCell {

	return &TrafficLightCell{
		CommonCell:        *NewCommonCell(id, speed, capacity),
		truePhaseInterval: truePhaseInterval,
		interval:          interval,
		phase:             true,
		count:             0,
	}
}

func (light *TrafficLightCell) Interval() (int, [2]int) {
	return light.interval, light.truePhaseInterval
}

func (light *TrafficLightCell) SetInterval(interval int, truePhaseInterval [2]int) {
	light.interval = interval
	light.truePhaseInterval = truePhaseInterval
}

func (light *TrafficLightCell) Cycle() {
	light.count++
	if light.count > light.interval {
		light.count = 1
	}
	if light.count > light.truePhaseInterval[0] && light.count <= light.truePhaseInterval[1] {
		light.phase = true
	} else {
		light.phase = false
	}
}

func (light *TrafficLightCell) Loadable(vehicle *Vehicle) bool {
	light.containerMux.Lock()
	defer light.containerMux.Unlock()

	return light.phase && (light.occupation+vehicle.occupy <= light.capacity)
}

func (light *TrafficLightCell) ChangeInterval(mul float64) {
	light.interval = int(float64(light.interval) * mul)
	newTruePhase := [2]int{int(float64(light.truePhaseInterval[0]) * mul), int(float64(light.truePhaseInterval[1]) * mul)}
	light.truePhaseInterval = newTruePhase
	light.count = int(float64(light.count) * mul)
}
