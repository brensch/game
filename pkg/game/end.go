package game

import (
	"image/color"
)

// End represents an end machine.
type End struct{}

// GetType returns the machine type.
func (e *End) GetType() MachineType {
	return MachineEnd
}

// GetColor returns the machine color.
func (e *End) GetColor() color.RGBA {
	return color.RGBA{R: 255, G: 150, B: 150, A: 255}
}

// Process handles object interaction for end.
func (e *End) Process(position int, objects []*Object, round int, orientation Orientation) []*Change {
	for _, obj := range objects {
		if obj.GridPosition == position {
			return []*Change{{
				StartObject: obj,
				EndObject:   nil,
			}}
		}
	}
	return nil
}

// EmitEffects emits effects from end.
func (e *End) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// For now, no effects
	return nil
}
