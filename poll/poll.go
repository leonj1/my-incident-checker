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
	var cachedIncidents []types.Incident
	currentLightState := "green" // Track current light state

	for {
		// Connectivity check disabled

		incidents, err := fetchIncidents()
		if err != nil {
			logger.ErrorLog.Printf("Failed to fetch incidents: %s", err.Error())
			time.Sleep(pollInterval)
			continue
		}

		// Log state changes first
		state, err := AlertLogic(incidents, light, notifiedIncidents, startTime, logger, currentLightState)
		if err != nil {
			logger.ErrorLog.Printf("Alert logic error: %s", err.Error())
		} else if state != nil {
			stateColor := "unknown"
			switch state.(type) {
			case lights.RedState:
				stateColor = "red"
			case lights.YellowState:
				stateColor = "yellow"
			case lights.GreenState:
				stateColor = "green"
			}
			if stateColor != currentLightState {
				logger.InfoLog.Printf("⚠️ Light color changed to: %s", strings.ToUpper(stateColor))
				currentLightState = stateColor
			}
			if err := state.Apply(light); err != nil {
				logger.ErrorLog.Printf("Failed to apply light state: %s", err.Error())
			}
		}

		// Then log if incidents have changed
		if !incidentsEqual(incidents, cachedIncidents) {
			logIncidentChanges(logger, cachedIncidents, incidents)
			cachedIncidents = make([]types.Incident, len(incidents))
			copy(cachedIncidents, incidents)
		}

		for _, incident := range incidents {
			if !seenIncidents[incident.ID] {
				logger.InfoLog.Printf("New incident detected: [%s] %s - Current State: %s",
					incident.Service,
					incident.Incident.Title,
					incident.CurrentState)
				seenIncidents[incident.ID] = true
			}
		}

		time.Sleep(pollInterval)
	}
}

// incidentsEqual compares two slices of incidents for equality, regardless of order
func incidentsEqual(a, b []types.Incident) bool {
	if len(a) != len(b) {
		return false
	}

	// Create a map to track incidents by their unique properties
	incidentMap := make(map[string]int)

	// Count occurrences of each incident in slice a
	for _, inc := range a {
		key := fmt.Sprintf("%d-%s-%s", inc.ID, inc.CurrentState, inc.Service)
		incidentMap[key]++
	}

	// Verify each incident in slice b exists in the map
	for _, inc := range b {
		key := fmt.Sprintf("%d-%s-%s", inc.ID, inc.CurrentState, inc.Service)
		count := incidentMap[key]
		if count == 0 {
			return false
		}
		incidentMap[key]--
	}

	return true
}

// logIncidentChanges logs detailed changes between two sets of incidents
func logIncidentChanges(logger *types.Logger, oldIncidents, newIncidents []types.Incident) {
	logger.DebugLog.Printf("Incidents changed: previous count=%d, new count=%d",
		len(oldIncidents), len(newIncidents))

	// Create map of old incidents for easy lookup
	oldIncidentMap := make(map[int]types.Incident)
	for _, inc := range oldIncidents {
		oldIncidentMap[inc.ID] = inc
	}

	// Check for new or modified incidents
	for _, newInc := range newIncidents {
		oldInc, exists := oldIncidentMap[newInc.ID]
		if !exists {
			logger.DebugLog.Printf("New incident detected [%d]: %s - %s",
				newInc.ID, newInc.Service, newInc.CurrentState)
			continue
		}
		if newInc.CurrentState != oldInc.CurrentState {
			logger.DebugLog.Printf("Incident [%d] state changed: %s -> %s",
				newInc.ID, oldInc.CurrentState, newInc.CurrentState)
		}
	}

	// Check for removed incidents
	newIncidentMap := make(map[int]struct{})
	for _, inc := range newIncidents {
		newIncidentMap[inc.ID] = struct{}{}
	}
	for _, oldInc := range oldIncidents {
		if _, exists := newIncidentMap[oldInc.ID]; !exists {
			logger.DebugLog.Printf("Incident removed [%d]: %s", oldInc.ID, oldInc.Service)
		}
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
func AlertLogic(incidents []types.Incident, light lights.Light, notifiedIncidents map[int]bool, startTime time.Time, logger *types.Logger, currentLightState string) (lights.State, error) {
	if len(incidents) == 0 {
		// Clear notification history and return green state when no incidents
		for k := range notifiedIncidents {
			delete(notifiedIncidents, k)
		}
		if currentLightState != "green" {
			logger.InfoLog.Printf("No active incidents, setting light to green")
		}
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
			if currentLightState != "green" {
				logger.InfoLog.Printf("Most recent incident [%s] is in normal state (%s), setting light to green",
					mostRecent.Service, mostRecent.CurrentState)
			}
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
			if currentLightState != "red" {
				logger.InfoLog.Printf("New critical incident detected [%s] in state %s, setting light to red",
					incident.Service, incident.CurrentState)
			}
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

// isRelevantState checks if the state is critical, outage, degraded, or major
func isRelevantState(state string) bool {
	// Convert to lowercase for case-insensitive comparison
	stateLower := strings.ToLower(state)
	return stateLower == strings.ToLower(types.StateCritical) || 
		stateLower == strings.ToLower(types.StateOutage) || 
		stateLower == strings.ToLower(types.StateDegraded) || 
		stateLower == strings.ToLower(types.StateMajor)
}
