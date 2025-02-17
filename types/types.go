package types

import "log"

// Log levels
const (
	LogDebug = "DEBUG"
	LogInfo  = "INFO"
	LogWarn  = "WARN"
	LogError = "ERROR"
)

const (
	TimeFormat = "2006-01-02T15:04:05.999999"

	StateOperational = "operational"
	StateMaintenance = "maintenance"
	StateCritical    = "critical"
	StateOutage      = "outage"
	StateDegraded    = "degraded"
)

type Logger struct {
	DebugLog *log.Logger
	InfoLog  *log.Logger
	WarnLog  *log.Logger
	ErrorLog *log.Logger
}

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
