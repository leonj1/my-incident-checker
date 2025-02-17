package lights

import (
	"fmt"

	"github.com/tarm/serial"
)

// sendCommand sends a command byte to the serial port
func sendCommand(port *serial.Port, cmd byte) error {
	_, err := port.Write([]byte{cmd})
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}
	return nil
}
