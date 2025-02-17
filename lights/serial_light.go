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

const (
    // Command bytes for LEDs and buzzer
    cmdRedOn    byte = 0x11
    cmdRedOff   byte = 0x21
    cmdRedBlink byte = 0x41

    cmdYellowOn    byte = 0x12
    cmdYellowOff   byte = 0x22
    cmdYellowBlink byte = 0x42

    cmdGreenOn    byte = 0x14
    cmdGreenOff   byte = 0x24
    cmdGreenBlink byte = 0x44

    cmdBuzzerOn    byte = 0x18
    cmdBuzzerOff   byte = 0x28
    cmdBuzzerBlink byte = 0x48
)

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

    l.Clear()

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

func sendCommand(port *serial.Port, cmd byte) error {
    _, err := port.Write([]byte{cmd})
    return err
}