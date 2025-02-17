package lights

import "github.com/tarm/serial"

// sendCommand sends a command byte to the serial port
func sendCommand(port *serial.Port, cmd byte) error {
	_, err := port.Write([]byte{cmd})
	return err
}
