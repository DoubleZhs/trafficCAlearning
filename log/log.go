package log

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

func LogEnvironment(numClosedVehicle int) {
	log.Println("----------- Simulation Sets ------------")
	log.Printf("Operating System: %s\n", runtime.GOOS)
	log.Printf("Architecture: %s\n", runtime.GOARCH)
	log.Printf("Number of CPUs: %d\n", runtime.NumCPU())
	log.Printf("Compiler Version: %s\n", runtime.Version())
	log.Println("Number of Closed Vehicles:", numClosedVehicle)
	log.Println("----------------------------------------")
}

func InitLog() {
	currentTime := time.Now().Format("2006010215040506")
	logFilename := fmt.Sprintf("%s.log", currentTime)
	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %s", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
}
