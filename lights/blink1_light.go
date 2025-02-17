package lights

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// USB vendor ID and product ID for BLINK1MK3
	blink1VendorID  = "27b8"
	blink1ProductID = "01ed"
)

// Blink1Light implements Light interface for BLINK1MK3
type Blink1Light struct {
	devicePath string
	isOn       bool
	isBlinking bool
	current    StandardState
}

// NewBlink1Light creates a new Blink1Light instance
func NewBlink1Light() (*Blink1Light, error) {
	// Check for blink1 device in /sys/bus/usb/devices/
	// This is a Linux-specific implementation that looks for USB devices
	devicePath, err := findBlink1Device()
	if err != nil {
		return nil, fmt.Errorf("BLINK1MK3 not available: %w", err)
	}

	return &Blink1Light{
		devicePath: devicePath,
		isOn:       false,
		isBlinking: false,
		current:    StateOff,
	}, nil
}

// findBlink1Device searches for a BLINK1MK3 device in the system
func findBlink1Device() (string, error) {
	// Walk through USB devices in sysfs
	var devicePath string
	err := filepath.Walk("/sys/bus/usb/devices", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a directory
		if !info.IsDir() {
			return nil
		}

		// Check vendor ID
		vendorPath := filepath.Join(path, "idVendor")
		if vendor, err := os.ReadFile(vendorPath); err == nil {
			if string(vendor[:4]) == blink1VendorID {
				// Check product ID
				productPath := filepath.Join(path, "idProduct")
				if product, err := os.ReadFile(productPath); err == nil {
					if string(product[:4]) == blink1ProductID {
						devicePath = path
						return filepath.SkipDir // Stop traversing this directory
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error searching for device: %w", err)
	}

	if devicePath == "" {
		return "", fmt.Errorf("no BLINK1MK3 device found")
	}

	return devicePath, nil
}

// On turns on the light with a steady state
func (l *Blink1Light) On(cmd interface{}) error {
	state, ok := cmd.(StandardState)
	if !ok {
		return fmt.Errorf("invalid command type for Blink1Light")
	}

	// For now, just track the state since we don't have direct USB control implemented
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

	// For now, just track the state since we don't have direct USB control implemented
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
