package game

func (g *Game) handleDragAndDrop() {
	cx, cy := g.lastInput.X, g.lastInput.Y

	if g.lastInput.JustPressed {
		g.state.inputPressed = true
		g.state.pressX = cx
		g.state.pressY = cy

		// Check rotation buttons first
		if g.state.buttons["rotate_left"].IsClicked(g.lastInput) {
			selected := g.getSelectedMachine()
			if selected != nil {
				selected.Orientation = (selected.Orientation + 3) % 4
			}
			return
		}
		if g.state.buttons["rotate_right"].IsClicked(g.lastInput) {
			selected := g.getSelectedMachine()
			if selected != nil {
				selected.Orientation = (selected.Orientation + 1) % 4
			}
			return
		}

		// Deselect all
		for _, m := range g.state.machines {
			if m != nil {
				m.Selected = false
			}
		}
		for _, m := range g.state.availableMachines {
			if m != nil {
				m.Selected = false
			}
		}

		// Check if picking from available
		for i, ms := range g.state.availableMachines {
			x := g.gridStartX + i*(g.cellSize+g.gridMargin)
			y := g.availableY
			if cx >= x-10 && cx <= x+g.cellSize+10 && cy >= y-10 && cy <= y+g.cellSize+10 {
				ms.Selected = true
				break
			}
		}

		// Check if picking placed machine
		ms := g.getMachineAt(cx, cy)
		if ms != nil {
			ms.Selected = true
		}

	}

	if g.state.inputPressed && g.lastInput.Pressed {
		dx := cx - g.state.pressX
		dy := cy - g.state.pressY
		if dx*dx+dy*dy > 1000 { // threshold
			selected := g.getSelectedMachine()
			if selected != nil {
				if selected.IsPlaced && selected.Machine.GetType() != MachineEnd {
					selected.BeingDragged = true
					pos := g.getPos(selected)
					selected.OriginalPos = pos
				} else if !selected.IsPlaced {
					// from available
					selected.BeingDragged = true
					selected.RunAdded = g.state.run
				}
			}
		}
	}

	if g.lastInput.JustReleased {
		g.state.inputPressed = false
		dragging := g.getDraggingMachine()
		if dragging != nil {
			// Place at cursor position
			gridX, gridY := -1, -1
			for r := 0; r < displayRows; r++ {
				for c := 0; c < displayCols; c++ {
					x := g.gridStartX + c*(g.cellSize+g.gridMargin)
					y := g.gridStartY + r*(g.cellSize+g.gridMargin)
					if cx >= x-10 && cx <= x+g.cellSize+10 && cy >= y-10 && cy <= y+g.cellSize+10 {
						position := (r+1)*gridCols + (c + 1)
						if g.state.machines[position] == nil {
							gridX, gridY = c, r
						}
						break
					}
				}
				if gridX != -1 {
					break
				}
			}
			if gridX != -1 {
				var placedMS *MachineState
				if !dragging.IsPlaced {
					g.state.money -= 1
					// Create a new instance for placed machine
					newMS := &MachineState{
						Machine:      dragging.Machine,
						Orientation:  dragging.Orientation,
						BeingDragged: false,
						IsPlaced:     true,
						RunAdded:     g.state.run,
					}
					position := (gridY+1)*gridCols + (gridX + 1)
					g.state.machines[position] = newMS
					placedMS = newMS
				} else {
					// Moving existing placed machine
					position := (gridY+1)*gridCols + (gridX + 1)
					g.state.machines[position] = dragging
					if position != dragging.OriginalPos {
						g.state.machines[dragging.OriginalPos] = nil
					}
					placedMS = dragging
				}
				// Select the newly placed machine
				// Deselect all
				for _, m := range g.state.machines {
					if m != nil {
						m.Selected = false
					}
				}
				for _, m := range g.state.availableMachines {
					if m != nil {
						m.Selected = false
					}
				}
				// Select the placed one
				placedMS.Selected = true
			} else {
				// Check if over sell area
				sellX := 10
				sellY := g.bottomY + 10
				sellWidth := 120
				sellHeight := g.bottomHeight - 20
				if cx >= sellX-10 && cx <= sellX+sellWidth+10 && cy >= sellY-10 && cy <= sellY+sellHeight+10 {
					if dragging.IsPlaced && dragging.Machine.GetType() != MachineMiner && dragging.Machine.GetType() != MachineEnd {
						// Sell the machine
						g.state.money += 1
						// Remove from grid
						for pos, ms := range g.state.machines {
							if ms == dragging {
								g.state.machines[pos] = nil
								break
							}
						}
					}
				}
			}
			dragging.BeingDragged = false
		}

		g.state.inputPressed = false
	}
}
