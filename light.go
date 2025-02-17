package main

import (
	"log"

	"github.com/tarm/serial"
)

const (
	serialPort = "/dev/ttyUSB0" // Change to the serial/COM port of the tower light
	baudRate   = 9600

	// Command bytes for LEDs and buzzer
	RED_ON    byte = 0x11
	RED_OFF   byte = 0x21
	RED_BLINK byte = 0x41

	YELLOW_ON    byte = 0x12
	YELLOW_OFF   byte = 0x22
	YELLOW_BLINK byte = 0x42

	GREEN_ON    byte = 0x14
	GREEN_OFF   byte = 0x24
	GREEN_BLINK byte = 0x44

	BUZZER_ON    byte = 0x18
	BUZZER_OFF   byte = 0x28
	BUZZER_BLINK byte = 0x48
)

type TrafficLight struct{}

type LightState interface {
	Apply(light *TrafficLight) error
}

type RedLight struct{}
type GreenLight struct{}
type YellowLight struct{}

func (s RedLight) Apply(light *TrafficLight) error {
	return light.On(RED_ON)
}

func (s GreenLight) Apply(light *TrafficLight) error {
	return light.On(GREEN_ON)
}

func (s YellowLight) Apply(light *TrafficLight) error {
	return light.On(YELLOW_ON)
}

func (l *TrafficLight) On(onCmd byte) error {
	l.Clear()

	// Configure serial port settings
	c := &serial.Config{
		Name: serialPort,
		Baud: baudRate,
	}

	// Open the serial port
	s, err := serial.OpenPort(c)
	if err != nil {
		return err
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Printf("Error closing serial port: %s", err.Error())
		}
	}()

	if err := sendCommand(s, onCmd); err != nil {
		return err
	}
	return nil
}

func (l *TrafficLight) Clear() error {
	// Configure serial port settings
	c := &serial.Config{
		Name: serialPort,
		Baud: baudRate,
	}

	// Open the serial port
	s, err := serial.OpenPort(c)
	if err != nil {
		return err
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Printf("Error closing serial port: %s", err.Error())
		}
	}()

	// Clean up any old state by turning off all LEDs and buzzer
	initialCommands := []byte{
		BUZZER_OFF,
		RED_OFF,
		YELLOW_OFF,
		GREEN_OFF,
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
