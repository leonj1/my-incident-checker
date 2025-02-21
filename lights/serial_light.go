package lights

import (
	"fmt"

	"github.com/tarm/serial"
)

// SerialLight implements Light interface for serial-based tower lights
type SerialLight struct {
	port     string
	baudRate int
	conn     *serial.Port
}

// NewSerialLight creates a new SerialLight instance
func NewSerialLight(port string, baudRate int) (*SerialLight, error) {
	light := &SerialLight{
		port:     port,
		baudRate: baudRate,
	}

	// Test the connection immediately
	conn, err := light.openPort()
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port: %w", err)
	}
	light.conn = conn

	return light, nil
}

func (l *SerialLight) openPort() (*serial.Port, error) {
	if l.conn != nil {
		return l.conn, nil
	}

	c := &serial.Config{
		Name: l.port,
		Baud: l.baudRate,
	}
	conn, err := serial.OpenPort(c)
	if err != nil {
		return nil, err
	}
	l.conn = conn
	return conn, nil
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

	return sendCommand(l.conn, cmdByte)
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

	return sendCommand(l.conn, cmdByte)
}

func (l *SerialLight) Clear() error {
	initialCommands := []byte{
		cmdBuzzerOff,
		cmdRedOff,
		cmdYellowOff,
		cmdGreenOff,
	}

	for _, cmd := range initialCommands {
		if err := sendCommand(l.conn, cmd); err != nil {
			return err
		}
	}
	return nil
}

// Close implements io.Closer interface
func (l *SerialLight) Close() error {
	if l.conn != nil {
		err := l.conn.Close()
		l.conn = nil
		return err
	}
	return nil
}
