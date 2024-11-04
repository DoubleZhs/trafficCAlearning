package recorder

import (
	"fmt"
	"graphCA/element"
	"strconv"
)

var (
	linkDataCache [][]string = make([][]string, 914)
)

func RecordLinkData(simTime int, links []*element.Link) {
	for _, link := range links {
		linkDataCache = append(linkDataCache, GetLinkData(simTime, link))
	}
}

func GetLinkData(simTime int, link *element.Link) []string {
	timeOfDay := simTime % 57600
	day := simTime/57600 + 1
	numCell, speedLim, capacity, numVehicle, averageSpeed := link.Report()
	density := float64(numVehicle) / (float64(numCell) * capacity)

	return []string{
		strconv.FormatInt(link.ID(), 10),
		strconv.Itoa(day),       // 当前天数
		strconv.Itoa(timeOfDay), // 时间步
		strconv.Itoa(numCell),   // 道路单元数
		strconv.Itoa(speedLim),
		strconv.Itoa(int(capacity)),
		strconv.Itoa(numVehicle),
		fmt.Sprintf("%.4f", averageSpeed),
		fmt.Sprintf("%.4f", density),
	}
}

func InitLinkDataCSV(filename string) {
	header := []string{
		"ID", "Day", "TimeOfDay", "NumCell", "SpeedLim", "Capacity", "NumVehicle", "AverageSpeed", "Density",
	}
	initializeCSV(filename, header)
}

func WriteToLinkDataCSV(filename string, simTime int, links []*element.Link) {
	if len(linkDataCache) == 0 {
		return
	}
	appendToCSV(filename, linkDataCache)
}
