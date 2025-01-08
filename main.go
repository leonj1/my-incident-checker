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
)

const (
	notificationEndpoint = "https://ntfy.sh/dapidi_alerts"
	incidentsEndpoint    = "https://status-api.joseserver.com/incidents/recent?count=10"
	pollInterval         = 60 * time.Second
	timeFormat           = "2006-01-02T15:04:05.999999"
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
	log.Printf("Fetching incidents\n")
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

	log.Printf("Successfully fetched incidents: %d\n", len(incidents))
	return incidents, nil
}

func isRelevantState(state string) bool {
	return state == "outage" || state == "degraded"
}

func pollIncidents(startTime time.Time) {
	// Track notified incidents by ID
	notifiedIncidents := make(map[int]bool)

	for {
		incidents, err := fetchIncidents()
		if err != nil {
			log.Printf("Error polling incidents: %v", err)
			time.Sleep(pollInterval)
			continue
		}

		log.Printf("Processing %d incidents\n", len(incidents))
		for _, incident := range incidents {
			// Skip if we've already notified about this incident
			if notifiedIncidents[incident.ID] {
				log.Printf("Skipping already notified incident: %d\n", incident.ID)
				continue
			}

			createdAt, err := time.Parse(timeFormat, strings.Split(incident.CreatedAt, ".")[0])
			if err != nil {
				log.Printf("Error parsing incident time: %v", err)
				continue
			}

			if createdAt.After(startTime) && isRelevantState(incident.CurrentState) {
				message := fmt.Sprintf("New incident for %s service: %s\nState: %s\nDescription: %s\nURL: %s",
					incident.Service,
					incident.Incident.Title,
					incident.CurrentState,
					incident.Incident.Description,
					incident.Incident.URL)

				if err := notify(message); err != nil {
					log.Printf("Failed to send notification: %v", err)
					continue
				}

				// Mark incident as notified
				notifiedIncidents[incident.ID] = true
			}
		}

		time.Sleep(pollInterval)
	}
}

func main() {
	nodeName := getNodeName()
	startupMessage := fmt.Sprintf("%s is online", nodeName)

	if err := notify(startupMessage); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Startup notification sent successfully")

	startTime := time.Now()
	fmt.Printf("Starting incident polling at %s\n", startTime.Format(time.RFC3339))
	pollIncidents(startTime)
}
