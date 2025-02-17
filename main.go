package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"my-incident-checker/lights"
)

// Log levels
const (
	LogDebug = "DEBUG"
	LogInfo  = "INFO"
	LogWarn  = "WARN"
	LogError = "ERROR"
)

type Logger struct {
	debugLog *log.Logger
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
}

func NewLogger() (*Logger, error) {
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

	return &Logger{
		debugLog: log.New(file, "DEBUG: ", flags),
		infoLog:  log.New(file, "INFO:  ", flags),
		warnLog:  log.New(file, "WARN:  ", flags),
		errorLog: log.New(file, "ERROR: ", flags),
	}, nil
}

const (
	timeFormat = "2006-01-02T15:04:05.999999"

	stateOperational = "operational"
	stateMaintenance = "maintenance"
	stateCritical    = "critical"
	stateOutage      = "outage"
	stateDegraded    = "degraded"
)

type IncidentDetails struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Components  []string `json:"components"`
	URL         string   `json:"url"`
}

type IncidentHistory struct {
	ID           int             `json:"id"`
	IncidentID   int             `json:"incident_id"`
	RecordedAt   string          `json:"recorded_at"`
	Service      string          `json:"service"`
	PrevState    string          `json:"previous_state"`
	CurrentState string          `json:"current_state"`
	Incident     IncidentDetails `json:"incident"`
}

type Incident struct {
	ID           int               `json:"id"`
	Service      string            `json:"service"`
	PrevState    string            `json:"previous_state"`
	CurrentState string            `json:"current_state"`
	CreatedAt    string            `json:"created_at"`
	Incident     IncidentDetails   `json:"incident"`
	History      []IncidentHistory `json:"history"`
}

func main() {
	logger, err := NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %s", err.Error())
	}

	// Add deferred exit log message
	defer logger.infoLog.Printf("Program shutting down")

	logger.infoLog.Printf("Starting Incident Checker")

	// Check initial connectivity
	if err := checkConnectivity(); err != nil {
		logger.warnLog.Printf("Initial connectivity check failed: %s", err.Error())
		logger.infoLog.Printf("Will continue and retry during polling...")
	} else {
		logger.infoLog.Printf("Internet connectivity confirmed")
	}

	nodeName := getNodeName()
	startupMessage := fmt.Sprintf("%s is online", nodeName)

	if err := notify(startupMessage); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Startup notification sent successfully")

	// Initialize the light
	light := lights.NewTrafficLight("/dev/ttyUSB0", 9600)
	light.Clear()

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
	logger.infoLog.Printf("Starting heartbeat...")
	go runHeartbeat()

	fmt.Println("Polling for incidents")
	startTime := time.Now()
	pollIncidents(startTime, light, logger)
	fmt.Println("Stopped polling for incidents")
}
