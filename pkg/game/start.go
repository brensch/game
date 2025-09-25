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
func (s *Start) Process(position int, objects [][]*Object, round int, orientation Orientation) []*Change {
	current := objects[len(objects)-1]
	if len(objects) == 1 { // first tick
		// Check if object already at position
		for _, obj := range current {
			if obj.GridPosition == position {
				return nil
			}
		}
		// Emit to next position based on orientation
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
		return []*Change{{
			StartObject: nil,
			EndObject:   &Object{GridPosition: nextPos, Type: ObjectRed},
		}}
	}
	return nil
}

// EmitEffects emits effects from start.
func (s *Start) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// For now, no effects
	return nil
}
