package game

import (
	"image/color"
)

// Splitter represents a splitter machine.
type Splitter struct{}

// GetType returns the machine type.
func (s *Splitter) GetType() MachineType {
	return MachineSplitter
}

// GetRoles returns the machine roles.
func (s *Splitter) GetRoles() []MachineRole {
	return []MachineRole{RoleMover, RoleProducer}
}

// GetRoleNames returns the names of the machine roles.
func (s *Splitter) GetRoleNames() []string {
	return []string{"Mover", "Producer"}
}

// GetColor returns the machine color.
func (s *Splitter) GetColor() color.RGBA {
	return color.RGBA{R: 150, G: 150, B: 255, A: 255} // Light blue
}

// Process handles object interaction for splitter.
func (s *Splitter) Process(position int, history [][]*Object, tick int, orientation Orientation) []*Change {
	current := history[len(history)-1]
	for _, obj := range current {
		if obj.GridPosition == position {
			nextPos := GetAdjacentPosition(position, orientation)
			halfValue := obj.Score.Value / 2
			if halfValue < 1 {
				halfValue = 1 // Minimum value of 1
			}
			// Create two objects of half value
			return []*Change{{
				StartObject: obj,
				EndObject:   &Object{GridPosition: nextPos, Type: obj.Type, Score: &Score{Value: halfValue, MultAdd: obj.Score.MultAdd, MultMult: obj.Score.MultMult}},
				Score:       nil,
			}, {
				StartObject: obj,
				EndObject:   &Object{GridPosition: nextPos, Type: obj.Type, Score: &Score{Value: halfValue, MultAdd: obj.Score.MultAdd, MultMult: obj.Score.MultMult}},
				Score:       nil,
			}}
		}
	}
	return nil
}

// EmitEffects emits effects from splitter.
func (s *Splitter) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// For now, no effects
	return nil
}

// GetDescription returns the machine description.
func (s *Splitter) GetDescription() string {
	return "Takes one object and splits it into two objects of half the value, moving them forward."
}

// GetCost returns the cost to place this machine.
func (s *Splitter) GetCost() int {
	return 4
}

// GetName returns the machine name.
func (s *Splitter) GetName() string {
	return "Splitter"
}
