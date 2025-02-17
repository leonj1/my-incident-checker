package lights

// State represents the possible states of a light
type State interface {
	Apply(light Light) error
}

// Light defines the interface for different light implementations
type Light interface {
	// On turns on a specific light state
	On(cmd interface{}) error
	// Clear turns off all lights
	Clear() error
	// Blink makes a light blink
	Blink(cmd interface{}) error
}

// StandardState represents common light states
type StandardState string

const (
	StateRed    StandardState = "red"
	StateYellow StandardState = "yellow"
	StateGreen  StandardState = "green"
	StateOff    StandardState = "off"
)

// RedState implements State
type RedState struct{}

func (s RedState) Apply(light Light) error {
	return light.On(StateRed)
}

// YellowState implements State
type YellowState struct{}

func (s YellowState) Apply(light Light) error {
	return light.On(StateYellow)
}

// GreenState implements State
type GreenState struct{}

func (s GreenState) Apply(light Light) error {
	return light.On(StateGreen)
}

// BlinkingRedState implements State for blinking red light
type BlinkingRedState struct{}

func (s BlinkingRedState) Apply(light Light) error {
	return light.Blink(StateRed)
}

// BlinkingYellowState implements State for blinking yellow light
type BlinkingYellowState struct{}

func (s BlinkingYellowState) Apply(light Light) error {
	return light.Blink(StateYellow)
}

// BlinkingGreenState implements State for blinking green light
type BlinkingGreenState struct{}

func (s BlinkingGreenState) Apply(light Light) error {
	return light.Blink(StateGreen)
}
