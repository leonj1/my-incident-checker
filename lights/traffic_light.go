package lights

import (
	"fmt"
	"log"

	"github.com/tarm/serial"
)

// TrafficLight implements Light interface for serial-based tower lights
type TrafficLight struct {
	port     string
	baudRate int
}

// NewTrafficLight creates a new TrafficLight instance
func NewTrafficLight(port string, baudRate int) *TrafficLight {
	return &TrafficLight{
		port:     port,
		baudRate: baudRate,
	}
}

func (l *TrafficLight) openPort() (*serial.Port, error) {
	c := &serial.Config{
		Name: l.port,
		Baud: l.baudRate,
	}
	return serial.OpenPort(c)
}

func (l *TrafficLight) On(cmd interface{}) error {
	state, ok := cmd.(StandardState)
	if !ok {
		return fmt.Errorf("invalid command type for TrafficLight")
	}

	// Maps the StandardState to corresponding command bytes for the traffic light.
	// Command bytes are defined in commands.go with the following values:
	// - cmdRedOn (0x11): Turns on the red light
	// - cmdYellowOn (0x12): Turns on the yellow light
	// - cmdGreenOn (0x14): Turns on the green light
	// These command bytes are specific to the hardware protocol used by the traffic light.
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

func (l *TrafficLight) Blink(cmd interface{}) error {
	state, ok := cmd.(StandardState)
	if !ok {
		return fmt.Errorf("invalid command type for TrafficLight")
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

func (l *TrafficLight) Clear() error {
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
