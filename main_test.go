package main

import (
	"reflect"
	"testing"
	"time"

	"my-incident-checker/lights"
	"my-incident-checker/poll"
	"my-incident-checker/types"
	"io"
	"log"
)

// MockLight implements lights.Light interface for testing
type MockLight struct{}

func (l *MockLight) On(cmd interface{}) error    { return nil }
func (l *MockLight) Clear() error                { return nil }
func (l *MockLight) Blink(cmd interface{}) error { return nil }

func TestAlertLogic(t *testing.T) {
	startTime := time.Date(2025, 1, 9, 3, 17, 41, 0, time.UTC)
	light := &MockLight{}

	// Create test logger that writes to io.Discard
	testLogger := &types.Logger{
		DebugLog: log.New(io.Discard, "", 0),
		InfoLog:  log.New(io.Discard, "", 0),
		WarnLog:  log.New(io.Discard, "", 0),
		ErrorLog: log.New(io.Discard, "", 0),
	}

	tests := []struct {
		name              string
		incidents         []types.Incident
		notifiedIncidents map[int]bool
		startTime         time.Time
		wantState         lights.State
		wantErr           bool
	}{
		{
			name: "new critical incident",
			incidents: []types.Incident{
				{
					ID:           1,
					Service:      "api",
					CurrentState: "critical",
					CreatedAt:    "2025-01-09T03:18:00",
					Incident: types.IncidentDetails{
						Title:       "API Down",
						Description: "API is not responding",
					},
				},
			},
			notifiedIncidents: map[int]bool{},
			startTime:         startTime,
			wantState:         lights.RedState{},
			wantErr:           false,
		},
		{
			name: "already notified incident",
			incidents: []types.Incident{
				{
					ID:           1,
					Service:      "api",
					CurrentState: "critical",
					CreatedAt:    "2025-01-09T03:18:00",
					Incident: types.IncidentDetails{
						Title:       "API Down",
						Description: "API is not responding",
					},
				},
			},
			notifiedIncidents: map[int]bool{1: true},
			startTime:         startTime,
			wantState:         nil,
			wantErr:           false,
		},
		{
			name: "old incident",
			incidents: []types.Incident{
				{
					ID:           1,
					Service:      "api",
					CurrentState: "critical",
					CreatedAt:    "2025-01-09T03:16:00",
					Incident: types.IncidentDetails{
						Title:       "API Down",
						Description: "API is not responding",
					},
				},
			},
			notifiedIncidents: map[int]bool{},
			startTime:         startTime,
			wantState:         nil,
			wantErr:           false,
		},
		{
			name: "operational state",
			incidents: []types.Incident{
				{
					ID:           1,
					Service:      "api",
					CurrentState: "operational",
					CreatedAt:    "2025-01-09T03:18:00",
					Incident: types.IncidentDetails{
						Title:       "API Recovered",
						Description: "API is back online",
					},
				},
			},
			notifiedIncidents: map[int]bool{},
			startTime:         startTime,
			wantState:         lights.GreenState{},
			wantErr:           false,
		},
		{
			name: "maintenance state",
			incidents: []types.Incident{
				{
					ID:           1,
					Service:      "api",
					CurrentState: "maintenance",
					CreatedAt:    "2025-01-09T03:18:00",
					Incident: types.IncidentDetails{
						Title:       "API Maintenance",
						Description: "Scheduled maintenance",
					},
				},
			},
			notifiedIncidents: map[int]bool{},
			startTime:         startTime,
			wantState:         lights.GreenState{},
			wantErr:           false,
		},
		{
			name: "multiple incidents - most recent critical",
			incidents: []types.Incident{
				{
					ID:           1,
					Service:      "api",
					CurrentState: "operational",
					CreatedAt:    "2025-01-09T03:18:00",
					Incident: types.IncidentDetails{
						Title:       "API Recovered",
						Description: "API is back online",
					},
				},
				{
					ID:           2,
					Service:      "database",
					CurrentState: "critical",
					CreatedAt:    "2025-01-09T03:19:00",
					Incident: types.IncidentDetails{
						Title:       "Database Down",
						Description: "Database not responding",
					},
				},
			},
			notifiedIncidents: map[int]bool{},
			startTime:         startTime,
			wantState:         lights.RedState{},
			wantErr:           false,
		},
		{
			name: "multiple incidents - most recent operational",
			incidents: []types.Incident{
				{
					ID:           1,
					Service:      "database",
					CurrentState: "critical",
					CreatedAt:    "2025-01-09T03:18:00",
					Incident: types.IncidentDetails{
						Title:       "Database Down",
						Description: "Database not responding",
					},
				},
				{
					ID:           2,
					Service:      "api",
					CurrentState: "operational",
					CreatedAt:    "2025-01-09T03:19:00",
					Incident: types.IncidentDetails{
						Title:       "API Recovered",
						Description: "API is back online",
					},
				},
			},
			notifiedIncidents: map[int]bool{},
			startTime:         startTime,
			wantState:         lights.GreenState{},
			wantErr:           false,
		},
		{
			name: "invalid time format",
			incidents: []types.Incident{
				{
					ID:           1,
					Service:      "api",
					CurrentState: "critical",
					CreatedAt:    "invalid-time",
					Incident: types.IncidentDetails{
						Title:       "API Down",
						Description: "API is not responding",
					},
				},
			},
			notifiedIncidents: map[int]bool{},
			startTime:         startTime,
			wantState:         lights.YellowState{},
			wantErr:           true,
		},
		{
			name: "storage outage with web operational",
			incidents: []types.Incident{
				{
					ID:           2,
					Service:      "storage",
					PrevState:    "maintenance",
					CurrentState: "outage",
					CreatedAt:    "2025-02-20T16:27:39.134631",
					Incident: types.IncidentDetails{
						Title:       "Storage Service Outage Detected",
						Description: "Automated systems detected abnormal behavior in storage service api.",
						Components:  []string{"api", "server", "database"},
						URL:        "https://status.joseserver.com/incidents/storage-1740068859",
					},
					History: []types.IncidentHistory{
						{
							ID:           2,
							IncidentID:   2,
							RecordedAt:   "2025-02-20T16:27:39.168785",
							Service:      "storage",
							PrevState:    "maintenance",
							CurrentState: "outage",
							Incident: types.IncidentDetails{
								Title:       "Storage Service Outage Detected",
								Description: "Automated systems detected abnormal behavior in storage service api.",
								Components:  []string{"api", "server", "database"},
								URL:        "https://status.joseserver.com/incidents/storage-1740068859",
							},
						},
					},
				},
				{
					ID:           1,
					Service:      "web",
					PrevState:    "outage",
					CurrentState: "operational",
					CreatedAt:    "2025-02-20T16:27:34.010324",
					Incident: types.IncidentDetails{
						Title:       "Unexpected Operational in Web System",
						Description: "We are investigating reports of operational performance in the web system.",
						Components:  []string{"load-balancer", "server", "api"},
						URL:        "https://status.joseserver.com/incidents/web-1740068854",
					},
					History: []types.IncidentHistory{
						{
							ID:           1,
							IncidentID:   1,
							RecordedAt:   "2025-02-20T16:27:34.097130",
							Service:      "web",
							PrevState:    "outage",
							CurrentState: "operational",
							Incident: types.IncidentDetails{
								Title:       "Unexpected Operational in Web System",
								Description: "We are investigating reports of operational performance in the web system.",
								Components:  []string{"load-balancer", "server", "api"},
								URL:        "https://status.joseserver.com/incidents/web-1740068854",
							},
						},
					},
				},
			},
			notifiedIncidents: map[int]bool{},
			startTime:         time.Date(2025, 2, 20, 16, 27, 0, 0, time.UTC),
			wantState:         lights.RedState{},
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotState, err := poll.AlertLogic(tt.incidents, light, tt.notifiedIncidents, tt.startTime, testLogger)

			if (err != nil) != tt.wantErr {
				t.Errorf("AlertLogic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (gotState == nil) != (tt.wantState == nil) {
				t.Errorf("AlertLogic() got state = %v, want state = %v", gotState, tt.wantState)
				return
			}

			if gotState != nil {
				gotType := reflect.TypeOf(gotState)
				wantType := reflect.TypeOf(tt.wantState)
				if gotType != wantType {
					t.Errorf("AlertLogic() got state type = %v, want state type = %v", gotType, wantType)
				}
			}
		})
	}
}
