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
func (c *Conveyor) Process(position int, objects []*Object, round int, orientation Orientation) *Change {
	for _, obj := range objects {
		if obj.GridPosition == position {
			nextPos := position + 1 // default east
			switch orientation {
			case OrientationNorth:
				nextPos = position - 7
			case OrientationEast:
				nextPos = position + 1
			case OrientationSouth:
				nextPos = position + 7
			case OrientationWest:
				nextPos = position - 1
			}
			return &Change{
				Type:         ChangeTypeMove,
				FromPosition: position,
				ToPosition:   nextPos,
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
