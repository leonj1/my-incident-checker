package main

import (
	"reflect"
	"testing"
	"time"

	"my-incident-checker/lights"
)

// MockLight implements lights.Light interface for testing
type MockLight struct{}

func (l *MockLight) On(cmd interface{}) error    { return nil }
func (l *MockLight) Clear() error                { return nil }
func (l *MockLight) Blink(cmd interface{}) error { return nil }

func TestAlertLogic(t *testing.T) {
	startTime := time.Date(2025, 1, 9, 3, 17, 41, 0, time.UTC)
	light := &MockLight{}

	tests := []struct {
		name              string
		incidents         []Incident
		notifiedIncidents map[int]bool
		startTime         time.Time
		wantState         lights.State
		wantErr           bool
	}{
		{
			name: "new critical incident",
			incidents: []Incident{
				{
					ID:           1,
					Service:      "api",
					CurrentState: "critical",
					CreatedAt:    "2025-01-09T03:18:00",
					Incident: IncidentDetails{
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
			incidents: []Incident{
				{
					ID:           1,
					Service:      "api",
					CurrentState: "critical",
					CreatedAt:    "2025-01-09T03:18:00",
					Incident: IncidentDetails{
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
			incidents: []Incident{
				{
					ID:           1,
					Service:      "api",
					CurrentState: "critical",
					CreatedAt:    "2025-01-09T03:16:00",
					Incident: IncidentDetails{
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
			incidents: []Incident{
				{
					ID:           1,
					Service:      "api",
					CurrentState: "operational",
					CreatedAt:    "2025-01-09T03:18:00",
					Incident: IncidentDetails{
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
			incidents: []Incident{
				{
					ID:           1,
					Service:      "api",
					CurrentState: "maintenance",
					CreatedAt:    "2025-01-09T03:18:00",
					Incident: IncidentDetails{
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
			incidents: []Incident{
				{
					ID:           1,
					Service:      "api",
					CurrentState: "operational",
					CreatedAt:    "2025-01-09T03:18:00",
					Incident: IncidentDetails{
						Title:       "API Recovered",
						Description: "API is back online",
					},
				},
				{
					ID:           2,
					Service:      "database",
					CurrentState: "critical",
					CreatedAt:    "2025-01-09T03:19:00",
					Incident: IncidentDetails{
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
			incidents: []Incident{
				{
					ID:           1,
					Service:      "database",
					CurrentState: "critical",
					CreatedAt:    "2025-01-09T03:18:00",
					Incident: IncidentDetails{
						Title:       "Database Down",
						Description: "Database not responding",
					},
				},
				{
					ID:           2,
					Service:      "api",
					CurrentState: "operational",
					CreatedAt:    "2025-01-09T03:19:00",
					Incident: IncidentDetails{
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
			incidents: []Incident{
				{
					ID:           1,
					Service:      "api",
					CurrentState: "critical",
					CreatedAt:    "invalid-time",
					Incident: IncidentDetails{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotState, err := AlertLogic(tt.incidents, light, tt.notifiedIncidents, tt.startTime)

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
