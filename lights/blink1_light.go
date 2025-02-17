package lights

import (
	"fmt"
)

// Blink1Light implements Light interface for BLINK1MK3
type Blink1Light struct {
	// For now, we'll use a simple implementation without the actual device
	// since we don't have the real hardware library
	isOn       bool
	isBlinking bool
	current    StandardState
}

// NewBlink1Light creates a new Blink1Light instance
func NewBlink1Light() (*Blink1Light, error) {
	// For now, always return error to fall back to SerialLight
	// until we have the actual BLINK1MK3 hardware and library
	return nil, fmt.Errorf("BLINK1MK3 not available")
}

// On turns on the light with a steady state
func (l *Blink1Light) On(cmd interface{}) error {
	state, ok := cmd.(StandardState)
	if !ok {
		return fmt.Errorf("invalid command type for Blink1Light")
	}

	l.isOn = true
	l.isBlinking = false
	l.current = state
	return nil
}

// Blink turns on the light with a blinking state
func (l *Blink1Light) Blink(cmd interface{}) error {
	state, ok := cmd.(StandardState)
	if !ok {
		return fmt.Errorf("invalid command type for Blink1Light")
	}

	l.isOn = true
	l.isBlinking = true
	l.current = state
	return nil
}

// Clear turns off the light and resets all states
func (l *Blink1Light) Clear() error {
	l.isOn = false
	l.isBlinking = false
	l.current = StateOff
	return nil
}

// Close cleans up resources by clearing the light state
func (l *Blink1Light) Close() {
	l.Clear()
}
