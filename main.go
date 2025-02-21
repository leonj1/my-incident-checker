package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"my-incident-checker/heartbeat"
	"my-incident-checker/lights"
	"my-incident-checker/network"
	"my-incident-checker/node"
	"my-incident-checker/notify"
	"my-incident-checker/poll"
	"my-incident-checker/types"
)

func NewLogger() (*types.Logger, error) {
	// Create logs directory if it doesn't exist
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	// Create or append to log file with timestamp
	currentTime := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, "incident-checker-"+currentTime+".log")
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// Create multi-writer for both file and stdout
	flags := log.LstdFlags | log.Lmicroseconds | log.LUTC

	return &types.Logger{
		DebugLog: log.New(file, "DEBUG: ", flags),
		InfoLog:  log.New(file, "INFO:  ", flags),
		WarnLog:  log.New(file, "WARN:  ", flags),
		ErrorLog: log.New(file, "ERROR: ", flags),
	}, nil
}

func main() {
	logger, err := NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %s", err.Error())
	}

	// Add deferred exit log message
	defer logger.InfoLog.Printf("Program shutting down")

	logger.InfoLog.Printf("Starting Incident Checker")

	// Check initial connectivity
	if err := network.CheckConnectivity(); err != nil {
		logger.WarnLog.Printf("Initial connectivity check failed: %s", err.Error())
		logger.InfoLog.Printf("Will continue and retry during polling...")
	} else {
		logger.InfoLog.Printf("Internet connectivity confirmed")
	}

	nodeName := node.GetNodeName()
	startupMessage := fmt.Sprintf("%s is online", nodeName)

	if err := notify.Send(startupMessage); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Startup notification sent successfully")

	// Initialize the light with automatic detection
	light, cleanup, err := initializeLight(logger)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	fmt.Println("Yellow light on for 2 seconds")
	light.Blink(lights.StateYellow)
	time.Sleep(2 * time.Second)

	fmt.Println("Red light on for 2 seconds")
	light.Blink(lights.StateRed)
	time.Sleep(2 * time.Second)

	fmt.Println("Green light on for 2 seconds")
	light.Blink(lights.StateGreen)
	time.Sleep(2 * time.Second)

	light.Clear()
	fmt.Println("Lights cleared")
	light.On(lights.StateGreen)

	// Start heartbeat in a goroutine
	fmt.Println("Starting heartbeat")
	logger.InfoLog.Printf("Starting heartbeat...")
	go heartbeat.Run()

	fmt.Println("Polling for incidents")
	startTime := time.Now()
	poll.PollIncidents(startTime, light, logger)
	fmt.Println("Stopped polling for incidents")
}

func initializeLight(logger *types.Logger) (lights.Light, func(), error) {
	var light lights.Light
	var cleanup func()

	// Try to initialize BLINK1MK3 first
	if blink1Light, err := lights.NewBlink1Light(); err == nil {
		fmt.Println("Using BLINK1MK3 light")
		logger.InfoLog.Printf("Using BLINK1MK3 light")
		light = blink1Light
		cleanup = func() {
			blink1Light.Close()
		}
	} else {
		// Fall back to SerialLight
		fmt.Println("BLINK1MK3 not found, using SerialLight")
		logger.InfoLog.Printf("BLINK1MK3 not found, using SerialLight")
		light = lights.NewSerialLight("/dev/ttyUSB0", 9600)
		cleanup = func() {}
	}

	// Initialize to green state
	initialState := lights.GreenState{}
	if err := initialState.Apply(light); err != nil {
		return nil, nil, fmt.Errorf("failed to set initial light state: %w", err)
	}
	logger.InfoLog.Printf("Light initialized to green")

	return light, cleanup, nil
}
