package game

import (
	"image/color"
)

// Booster represents a booster machine.
type Booster struct{}

// GetType returns the machine type.
func (b *Booster) GetType() MachineType {
	return MachineBooster
}

// GetRoles returns the machine roles.
func (b *Booster) GetRoles() []MachineRole {
	return []MachineRole{RoleMover}
}

// GetRoleNames returns the names of the machine roles.
func (b *Booster) GetRoleNames() []string {
	return []string{"Mover"}
}

// GetColor returns the machine color.
func (b *Booster) GetColor() color.RGBA {
	return color.RGBA{R: 0, B: 255, G: 255, A: 255} // Cyan
}

// Process handles object interaction for booster.
func (b *Booster) Process(position int, history [][]*Object, tick int, orientation Orientation) []*Change {
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

// EmitEffects emits effects from booster.
func (b *Booster) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// Emit speed buff to adjacent machines
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
				emissions = append(emissions, EffectEmission{
					TargetGridX: nc,
					TargetGridY: nr,
					Effect: &Effect{
						Type:         EffectBuffSpeed,
						Duration:     5,
						DurationType: DurationTick,
					},
				})
			}
		}
	}
	return emissions
}

// GetDescription returns the machine description.
func (b *Booster) GetDescription() string {
	return "Moves objects forward and boosts the speed of adjacent machines."
}

// GetCost returns the cost to place this machine.
func (b *Booster) GetCost() int {
	return 4
}

// GetName returns the machine name.
func (b *Booster) GetName() string {
	return "Booster"
}