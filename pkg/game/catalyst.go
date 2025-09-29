package game

import (
	"image/color"
)

// Catalyst represents a catalyst machine.
type Catalyst struct{}

// GetType returns the machine type.
func (c *Catalyst) GetType() MachineType {
	return MachineCatalyst
}

// GetRoles returns the machine roles.
func (c *Catalyst) GetRoles() []MachineRole {
	return []MachineRole{RoleMover}
}

// GetRoleNames returns the names of the machine roles.
func (c *Catalyst) GetRoleNames() []string {
	return []string{"Mover"}
}

// GetColor returns the machine color.
func (c *Catalyst) GetColor() color.RGBA {
	return color.RGBA{R: 255, G: 165, B: 0, A: 255} // Orange
}

// Process handles object interaction for catalyst.
func (c *Catalyst) Process(position int, history [][]*Object, tick int, orientation Orientation) []*Change {
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

// EmitEffects emits effects from catalyst.
func (c *Catalyst) EmitEffects(game *Game, state *MachineState) []EffectEmission {
	// Emit efficiency buff to adjacent machines
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
						Type:         EffectBuffEfficiency,
						Duration:     10,
						DurationType: DurationTick,
					},
				})
			}
		}
	}
	return emissions
}

// GetDescription returns the machine description.
func (c *Catalyst) GetDescription() string {
	return "Moves objects forward and increases efficiency of adjacent machines."
}

// GetCost returns the cost to place this machine.
func (c *Catalyst) GetCost() int {
	return 5
}

// GetName returns the machine name.
func (c *Catalyst) GetName() string {
	return "Catalyst"
}