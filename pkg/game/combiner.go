package game

import (
	"image/color"
)

// Combiner represents a combiner machine.
type Combiner struct{}

// GetType returns the machine type.
func (c *Combiner) GetType() MachineType {
	return MachineCombiner
}

// GetRoles returns the machine roles.
func (c *Combiner) GetRoles() []MachineRole {
	return []MachineRole{RoleConsumer, RoleProducer}
}

// GetRoleNames returns the names of the machine roles.
func (c *Combiner) GetRoleNames() []string {
	return []string{"Consumer", "Producer"}
}

// GetColor returns the machine color.
func (c *Combiner) GetColor() color.RGBA {
	return color.RGBA{R: 255, G: 0, B: 255, A: 255} // Magenta
}

// Process handles object interaction for combiner.
func (c *Combiner) Process(position int, history [][]*Object, tick int, orientation Orientation) []*Change {
	current := history[len(history)-1]
	var objectsAtPos []*Object
	for _, obj := range current {
		if obj.GridPosition == position {
			objectsAtPos = append(objectsAtPos, obj)
		}
	}
	if len(objectsAtPos) >= 2 {
		// Combine the first two objects
		obj1, obj2 := objectsAtPos[0], objectsAtPos[1]
		nextPos := GetAdjacentPosition(position, orientation)
		combinedValue := obj1.Score.Value + obj2.Score.Value
		combinedMultAdd := obj1.Score.MultAdd + obj2.Score.MultAdd
		combinedMultMult := obj1.Score.MultMult * obj2.Score.MultMult // Or average, but multiply for synergy
		return []*Change{{
			StartObject: obj1,
			EndObject:   &Object{GridPosition: nextPos, Type: obj1.Type, Score: &Score{Value: combinedValue, MultAdd: combinedMultAdd, MultMult: combinedMultMult}},
			Score:       nil,
		}, {
			StartObject: obj2,
			EndObject:   nil, // Remove the second object
			Score:       nil,
		}}
	}
	return nil
}

// EmitEffects emits effects from combiner.
func (c *Combiner) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// No effects for now
	return nil
}

// GetDescription returns the machine description.
func (c *Combiner) GetDescription() string {
	return "Combines two objects into one with combined value and multipliers."
}

// GetCost returns the cost to place this machine.
func (c *Combiner) GetCost() int {
	return 6
}

// GetName returns the machine name.
func (c *Combiner) GetName() string {
	return "Combiner"
}
