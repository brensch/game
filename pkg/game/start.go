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
func (s *Start) Process(position int, objects []*Object, round int) *Change {
	if round%60 == 0 {
		// Check if object already at position
		for _, obj := range objects {
			if obj.GridPosition == position {
				return nil
			}
		}
		// Emit to next position
		return &Change{
			Type:         ChangeTypeCreate,
			GridPosition: position + 1,
			ObjectType:   ObjectRed,
		}
	}
	return nil
}

// EmitEffects emits effects from start.
func (s *Start) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// For now, no effects
	return nil
}
