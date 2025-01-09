package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tarm/serial"
)

const (
	notificationEndpoint = "https://ntfy.sh/dapidi_alerts"
	incidentsEndpoint    = "https://status-api.joseserver.com/incidents/recent?count=10"
	connectivityCheck    = "https://www.google.com"
	heartbeatEndpoint    = "https://nosnch.in/2b7bdbea9e"
	pollInterval         = 5 * time.Second
	heartbeatInterval    = 5 * time.Minute
	timeFormat           = "2006-01-02T15:04:05.999999"
	connectTimeout       = 10 * time.Second

	serialPort = "/dev/ttyUSB0" // Change to the serial/COM port of the tower light
	baudRate   = 9600

	// Command bytes for LEDs and buzzer
	RED_ON    = 0x11
	RED_OFF   = 0x21
	RED_BLINK = 0x41

	YELLOW_ON    = 0x12
	YELLOW_OFF   = 0x22
	YELLOW_BLINK = 0x42

	GREEN_ON    = 0x14
	GREEN_OFF   = 0x24
	GREEN_BLINK = 0x44

	BUZZER_ON    = 0x18
	BUZZER_OFF   = 0x28
	BUZZER_BLINK = 0x48
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

type Light struct {
}

func (l *Light) On(onCmd byte) error {
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
			log.Printf("Error closing serial port: %v", err)
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
			log.Printf("Error closing serial port: %v", err)
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

func checkConnectivity() error {
	client := &http.Client{
		Timeout: connectTimeout,
	}

	resp, err := client.Get(connectivityCheck)
	if err != nil {
		return fmt.Errorf("connectivity check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connectivity check failed with status code: %d", resp.StatusCode)
	}

	return nil
}

func getNodeName() string {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName = os.Getenv("HOSTNAME")
	}
	if nodeName == "" {
		nodeName = "unknown"
	}
	return nodeName
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
	return state == "outage" || state == "degraded"
}

func sendCommand(port *serial.Port, cmd byte) error {
	_, err := port.Write([]byte{cmd})
	return err
}

func pollIncidents(startTime time.Time, light *Light) {
	// Configure serial port
	c := &serial.Config{
		Name: serialPort,
		Baud: baudRate,
	}

	// Open the serial port
	port, err := serial.OpenPort(c)
	if err != nil {
		log.Printf("Failed to open serial port: %v", err)
		light.On(YELLOW_ON)
		return
	}
	defer port.Close()

	// Track notified incidents by ID
	notifiedIncidents := make(map[int]bool)

	for {
		// Check connectivity before polling
		if err := checkConnectivity(); err != nil {
			log.Printf("Internet connectivity issue: %v", err)
			light.On(YELLOW_ON)
			time.Sleep(pollInterval)
			continue
		}

		incidents, err := fetchIncidents()
		if err != nil {
			log.Printf("Error polling incidents: %v", err)
			light.On(YELLOW_ON)
			time.Sleep(pollInterval)
			continue
		}

		var mostRecentIncident *Incident = nil
		var notificationSent bool = false
		for _, incident := range incidents {
			if incident.CreatedAt > mostRecentIncident.CreatedAt {
				mostRecentIncident = &incident
			}

			// Skip if we've already notified about this incident
			if notifiedIncidents[incident.ID] {
				continue
			}

			createdAt, err := time.Parse(timeFormat, strings.Split(incident.CreatedAt, ".")[0])
			if err != nil {
				log.Printf("Error parsing incident time: %v", err)
				light.On(YELLOW_ON)
				continue
			}

			if createdAt.After(startTime) && isRelevantState(incident.CurrentState) {
				message := fmt.Sprintf("New incident for %s service: %s\nState: %s\nDescription: %s\nURL: %s",
					incident.Service,
					incident.Incident.Title,
					incident.CurrentState,
					incident.Incident.Description,
					incident.Incident.URL)

				notificationSent = true
				light.On(RED_ON)
				if err := notify(message); err != nil {
					log.Printf("Failed to send notification: %v", err)
					continue
				}

				// Mark incident as notified
				notifiedIncidents[incident.ID] = true
			}
		}

		// Update tower light based on most recent incident state
		if mostRecentIncident != nil && !notificationSent {
			// if mostRecent.CurrentState is "operational" or "maintenance", skip
			if mostRecentIncident.CurrentState == "operational" || mostRecentIncident.CurrentState == "maintenance" {
				light.On(GREEN_ON)
			}
		}

		time.Sleep(pollInterval)
	}
}

func sendHeartbeat() error {
	payload := strings.NewReader("m=just checking in")
	resp, err := http.Post(heartbeatEndpoint, "application/x-www-form-urlencoded", payload)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("heartbeat failed with status code: %d", resp.StatusCode)
	}

	return nil
}

func runHeartbeat() {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		if err := sendHeartbeat(); err != nil {
			log.Printf("Heartbeat error: %v", err)
		} else {
			log.Printf("Heartbeat sent successfully")
		}
		<-ticker.C
	}
}

func main() {
	// Check initial connectivity
	if err := checkConnectivity(); err != nil {
		log.Printf("Initial connectivity check failed: %v", err)
		log.Printf("Will continue and retry during polling...")
	} else {
		log.Println("Internet connectivity confirmed")
	}

	nodeName := getNodeName()
	startupMessage := fmt.Sprintf("%s is online", nodeName)

	if err := notify(startupMessage); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Startup notification sent successfully")

	light := Light{}

	// Start heartbeat in a goroutine
	go runHeartbeat()

	startTime := time.Now()
	fmt.Printf("Starting incident polling at %s\n", startTime.Format(time.RFC3339))
	pollIncidents(startTime, &light)
}
