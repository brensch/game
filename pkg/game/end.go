package game

import (
	"image/color"
)

// GeneralConsumer represents a general consumer machine.
type GeneralConsumer struct{}

// GetType returns the machine type.
func (e *GeneralConsumer) GetType() MachineType {
	return MachineGeneralConsumer
}

// GetRoles returns the machine roles.
func (e *GeneralConsumer) GetRoles() []MachineRole {
	return []MachineRole{RoleConsumer}
}

// GetRoleNames returns the names of the machine roles.
func (e *GeneralConsumer) GetRoleNames() []string {
	return []string{"Consumer"}
} // GetColor returns the machine color.
func (e *GeneralConsumer) GetColor() color.RGBA {
	return color.RGBA{R: 255, G: 150, B: 150, A: 255}
}

// Process handles object interaction for general consumer.
func (e *GeneralConsumer) Process(position int, history [][]*Object, tick int, orientation Orientation) []*Change {
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

// EmitEffects emits effects from general consumer.
func (e *GeneralConsumer) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// For now, no effects
	return nil
}

// GetDescription returns the machine description.
func (e *GeneralConsumer) GetDescription() string {
	return "Collects objects that reach it, scoring points based on their color."
}

// GetCost returns the cost to place this machine.
func (e *GeneralConsumer) GetCost() int {
	return 0
}

// GetName returns the machine name.
func (e *GeneralConsumer) GetName() string {
	return "General Consumer"
}
