package lights

// Command bytes for LEDs and buzzer
const (
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
