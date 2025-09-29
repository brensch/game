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

// GetRoles returns the machine roles.
func (e *End) GetRoles() []MachineRole {
	return []MachineRole{RoleConsumer}
}

// GetRoleNames returns the names of the machine roles.
func (e *End) GetRoleNames() []string {
	return []string{"Consumer"}
}

// GetColor returns the machine color.
func (e *End) GetColor() color.RGBA {
	return color.RGBA{R: 255, G: 150, B: 150, A: 255}
}

// Process handles object interaction for end.
func (e *End) Process(position int, history [][]*Object, round int, orientation Orientation) []*Change {
	current := history[len(history)-1]
	for _, obj := range current {
		if obj.GridPosition == position {
			return []*Change{{
				StartObject: obj,
				EndObject:   nil,
				Score:       obj.Score,
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

// GetDescription returns the machine description.
func (e *End) GetDescription() string {
	return "Collects objects that reach it, scoring points based on their color."
}

// GetCost returns the cost to place this machine.
func (e *End) GetCost() int {
	return 0
}

// GetName returns the machine name.
func (e *End) GetName() string {
	return "End"
}
