package lights

import (
    "fmt"
)

// Blink1Light implements Light interface for BLINK1MK3
type Blink1Light struct {
    // For now, we'll use a simple implementation without the actual device
    // since we don't have the real hardware library
    isOn    bool
    current StandardState
}

// NewBlink1Light creates a new Blink1Light instance
func NewBlink1Light() (*Blink1Light, error) {
    // For now, always return error to fall back to SerialLight
    // until we have the actual BLINK1MK3 hardware and library
    return nil, fmt.Errorf("BLINK1MK3 not available")
}

func (l *Blink1Light) On(cmd interface{}) error {
    state, ok := cmd.(StandardState)
    if !ok {
        return fmt.Errorf("invalid command type for Blink1Light")
    }

    l.isOn = true
    l.current = state
    return nil
}

func (l *Blink1Light) Blink(cmd interface{}) error {
    state, ok := cmd.(StandardState)
    if !ok {
        return fmt.Errorf("invalid command type for Blink1Light")
    }

    l.isOn = true
    l.current = state
    return nil
}

func (l *Blink1Light) Clear() error {
    l.isOn = false
    l.current = StateOff
    return nil
}

func (l *Blink1Light) Close() {
    l.Clear()
}