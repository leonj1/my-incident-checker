package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	notificationEndpoint = "https://ntfy.sh/dapidi_alerts"
	incidentsEndpoint    = "https://status-api.joseserver.com/incidents/recent?count=10"
	heartbeatEndpoint    = "https://nosnch.in/2b7bdbea9e"
	pollInterval         = 5 * time.Second
	heartbeatInterval    = 5 * time.Minute
	timeFormat           = "2006-01-02T15:04:05.999999"

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

func notify(message string) error {
	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}
	payload := strings.NewReader(message)
	resp, err := http.Post(notificationEndpoint, "text/plain", payload)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func fetchIncidents() ([]Incident, error) {
	resp, err := http.Get(incidentsEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch incidents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from incidents API: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var incidents []Incident
	if err := json.Unmarshal(body, &incidents); err != nil {
		return nil, fmt.Errorf("failed to parse incidents: %w", err)
	}

	return incidents, nil
}

func isRelevantState(state string) bool {
	return state == stateCritical || state == stateOutage || state == stateDegraded
}

func sendCommand(port *serial.Port, cmd byte) error {
	_, err := port.Write([]byte{cmd})
	return err
}

func pollIncidents(startTime time.Time, light *Light, logger *Logger) {
	logger.infoLog.Printf("Starting incident polling at %s", startTime.Format(time.RFC3339))
	fmt.Printf("*** Starting incident polling at %s\n", startTime.Format(time.RFC3339))

	port, err := serial.OpenPort(&serial.Config{
		Name: serialPort,
		Baud: baudRate,
	})
	if err != nil {
		logger.errorLog.Printf("Failed to open serial port: %s", err.Error())
		light.On(YELLOW_ON)
		return
	}
	defer port.Close()

	notifiedIncidents := make(map[int]bool)
	seenIncidents := make(map[int]bool)

	for {
		if err := checkConnectivity(); err != nil {
			logger.warnLog.Printf("Internet connectivity issue: %s", err.Error())
			light.On(YELLOW_ON)
			time.Sleep(pollInterval)
			continue
		}

		incidents, err := fetchIncidents()
		if err != nil {
			logger.errorLog.Printf("Failed to fetch incidents: %s", err.Error())
			time.Sleep(pollInterval)
			continue
		}

		logger.debugLog.Printf("Fetched %d incidents", len(incidents))

		for _, incident := range incidents {
			if !seenIncidents[incident.ID] {
				logger.infoLog.Printf("New incident detected: [%s] %s - Current State: %s",
					incident.Service,
					incident.Incident.Title,
					incident.CurrentState)
				seenIncidents[incident.ID] = true
			}
		}

		// Log state changes
		state, err := AlertLogic(incidents, light, notifiedIncidents, startTime)
		if err != nil {
			logger.errorLog.Printf("Alert logic error: %s", err.Error())
		} else if state != nil {
			logger.infoLog.Printf("Light state changed to: %T", state)
		}

		time.Sleep(pollInterval)
	}
}

func sortIncidentsByTime(incidents []Incident) []Incident {
	sorted := make([]Incident, len(incidents))
	copy(sorted, incidents)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].CreatedAt < sorted[j].CreatedAt {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

func AlertLogic(incidents []Incident, light *Light, notifiedIncidents map[int]bool, startTime time.Time) (LightState, error) {
	if len(incidents) == 0 {
		return nil, nil
	}

	// Sort incidents by creation time, most recent first
	sortedIncidents := sortIncidentsByTime(incidents)
	mostRecent := sortedIncidents[0]

	// First check the most recent incident
	createdAt, err := parseIncidentTime(mostRecent)
	if err != nil {
		return YellowLight{}, fmt.Errorf("error parsing incident time: %s", err.Error())
	}

	if createdAt.After(startTime) && isNormalState(mostRecent.CurrentState) {
		fmt.Printf("Notification not sent for incident: %s [%s]\n", mostRecent.Incident.Title, mostRecent.CurrentState)
		return GreenLight{}, nil
	}

	// Then check for any unnotified critical incidents
	for _, incident := range sortedIncidents {
		createdAt, err := parseIncidentTime(incident)
		if err != nil {
			return YellowLight{}, fmt.Errorf("error parsing incident time: %s", err.Error())
		}

		if !createdAt.After(startTime) {
			continue
		}

		if !notifiedIncidents[incident.ID] && isRelevantState(incident.CurrentState) {
			notifiedIncidents[incident.ID] = true
			return RedLight{}, nil
		}
	}

	return nil, nil
}

func parseIncidentTime(incident Incident) (time.Time, error) {
	return time.Parse(timeFormat, strings.Split(incident.CreatedAt, ".")[0])
}

func isNormalState(state string) bool {
	return state == stateOperational || state == stateMaintenance
}

func sendHeartbeat() error {
	payload := strings.NewReader("m=just checking in")
	resp, err := http.Post(heartbeatEndpoint, "application/x-www-form-urlencoded", payload)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat:: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("heartbeat failed with status code: %d", resp.StatusCode)
	}

	return nil
}

func runHeartbeat() {
	fmt.Printf("In runHeartbeat\n")
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	fmt.Printf("Starting loop\n")
	for {
		fmt.Printf("Sending heartbeat\n")
		if err := sendHeartbeat(); err != nil {
			fmt.Printf("Heartbeat error: %s\n", err.Error())
			log.Printf("Heartbeat error:: %s", err.Error())
		} else {
			fmt.Printf("Heartbeat sent successfully.\n")
			log.Printf("Heartbeat sent successfully..")
		}
		fmt.Printf("Finshed sending heartbeat\n")
		<-ticker.C
	}
	fmt.Printf("Exiting runHeartbeat\n")
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
