package main

import (
	"fmt"
	"graphCA/element"
	"graphCA/log"
	"graphCA/recorder"
	"graphCA/simulator"
	"graphCA/utils"
	"runtime"
	"sync"
	"time"

	"gonum.org/v1/gonum/graph"
)

const (
	oneDayTimeSteps int = 57600

	// 日志间隔
	intervalWriteToLog int = 1200
	// 写入linkData间隔，间隔内均值
	intervalWriteLinkData int = 400

	intervalWriteOtherData int = 4800

	// 需求分布仿射变换参数
	Multiplier float64 = 1.5e5
	FixedNum   float64 = 0.25e5 / 57600
	// 时间步需求扰动范围[1-RandomDisRange, 1+RandomDisRange]
	RandomDisRange float64 = 0.25
	// 固定车辆数
	numClosedVehicle int = 50

	simDay                        int     = 2
	initTrafficLightPhaseInterval int     = 20
	trafficLightChangeDay         int     = 2
	traffcLightMul                float64 = 3.0
	inDegreeThreshold             int     = 4

	rawNodeIds int64 = 416
)

// 并发数
var numWorkers int = runtime.GOMAXPROCS(0)

func main() {
	initTime := time.Now().Format("2006010215040506")

	// 日志
	logFile := fmt.Sprintf("./log/log_%s_%d.log", initTime, numClosedVehicle)
	log.InitLog(logFile)
	log.LogEnvironment()
	log.LogSimParameters(oneDayTimeSteps, Multiplier, FixedNum, RandomDisRange, numClosedVehicle, simDay, initTrafficLightPhaseInterval, trafficLightChangeDay, traffcLightMul)
	log.WriteLog(fmt.Sprintf("Concurrent Volume in Vehicle Process: %d", numWorkers))
	defer log.CloseLog()

	// 数据CSV
	systemDataFile := fmt.Sprintf("./data/1_SystemData_%s_%d.csv", initTime, numClosedVehicle)
	linkDataFile := fmt.Sprintf("./data/2_LinkData_%s_%d.csv", initTime, numClosedVehicle)
	vehicleDataFile := fmt.Sprintf("./data/3_VehicleData_%s_%d.csv", initTime, numClosedVehicle)
	traceDataFile := fmt.Sprintf("./data/4_TraceData_%s_%d.csv", initTime, numClosedVehicle)
	recorder.InitSystemDataCSV(systemDataFile)
	recorder.InitLinkDataCSV(linkDataFile)
	recorder.InitVehicleDataCSV(vehicleDataFile)
	recorder.InitTraceDataCSV(traceDataFile)

	// 仿真图
	simpleG, simpleNodes := simulator.CreateAnaheimSimpleGraph()
	simulationG, simulationNodes, lightGroups := simulator.CreateAnaheimSimulationGraph(simpleG, simpleNodes, inDegreeThreshold, initTrafficLightPhaseInterval)
	numNodes := len(simulationNodes)
	log.WriteLog(fmt.Sprintf("Number of Nodes: %d", numNodes))
	log.WriteLog(fmt.Sprintf("Number of TrafficLight Group: %d", len(lightGroups)))
	// for i, nodes := range lightGroups {
	// 	log.WriteLog(fmt.Sprintf("Group %d: %d", i, len(nodes)))
	// 	for _, node := range nodes {
	// 		inter, trueinter := node.Interval()
	// 		log.WriteLog(fmt.Sprintf("Node %d: Interval: %d, TrueInterval: %d.", node.ID(), inter, trueinter))
	// 	}
	// }

	// 检查连通性
	simpleConnect, simulationConnect := utils.IsStronglyConnected(simpleG), utils.IsStronglyConnected(simulationG)
	log.WriteLog(fmt.Sprintf("SimpleGraph Connected: %v, SimulationGraph Connected: %v", simpleConnect, simulationConnect))

	// keyNodes - 输出的车辆路径中仅记录这些节点
	keyNodes := make(map[int64]graph.Node)
	// linkNodes - 用一个节点在simpleG中表示一个link（计算路径）
	linkNodes := make(map[int64]graph.Node)
	for i, node := range simpleNodes {
		if i <= rawNodeIds {
			keyNodes[i] = node
		} else {
			linkNodes[i] = node
		}
	}

	// graph.Node断言为Link存储
	links := make([]*element.Link, 0)
	for _, node := range linkNodes {
		assertlink, ok := node.(*element.Link)
		if !ok {
			fmt.Println("Link ID:", assertlink.ID())
			panic("Node is not a link")
		}
		links = append(links, assertlink)
	}

	// 允许作为起终点的节点
	var allowedOrigin, allowedDestination []graph.Node
	for i, node := range simpleNodes {
		if i <= rawNodeIds {
			allowedOrigin = append(allowedOrigin, node)
			allowedDestination = append(allowedDestination, node)
		}
	}

	// 初始化系统
	var demand []float64
	simulator.InitFixedVehicle(numClosedVehicle, simpleG, simulationG, allowedOrigin, allowedDestination)
	// 仿真主进程
	log.WriteLog("----------------------------------Simulation Start----------------------------------")
	for timeStep := 0; timeStep < simDay*oneDayTimeSteps; timeStep++ {
		// timeStep换算为日期和当日时间
		timeOfDay := timeStep % oneDayTimeSteps
		currentDay := timeStep/oneDayTimeSteps + 1

		// 每天需求分布随机扰动
		if timeOfDay == 0 {
			demand = simulator.AdjustDemand(Multiplier, FixedNum)
		}

		// 红绿灯周期改变
		if currentDay == trafficLightChangeDay && timeOfDay == 0 {
			for _, nodes := range lightGroups {
				for _, node := range nodes {
					node.ChangeInterval(traffcLightMul)
				}
			}
			log.WriteLog(fmt.Sprintf("TrafficLight Interval Changed: Multiplier - %.2f", traffcLightMul))
		}

		// 生成车辆
		generateNum := simulator.GetGenerateVehicleCount(timeOfDay, demand, RandomDisRange)
		simulator.GenerateScheduleVehicle(timeStep, generateNum, simpleG, simulationG, allowedOrigin, allowedDestination)

		// 红绿灯循环
		for _, nodes := range lightGroups {
			for _, node := range nodes {
				node.Cycle()
			}
		}

		// 处理车辆
		simulator.VehicleProcess(numWorkers, timeStep, simpleG, simulationG, allowedDestination)

		// 记录系统状态及链路状态至缓存
		recorder.RecordLinkData(timeStep, links)
		numVehicleGenerated, numVehiclesActive, numVehiclesWaiting, numVehicleCompleted := simulator.GetVehiclesNum()
		vehiclesOnRoad := simulator.GetVehiclesOnRoad()
		averageSpeed, density := simulator.GetAverageSpeed_Density(vehiclesOnRoad, numNodes)
		recorder.RecordSystemData(timeStep, numVehicleGenerated, numVehiclesActive, numVehiclesWaiting, numVehicleCompleted, averageSpeed, density)

		// 按时间间隔输出日志
		if timeOfDay%intervalWriteToLog == 0 {
			numVehicleGenerated, numVehiclesActive, numVehiclesWaiting, numVehicleCompleted := simulator.GetVehiclesNum()
			log.WriteLog(fmt.Sprintf("Day: %d, TimeOfDay: %v, Generated: %d, Active: %d, OnRoad: %d, Waiting: %d, Completed: %d", currentDay, log.ConvertTimeStepToTime(timeOfDay), numVehicleGenerated, numVehiclesActive, len(vehiclesOnRoad), numVehiclesWaiting, numVehicleCompleted))
		}

		// 按时间间隔写入CSV并清空缓存
		var wg sync.WaitGroup
		if timeOfDay%intervalWriteLinkData == 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				recorder.WriteToLinkDataCSV(linkDataFile, links)
			}()
		}
		if timeOfDay%intervalWriteOtherData == 0 {
			wg.Add(3)
			go func() {
				defer wg.Done()
				recorder.WriteToSystemDataCSV(systemDataFile)
			}()
			go func() {
				defer wg.Done()
				recorder.WriteToTraceDataCSV(traceDataFile)
			}()
			go func() {
				defer wg.Done()
				recorder.WriteToVehicleDataCSV(vehicleDataFile)
			}()
		}
		wg.Wait()
	}

	// 主循环结束检查缓存并写入CSV
	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		recorder.WriteToLinkDataCSV(linkDataFile, links)
	}()
	go func() {
		defer wg.Done()
		recorder.WriteToSystemDataCSV(systemDataFile)
	}()
	go func() {
		defer wg.Done()
		recorder.WriteToTraceDataCSV(traceDataFile)
	}()
	go func() {
		defer wg.Done()
		recorder.WriteToVehicleDataCSV(vehicleDataFile)
	}()
	wg.Wait()

	log.WriteLog("---------------------------------- Completed ----------------------------------")
}
