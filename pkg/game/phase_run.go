package game

import (
	"image/color"
	"math/rand"
)

func (g *Game) handleRunPhase() {
	if g.state.allChanges == nil {
		// Still calculating
		return
	}

	// Update animations
	for _, anim := range g.state.animations {
		anim.Elapsed++
	}
	// Remove completed animations
	newAnims := []*Animation{}
	for _, anim := range g.state.animations {
		if anim.Elapsed < anim.Duration {
			newAnims = append(newAnims, anim)
		}
	}
	g.state.animations = newAnims

	// Now check if need to start new tick or switch
	if len(g.state.animations) == 0 {
		changes := g.state.allChanges
		if g.state.animationTick >= len(changes) || len(changes) == 0 {
			if g.state.endRunDelay == 0 {
				g.state.endRunDelay = 30
			}
		} else {
			// Start new tick
			tickChanges := changes[g.state.animationTick]
			g.state.animations = []*Animation{}
			// Accumulate scores
			for _, ch := range tickChanges {
				if ch.Score != nil {
					g.state.roundScore += ch.Score.Value
					g.state.multiplier += ch.Score.MultAdd
					g.state.multiplier *= ch.Score.MultMult
				}
			}
			for _, ch := range tickChanges {
				if ch.StartObject == nil || ch.EndObject == nil {
					continue
				}
				startGridX := ch.StartObject.GridPosition % gridCols
				startGridY := ch.StartObject.GridPosition / gridCols
				endGridX := ch.EndObject.GridPosition % gridCols
				endGridY := ch.EndObject.GridPosition / gridCols
				startX := float64(g.gridStartX + (startGridX-1)*(g.cellSize+g.gridMargin) + g.cellSize/2)
				startY := float64(g.gridStartY + (startGridY-1)*(g.cellSize+g.gridMargin) + g.cellSize/2)
				endX := float64(g.gridStartX + (endGridX-1)*(g.cellSize+g.gridMargin) + g.cellSize/2)
				endY := float64(g.gridStartY + (endGridY-1)*(g.cellSize+g.gridMargin) + g.cellSize/2)
				objColor := color.RGBA{R: 255, A: 255}
				switch ch.StartObject.Type {
				case ObjectGreen:
					objColor.G = 255
				case ObjectBlue:
					objColor.B = 255
				}
				duration := 30.0 / g.state.animationSpeed // frames, decrease over time
				g.state.animations = append(g.state.animations, &Animation{
					StartX: startX, StartY: startY,
					EndX: endX, EndY: endY,
					Color: objColor, Duration: duration, Elapsed: 0,
				})
			}
			g.state.animationTick++
			g.state.animationSpeed += 0.3 // speed up significantly each tick
		}
	}

	if g.state.endRunDelay > 0 {
		g.state.endRunDelay--
		if g.state.endRunDelay == 0 {
			// End the run
			g.state.animationTick = 0
			g.state.animationSpeed = 1.0
			g.state.allChanges = nil
			g.state.runsLeft--
			// Move end to random location up to 2 squares away
			for pos, ms := range g.state.machines {
				if ms != nil && ms.Machine.GetType() == MachineEnd {
					currentPos := pos
					var candidates []int
					cr := currentPos / gridCols
					cc := currentPos % gridCols
					for dr := -2; dr <= 2; dr++ {
						for dc := -2; dc <= 2; dc++ {
							if abs(dr)+abs(dc) > 2 || (dr == 0 && dc == 0) {
								continue
							}
							nr := cr + dr
							nc := cc + dc
							if nr >= 1 && nr <= displayRows && nc >= 1 && nc <= displayCols {
								npos := nr*gridCols + nc
								if g.state.machines[npos] == nil {
									candidates = append(candidates, npos)
								}
							}
						}
					}
					if len(candidates) > 0 {
						newPos := candidates[rand.Intn(len(candidates))]
						g.state.machines[newPos] = ms
						g.state.machines[currentPos] = nil
					}
					break
				}
			}
			// Add run score to total
			g.state.totalScore += g.state.roundScore * g.state.multiplier
			g.state.roundScore = 0
			g.state.multiplier = 1
			if g.state.runsLeft > 0 {
				numToDeal := g.state.inventorySize - len(g.state.inventory)
				if numToDeal > 0 {
					newMachines := dealMachines(g.state.catalogue, numToDeal, g.state.runsLeft)
					g.state.inventory = append(g.state.inventory, newMachines...)
					g.state.inventorySelected = append(g.state.inventorySelected, make([]bool, numToDeal)...)
				}
			}
			if g.state.runsLeft == 0 {
				if g.state.totalScore >= g.state.targetScore {
					g.state.phase = PhaseRoundEnd
				} else {
					g.state.gameOver = true
					g.state.phase = PhaseGameOver
				}
			} else {
				g.state.phase = PhaseBuild
			}
		}
		return
	}
}
