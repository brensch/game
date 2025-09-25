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
func (c *Conveyor) Process(position int, objects []*Object, round int) *Change {
	for _, obj := range objects {
		if obj.GridPosition == position {
			return &Change{
				Type:         ChangeTypeMove,
				FromPosition: position,
				ToPosition:   position + 1, // move right
			}
		}
	}
	return nil
}

// EmitEffects emits effects from conveyor.
func (c *Conveyor) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// For now, no effects
	return nil
}
