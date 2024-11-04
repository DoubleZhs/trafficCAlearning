package recorder

import (
	"graphCA/element"
	"strconv"
	"sync"
)

var (
	traceDataCache [][][]string = make([][][]string, 0)
	traceDataMutex sync.Mutex   = sync.Mutex{}
)

func RecordTraceData(vehicle *element.Vehicle) {
	traceDataMutex.Lock()
	defer traceDataMutex.Unlock()
	traceDataCache = append(traceDataCache, getTraceData(vehicle))
}

func getTraceData(vehicle *element.Vehicle) [][]string {
	id := vehicle.Index()
	trace := vehicle.Trace()

	var records [][]string

	for nodeID, arrivalTime := range trace {
		record := []string{
			strconv.FormatInt(id, 10),     // 车辆 ID
			strconv.FormatInt(nodeID, 10), // 节点 ID
			strconv.Itoa(arrivalTime),     // 抵达时间
		}
		records = append(records, record)
	}

	return records
}

func InitTraceDataCSV(filename string) {
	header := []string{
		"VehicleID", "NodeID", "ArrivalTime",
	}
	initializeCSV(filename, header)
}

func WriteToTraceDataCSV(filename string) {
	if len(traceDataCache) == 0 {
		return
	}
	for _, records := range traceDataCache {
		appendToCSV(filename, records)
	}
}
