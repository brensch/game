package game

import (
	"image/color"
)

// Processor represents a processor machine.
type Processor struct{}

// GetType returns the machine type.
func (p *Processor) GetType() MachineType {
	return MachineProcessor
}

// GetColor returns the machine color.
func (p *Processor) GetColor() color.RGBA {
	return color.RGBA{R: 100, G: 200, B: 100, A: 255}
}

// Process handles object interaction for processor.
func (p *Processor) Process(position int, history [][]*Object, round int, orientation Orientation) []*Change {
	current := history[len(history)-1]
	for _, obj := range current {
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
			return []*Change{{
				StartObject: obj,
				EndObject:   &Object{GridPosition: nextPos, Type: (obj.Type + 1) % 3},
			}}
		}
	}
	return nil
}

// EmitEffects emits effects from processor.
func (p *Processor) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// For now, no effects
	return nil
}
