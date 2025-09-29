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

// GetRoles returns the machine roles.
func (c *Conveyor) GetRoles() []MachineRole {
	return []MachineRole{RoleMover}
}

// GetRoleNames returns the names of the machine roles.
func (c *Conveyor) GetRoleNames() []string {
	return []string{"Mover"}
}

// GetColor returns the machine color.
func (c *Conveyor) GetColor() color.RGBA {
	return color.RGBA{R: 200, G: 200, B: 200, A: 255}
}

// Process handles object interaction for conveyor.
func (c *Conveyor) Process(position int, history [][]*Object, tick int, orientation Orientation) []*Change {
	current := history[len(history)-1]
	for _, obj := range current {
		if obj.GridPosition == position {
			nextPos := GetAdjacentPosition(position, orientation)
			return []*Change{{
				StartObject: obj,
				EndObject:   &Object{GridPosition: nextPos, Type: obj.Type, Score: obj.Score},
				Score:       nil,
			}}
		}
	}
	return nil
}

// EmitEffects emits effects from conveyor.
func (c *Conveyor) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// For now, no effects
	return nil
}

// GetDescription returns the machine description.
func (c *Conveyor) GetDescription() string {
	return "Moves objects in the direction it's facing."
}

// GetCost returns the cost to place this machine.
func (c *Conveyor) GetCost() int {
	return 1
}

// GetName returns the machine name.
func (c *Conveyor) GetName() string {
	return "Conveyor"
}
