package lights

import (
	"fmt"
	"log"

	"github.com/tarm/serial"
)

// SerialLight implements Light interface for serial-based tower lights
type SerialLight struct {
	port     string
	baudRate int
}

// NewSerialLight creates a new SerialLight instance
func NewSerialLight(port string, baudRate int) *SerialLight {
	return &SerialLight{
		port:     port,
		baudRate: baudRate,
	}
}

func (l *SerialLight) openPort() (*serial.Port, error) {
	c := &serial.Config{
		Name: l.port,
		Baud: l.baudRate,
	}
	return serial.OpenPort(c)
}

func (l *SerialLight) On(cmd interface{}) error {
	state, ok := cmd.(StandardState)
	if !ok {
		return fmt.Errorf("invalid command type for SerialLight")
	}

	var cmdByte byte
	switch state {
	case StateRed:
		cmdByte = cmdRedOn
	case StateYellow:
		cmdByte = cmdYellowOn
	case StateGreen:
		cmdByte = cmdGreenOn
	default:
		return fmt.Errorf("unsupported state: %s", state)
	}

	if err := l.Clear(); err != nil {
		return fmt.Errorf("failed to clear light state: %w", err)
	}

	s, err := l.openPort()
	if err != nil {
		return err
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Printf("Error closing serial port: %s", err.Error())
		}
	}()

	return sendCommand(s, cmdByte)
}

func (l *SerialLight) Blink(cmd interface{}) error {
	state, ok := cmd.(StandardState)
	if !ok {
		return fmt.Errorf("invalid command type for SerialLight")
	}

	var cmdByte byte
	switch state {
	case StateRed:
		cmdByte = cmdRedBlink
	case StateYellow:
		cmdByte = cmdYellowBlink
	case StateGreen:
		cmdByte = cmdGreenBlink
	default:
		return fmt.Errorf("unsupported state: %s", state)
	}

	s, err := l.openPort()
	if err != nil {
		return err
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Printf("Error closing serial port: %s", err.Error())
		}
	}()

	return sendCommand(s, cmdByte)
}

func (l *SerialLight) Clear() error {
	s, err := l.openPort()
	if err != nil {
		return err
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Printf("Error closing serial port: %s", err.Error())
		}
	}()

	initialCommands := []byte{
		cmdBuzzerOff,
		cmdRedOff,
		cmdYellowOff,
		cmdGreenOff,
	}

	for _, cmd := range initialCommands {
		if err := sendCommand(s, cmd); err != nil {
			return err
		}
	}
	return nil
}
