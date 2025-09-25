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
func (s *Start) Process(position int, history [][]*Object, round int, orientation Orientation) []*Change {
	if len(history) <= 3 {
		// Emit one object per tick for first 3 ticks
		var objType ObjectType
		switch len(history) {
		case 1:
			objType = ObjectRed
		case 2:
			objType = ObjectGreen
		case 3:
			objType = ObjectBlue
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
			StartObject: &Object{GridPosition: position, Type: objType},
			EndObject:   &Object{GridPosition: nextPos, Type: objType},
		}}
	}
	return nil
}

// EmitEffects emits effects from start.
func (s *Start) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// For now, no effects
	return nil
}
