// +build noserial

package lights

import "fmt"

// SerialLight implements Light interface for serial-based tower lights
type SerialLight struct{}

// Error message constant for unavailable serial support
const errNoSerialSupport = "serial port support not available in this build"

// NewSerialLight creates a new SerialLight instance
func NewSerialLight(port string, baudRate int) (*SerialLight, error) {
	return nil, fmt.Errorf(errNoSerialSupport)
}

func (l *SerialLight) On(cmd interface{}) error {
	return fmt.Errorf(errNoSerialSupport)
}

func (l *SerialLight) Clear() error {
	return fmt.Errorf(errNoSerialSupport)
}

func (l *SerialLight) Blink(cmd interface{}) error {
	return fmt.Errorf(errNoSerialSupport)
}

func (l *SerialLight) Close() error {
	return fmt.Errorf(errNoSerialSupport)
}
