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
func (p *Processor) Process(obj *Object, game *Game, state *MachineState) bool {
	// Change object type
	obj.Type = (obj.Type + 1) % 3
	return false
}

// EmitEffects emits effects from processor.
func (p *Processor) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// For now, no effects
	return nil
}
