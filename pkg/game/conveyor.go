package game

import (
	"image/color"
)

// Conveyor represents a conveyor machine.
type Conveyor struct{}

// GetType returns the machine type.
func (c *Conveyor) GetType() MachineType {
	return MachineConveyor
}

// GetColor returns the machine color.
func (c *Conveyor) GetColor() color.RGBA {
	return color.RGBA{R: 200, G: 200, B: 200, A: 255}
}

// Process handles object interaction for conveyor.
func (c *Conveyor) Process(obj *Object, game *Game, state *MachineState) bool {
	// Move object down
	obj.Y += 5
	return false
}

// EmitEffects emits effects from conveyor.
func (c *Conveyor) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// For now, no effects
	return nil
}
