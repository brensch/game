package game

func (g *Game) handleDragAndDrop() {
	cx, cy := g.lastInput.X, g.lastInput.Y

	selected := g.getSelectedMachine()
	if selected != g.lastSelected {
		// Update button visibility and position
		if selected != nil && selected.IsPlaced && selected.RunAdded == g.state.runsLeft && selected.Machine.GetType() != MachineEnd {
			g.state.buttons["sell"].States[PhaseBuild].Visible = true
		} else {
			g.state.buttons["sell"].States[PhaseBuild].Visible = false
		}
		hasSelectedInventory := false
		for _, sel := range g.state.inventorySelected {
			if sel {
				hasSelectedInventory = true
				break
			}
		}
		if hasSelectedInventory && g.state.restocksLeft > 0 {
			g.state.buttons["restock"].States[PhaseBuild].Visible = true
		} else {
			g.state.buttons["restock"].States[PhaseBuild].Visible = false
		}
		// Position buttons below selected machine
		selectedPos := -1
		if selected != nil {
			for pos, ms := range g.state.machines {
				if ms == selected {
					selectedPos = pos
					break
				}
			}
		}
		if selectedPos != -1 {
			col := selectedPos%gridCols - 1
			row := selectedPos/gridCols - 1
			if row >= 0 && row < displayRows && col >= 0 && col < displayCols {
				buttonY := g.gridStartY + (row+1)*(g.cellSize+g.gridMargin) + g.gridMargin
				buttonX := g.gridStartX + col*(g.cellSize+g.gridMargin) + g.cellSize/2
				// Rotate buttons
				g.state.buttons["rotate_left"].X = buttonX - 45
				g.state.buttons["rotate_left"].Y = buttonY
				g.state.buttons["rotate_right"].X = buttonX + 5
				g.state.buttons["rotate_right"].Y = buttonY
				g.state.buttons["rotate_left"].States[PhaseBuild].Visible = true
				g.state.buttons["rotate_right"].States[PhaseBuild].Visible = true
				// Sell button
				g.state.buttons["sell"].X = buttonX - 40
				g.state.buttons["sell"].Y = buttonY + 35
			} else {
				// Hide
				g.state.buttons["rotate_left"].States[PhaseBuild].Visible = false
				g.state.buttons["rotate_right"].States[PhaseBuild].Visible = false
			}
		} else {
			// Hide rotate
			g.state.buttons["rotate_left"].States[PhaseBuild].Visible = false
			g.state.buttons["rotate_right"].States[PhaseBuild].Visible = false
		}
		g.lastSelected = selected
	}

	if g.lastInput.JustPressed {

		// Check if any button is clicked
		buttonClicked := false
		for _, button := range g.state.buttons {
			if button.IsClicked(g.lastInput, g.state) {
				buttonClicked = true
				break
			}
		}

		if !buttonClicked {
			// Deselect all machines
			for _, m := range g.state.machines {
				if m != nil {
					m.Selected = false
				}
			}

			// Check if picking from available
			for i, ms := range g.state.inventory {
				row := i / 7
				col := i % 7
				x := g.gridStartX + col*(g.cellSize+g.gridMargin)
				y := g.availableY + row*(g.cellSize+g.gridMargin)
				if cx >= x-10 && cx <= x+g.cellSize+10 && cy >= y-10 && cy <= y+g.cellSize+10 {
					g.state.inventorySelected[i] = !g.state.inventorySelected[i]
					ms.Selected = g.state.inventorySelected[i]
					break
				}
			}

			// Check if picking placed machine
			ms := g.getMachineAt(cx, cy)
			if ms != nil {
				ms.Selected = true
			}
		}

	}

	if g.lastInput.IsDragging {
		selected := g.getSelectedMachine()
		if selected != nil {
			if selected.IsPlaced && selected.Machine.GetType() != MachineEnd && selected.RunAdded == g.state.runsLeft {
				selected.BeingDragged = true
				pos := g.getPos(selected)
				selected.OriginalPos = pos
			} else if !selected.IsPlaced {
				// from available
				selected.BeingDragged = true
				selected.RunAdded = g.state.runsLeft
			}
		}
	}

	if g.lastInput.JustReleased {
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
					cost := dragging.Machine.GetCost()
					if g.state.money >= cost {
						g.state.money -= cost
						// Create a new instance for placed machine
						newMS := &MachineState{
							Machine:      dragging.Machine,
							Orientation:  dragging.Orientation,
							BeingDragged: false,
							IsPlaced:     true,
							RunAdded:     g.state.runsLeft,
						}
						position := (gridY+1)*gridCols + (gridX + 1)
						g.state.machines[position] = newMS
						placedMS = newMS
						// Remove from inventory
						for i, ms := range g.state.inventory {
							if ms == dragging {
								g.state.inventory = append(g.state.inventory[:i], g.state.inventory[i+1:]...)
								g.state.inventorySelected = append(g.state.inventorySelected[:i], g.state.inventorySelected[i+1:]...)
								break
							}
						}
					} else {
						// Not enough money, don't place
						dragging.BeingDragged = false
						return
					}
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
				for _, m := range g.state.inventory {
					if m != nil {
						m.Selected = false
					}
				}
				// Select the placed one
				placedMS.Selected = true
			}
			dragging.BeingDragged = false
		}
	}
}
