package game

import (
	"image/color"
)

// Amplifier represents an amplifier machine.
type Amplifier struct{}

// GetType returns the machine type.
func (a *Amplifier) GetType() MachineType {
	return MachineAmplifier
}

// GetRoles returns the machine roles.
func (a *Amplifier) GetRoles() []MachineRole {
	return []MachineRole{RoleMover, RoleUpgrader}
}

// GetRoleNames returns the names of the machine roles.
func (a *Amplifier) GetRoleNames() []string {
	return []string{"Mover", "Upgrader"}
}

// GetColor returns the machine color.
func (a *Amplifier) GetColor() color.RGBA {
	return color.RGBA{R: 255, G: 215, B: 0, A: 255} // Gold
}

// Process handles object interaction for amplifier.
func (a *Amplifier) Process(position int, history [][]*Object, tick int, orientation Orientation) []*Change {
	current := history[len(history)-1]
	for _, obj := range current {
		if obj.GridPosition == position {
			nextPos := GetAdjacentPosition(position, orientation)
			newValue := obj.Score.Value * 2 // Double the value
			return []*Change{{
				StartObject: obj,
				EndObject:   &Object{GridPosition: nextPos, Type: obj.Type, Score: &Score{Value: newValue, MultAdd: obj.Score.MultAdd, MultMult: obj.Score.MultMult}},
				Score:       nil,
			}}
		}
	}
	return nil
}

// EmitEffects emits effects from amplifier.
func (a *Amplifier) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// Emit amplify effect to adjacent producers
	var emissions []EffectEmission
	pos := state.OriginalPos
	row := pos / gridCols
	col := pos % gridCols
	directions := []struct{ dx, dy int }{{0, 1}, {1, 0}, {0, -1}, {-1, 0}}
	for _, dir := range directions {
		nr, nc := row+dir.dx, col+dir.dy
		if nr >= 0 && nr < gridRows && nc >= 0 && nc < gridCols {
			npos := nr*gridCols + nc
			if machineState := game.state.machines[npos]; machineState != nil {
				for _, role := range machineState.Machine.GetRoles() {
					if role == RoleProducer {
						emissions = append(emissions, EffectEmission{
							TargetGridX: nc,
							TargetGridY: nr,
							Effect: &Effect{
								Type:         EffectAmplifyValue,
								Duration:     1,
								DurationType: DurationTick,
							},
						})
					}
				}
			}
		}
	}
	return emissions
}

// GetDescription returns the machine description.
func (a *Amplifier) GetDescription() string {
	return "Doubles the value of objects passing through and boosts nearby producers."
}

// GetName returns the machine name.
func (a *Amplifier) GetName() string {
	return "Amplifier"
}
