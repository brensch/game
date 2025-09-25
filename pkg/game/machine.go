package game

import (
	"image/color"
)

// EffectType represents dif// MachineState holds the state of a machine.
type MachineState struct {
	Machine      MachineInterface
	Effects      []EffectInterface
	Orientation  Orientation
	BeingDragged bool
	IsPlaced     bool
	RoundAdded   int
	Selected     bool
}

// EffectType represents different effects machines can have.
type EffectType int

const (
	EffectHolographic EffectType = iota
	EffectShiny
	EffectBuffSpeed
)

// DurationType represents how effect duration is measured.
type DurationType int

const (
	DurationTick DurationType = iota
	DurationRun
	DurationRound
)

// EffectInterface defines the behavior of effects.
type EffectInterface interface {
	Update(state *MachineState)
	IsExpired() bool
}

// Effect represents an effect applied to a machine.
type Effect struct {
	Type         EffectType
	Duration     int
	DurationType DurationType
}

// Update applies the effect to the machine state.
func (e *Effect) Update(state *MachineState) {
	// Apply effect logic
	switch e.Type {
	case EffectHolographic:
		// Visual effect
	case EffectBuffSpeed:
		// Increase speed or something
	}
	e.Duration--
}

// IsExpired checks if the effect has expired.
func (e *Effect) IsExpired() bool {
	return e.Duration <= 0
}

// EffectEmission represents an effect emitted by a machine to other machines.
type EffectEmission struct {
	TargetGridX int
	TargetGridY int
	Effect      EffectInterface
}

// MachineInterface defines the behavior for different machine types.
type MachineInterface interface {
	GetType() MachineType
	GetColor() color.RGBA
	Process(position int, history [][]*Object, tick int, orientation Orientation) []*Change
	EmitEffects(game *Game, state *MachineState) []EffectEmission
	GetDescription() string
}

// Change represents a change to objects.
type Change struct {
	StartObject *Object
	EndObject   *Object
}
