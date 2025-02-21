// +build noserial

package lights

import "fmt"

// SerialLight implements Light interface for serial-based tower lights
type SerialLight struct{}

// NewSerialLight creates a new SerialLight instance
func NewSerialLight(port string, baudRate int) (*SerialLight, error) {
	return nil, fmt.Errorf("serial port support not available in this build")
}

func (l *SerialLight) On(cmd interface{}) error {
	return fmt.Errorf("serial port support not available in this build")
}

func (l *SerialLight) Clear() error {
	return fmt.Errorf("serial port support not available in this build")
}

func (l *SerialLight) Blink(cmd interface{}) error {
	return fmt.Errorf("serial port support not available in this build")
}

func (l *SerialLight) Close() error {
	return nil
}
