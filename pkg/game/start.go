package game

import (
	"image/color"
)

// Start represents a start machine.
type Start struct{}

// GetType returns the machine type.
func (s *Start) GetType() MachineType {
	return MachineStart
}

// GetColor returns the machine color.
func (s *Start) GetColor() color.RGBA {
	return color.RGBA{R: 150, G: 255, B: 150, A: 255}
}

// Process handles object interaction for start.
func (s *Start) Process(obj *Object, game *Game, state *MachineState) bool {
	// Start machine emits objects, doesn't process them
	return false
}

// EmitEffects emits effects from start.
func (s *Start) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// For now, no effects
	return nil
}
