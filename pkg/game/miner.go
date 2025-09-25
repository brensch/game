package game

import (
	"image/color"
)

// Miner represents a miner machine.
type Miner struct{}

// GetType returns the machine type.
func (m *Miner) GetType() MachineType {
	return MachineMiner
}

// GetColor returns the machine color.
func (m *Miner) GetColor() color.RGBA {
	return color.RGBA{R: 139, G: 69, B: 19, A: 255} // Brown
}

// Process handles object interaction for miner.
func (m *Miner) Process(position int, history [][]*Object, tick int, orientation Orientation) []*Change {
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
		nextPos := GetAdjacentPosition(position, orientation)

		// fmt.Printf("processed miner. tick: %d, nextPos: %d, position: %d\n", tick, nextPos, position)
		return []*Change{{
			StartObject: &Object{GridPosition: position, Type: objType},
			EndObject:   &Object{GridPosition: nextPos, Type: objType},
		}}
	}
	return nil
}

// EmitEffects emits effects from miner.
func (m *Miner) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// For now, no effects
	return nil
}

// GetDescription returns the machine description.
func (m *Miner) GetDescription() string {
	return "Generates red, green, and blue objects at the beginning of each run."
}
