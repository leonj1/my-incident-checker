package poll

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"my-incident-checker/lights"
	"my-incident-checker/network"
	"my-incident-checker/types"
)

const (
	incidentsEndpoint = "https://status-api.joseserver.com/incidents/recent?count=10"
	pollInterval      = 5 * time.Second
)

// PollIncidents continuously monitors for incidents and updates the light status
func PollIncidents(startTime time.Time, light lights.Light, logger *types.Logger) {
	logger.InfoLog.Printf("Starting incident polling at %s", startTime.Format(time.RFC3339))
	fmt.Printf("*** Starting incident polling at %s\n", startTime.Format(time.RFC3339))

	notifiedIncidents := make(map[int]bool)
	seenIncidents := make(map[int]bool)

	for {
		if err := network.CheckConnectivity(); err != nil {
			logger.WarnLog.Printf("Internet connectivity issue: %s", err.Error())
			light.On(lights.StateYellow)
			time.Sleep(pollInterval)
			continue
		}

		incidents, err := fetchIncidents()
		if err != nil {
			logger.ErrorLog.Printf("Failed to fetch incidents: %s", err.Error())
			time.Sleep(pollInterval)
			continue
		}

		logger.DebugLog.Printf("Fetched %d incidents", len(incidents))

		for _, incident := range incidents {
			if !seenIncidents[incident.ID] {
				logger.InfoLog.Printf("New incident detected: [%s] %s - Current State: %s",
					incident.Service,
					incident.Incident.Title,
					incident.CurrentState)
				seenIncidents[incident.ID] = true
			}
		}

		// Log state changes
		state, err := AlertLogic(incidents, light, notifiedIncidents, startTime, logger)
		if err != nil {
			logger.ErrorLog.Printf("Alert logic error: %s", err.Error())
		} else if state != nil {
			logger.InfoLog.Printf("Light state changed to: %T", state)
			if err := light.On(state); err != nil {
				logger.ErrorLog.Printf("Failed to apply light state: %s", err.Error())
			}
		}

		time.Sleep(pollInterval)
	}
}

// fetchIncidents retrieves the list of incidents from the API
func fetchIncidents() ([]types.Incident, error) {
	resp, err := http.Get(incidentsEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch incidents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from incidents API: %d", resp.StatusCode)
	}

	const maxResponseSize = 1 << 20 // 1 MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var incidents []types.Incident
	if err := json.Unmarshal(body, &incidents); err != nil {
		return nil, fmt.Errorf("failed to parse incidents: %w", err)
	}

	return incidents, nil
}

// sortIncidentsByTime sorts incidents by creation time, most recent first
func sortIncidentsByTime(incidents []types.Incident) []types.Incident {
	sorted := make([]types.Incident, len(incidents))
	copy(sorted, incidents)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt > sorted[j].CreatedAt
	})

	return sorted
}

// AlertLogic determines the appropriate light state based on incident status
func AlertLogic(incidents []types.Incident, light lights.Light, notifiedIncidents map[int]bool, startTime time.Time, logger *types.Logger) (lights.State, error) {
	if len(incidents) == 0 {
		// Clear notification history and return green state when no incidents
		for k := range notifiedIncidents {
			delete(notifiedIncidents, k)
		}
		logger.InfoLog.Printf("No active incidents, setting light to green")
		return lights.GreenState{}, nil
	}

	// Sort incidents by creation time, most recent first
	sortedIncidents := sortIncidentsByTime(incidents)

	// Check if most recent incident is in normal state
	if len(sortedIncidents) > 0 {
		mostRecent := sortedIncidents[0]
		createdAt, err := parseIncidentTime(mostRecent)
		if err != nil {
			logger.ErrorLog.Printf("Error parsing incident time: %s", err.Error())
			return lights.YellowState{}, fmt.Errorf("error parsing incident time: %s", err.Error())
		}

		if createdAt.After(startTime) && isNormalState(mostRecent.CurrentState) {
			logger.InfoLog.Printf("Most recent incident [%s] is in normal state (%s), setting light to green", 
				mostRecent.Service, mostRecent.CurrentState)
			return lights.GreenState{}, nil
		}
	}

	// Then check for any unnotified critical incidents
	for _, incident := range sortedIncidents {
		createdAt, err := parseIncidentTime(incident)
		if err != nil {
			logger.ErrorLog.Printf("Error parsing incident time: %s", err.Error())
			return lights.YellowState{}, fmt.Errorf("error parsing incident time: %s", err.Error())
		}

		if !createdAt.After(startTime) {
			continue
		}

		if !notifiedIncidents[incident.ID] && isRelevantState(incident.CurrentState) {
			notifiedIncidents[incident.ID] = true
			logger.InfoLog.Printf("New critical incident detected [%s] in state %s, setting light to red", 
				incident.Service, incident.CurrentState)
			return lights.RedState{}, nil
		}
	}

	return nil, nil
}

// parseIncidentTime parses the incident creation time
func parseIncidentTime(incident types.Incident) (time.Time, error) {
	return time.Parse(types.TimeFormat, strings.Split(incident.CreatedAt, ".")[0])
}

// isNormalState checks if the state is operational or maintenance
func isNormalState(state string) bool {
	return state == types.StateOperational || state == types.StateMaintenance
}

// isRelevantState checks if the state is critical, outage, or degraded
func isRelevantState(state string) bool {
	return state == types.StateCritical || state == types.StateOutage || state == types.StateDegraded
}
