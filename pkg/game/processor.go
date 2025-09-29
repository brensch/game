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

// GetRoles returns the machine roles.
func (p *Processor) GetRoles() []MachineRole {
	return []MachineRole{RoleConsumer, RoleProducer, RoleMover}
}

// GetRoleNames returns the names of the machine roles.
func (p *Processor) GetRoleNames() []string {
	return []string{"Consumer", "Producer", "Mover"}
}

// GetColor returns the machine color.
func (p *Processor) GetColor() color.RGBA {
	return color.RGBA{R: 100, G: 200, B: 100, A: 255}
}

// Process handles object interaction for processor.
func (p *Processor) Process(position int, history [][]*Object, tick int, orientation Orientation) []*Change {
	current := history[len(history)-1]
	for _, obj := range current {
		if obj.GridPosition == position {
			nextPos := GetAdjacentPosition(position, orientation)
			multAdd := 0
			if obj.Type == ObjectGreen {
				multAdd = 1
			}
			return []*Change{{
				StartObject: obj,
				EndObject:   &Object{GridPosition: nextPos, Type: (obj.Type + 1) % 3, Score: &Score{Value: obj.Score.Value, MultAdd: obj.Score.MultAdd + multAdd, MultMult: obj.Score.MultMult}},
				Score:       nil,
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

// GetDescription returns the machine description.
func (p *Processor) GetDescription() string {
	return "Transforms objects to the next color and moves them forward. Gives +1 multiplier when processing green objects."
}

// GetName returns the machine name.
func (p *Processor) GetName() string {
	return "Processor"
}
