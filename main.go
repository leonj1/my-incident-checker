package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/tarm/serial"
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

	serialPort = "/dev/ttyUSB0" // Change to the serial/COM port of the tower light
	baudRate   = 9600

	stateOperational = "operational"
	stateMaintenance = "maintenance"
	stateCritical    = "critical"
	stateOutage      = "outage"
	stateDegraded    = "degraded"

	// Command bytes for LEDs and buzzer
	RED_ON    byte = 0x11
	RED_OFF   byte = 0x21
	RED_BLINK byte = 0x41

	YELLOW_ON    byte = 0x12
	YELLOW_OFF   byte = 0x22
	YELLOW_BLINK byte = 0x42

	GREEN_ON    byte = 0x14
	GREEN_OFF   byte = 0x24
	GREEN_BLINK byte = 0x44

	BUZZER_ON    byte = 0x18
	BUZZER_OFF   byte = 0x28
	BUZZER_BLINK byte = 0x48
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

type Light struct{}

type LightState interface {
	Apply(light *Light) error
}

type RedLight struct{}
type GreenLight struct{}
type YellowLight struct{}

func (s RedLight) Apply(light *Light) error {
	return light.On(RED_ON)
}

func (s GreenLight) Apply(light *Light) error {
	return light.On(GREEN_ON)
}

func (s YellowLight) Apply(light *Light) error {
	return light.On(YELLOW_ON)
}

func (l *Light) On(onCmd byte) error {
	l.Clear()

	// Configure serial port settings
	c := &serial.Config{
		Name: serialPort,
		Baud: baudRate,
	}

	// Open the serial port
	s, err := serial.OpenPort(c)
	if err != nil {
		return err
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Printf("Error closing serial port: %s", err.Error())
		}
	}()

	if err := sendCommand(s, onCmd); err != nil {
		return err
	}
	return nil
}

func (l *Light) Clear() error {
	// Configure serial port settings
	c := &serial.Config{
		Name: serialPort,
		Baud: baudRate,
	}

	// Open the serial port
	s, err := serial.OpenPort(c)
	if err != nil {
		return err
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Printf("Error closing serial port: %s", err.Error())
		}
	}()

	// Clean up any old state by turning off all LEDs and buzzer
	initialCommands := []byte{
		BUZZER_OFF,
		RED_OFF,
		YELLOW_OFF,
		GREEN_OFF,
	}

	for _, cmd := range initialCommands {
		if err := sendCommand(s, cmd); err != nil {
			return err
		}
	}
	return nil
}

func sendCommand(port *serial.Port, cmd byte) error {
	_, err := port.Write([]byte{cmd})
	return err
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

	light := Light{}
	light.Clear()
	fmt.Println("Yellow light on for 2 seconds")
	light.On(YELLOW_BLINK)
	time.Sleep(2 * time.Second)
	fmt.Println("Red light on for 2 seconds")
	light.On(RED_BLINK)
	time.Sleep(2 * time.Second)
	fmt.Println("Green light on for 2 seconds")
	light.On(GREEN_BLINK)
	time.Sleep(2 * time.Second)
	light.Clear()
	fmt.Println("Lights cleared")
	light.On(GREEN_ON)

	// Start heartbeat in a goroutine
	fmt.Println("Starting heartbeat")
	logger.infoLog.Printf("Starting heartbeat...")
	go runHeartbeat()

	fmt.Println("Polling for incidents")
	startTime := time.Now()
	pollIncidents(startTime, &light, logger)
	fmt.Println("Stopped polling for incidents")
}
