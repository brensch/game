package game

func (g *Game) handleDragAndDrop() {
	cx, cy := g.lastInput.X, g.lastInput.Y

	selected := g.getSelectedMachine()
	// Update button visibility and position
	if selected != nil && selected.IsPlaced {
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
		col := selectedPos % gridCols
		row := selectedPos / gridCols
		if row >= 1 && row <= displayRows && col >= 1 && col <= displayCols {
			// Calculate screen position of the selected machine
			machineX := g.gridStartX + (col-1)*(g.cellSize+g.gridMargin)
			machineY := g.gridStartY + (row-1)*(g.cellSize+g.gridMargin)

			// Position buttons below the selected machine, offset from grid alignment
			buttonSize := g.cellSize                             // Make buttons bigger (full cell size instead of half)
			buttonY := machineY + g.cellSize + g.gridMargin + 10 // Extra offset from grid

			// Center the buttons below the machine
			machineCenterX := machineX + g.cellSize/2
			totalButtonWidth := 3*buttonSize + 2*5 // 3 buttons + 2 gaps of 5px
			startX := machineCenterX - totalButtonWidth/2

			// Update rotate left button
			if rotateLeft, exists := g.state.buttons["rotate_left"]; exists {
				rotateLeft.X = startX
				rotateLeft.Y = buttonY
				rotateLeft.Width = buttonSize
				rotateLeft.Height = buttonSize
				rotateLeft.States[PhaseBuild].Visible = true
			}

			// Update rotate right button
			if rotateRight, exists := g.state.buttons["rotate_right"]; exists {
				rotateRight.X = startX + buttonSize + 5
				rotateRight.Y = buttonY
				rotateRight.Width = buttonSize
				rotateRight.Height = buttonSize
				rotateRight.States[PhaseBuild].Visible = true
			}

			// Update sell button
			if sellBtn, exists := g.state.buttons["sell"]; exists {
				sellBtn.X = startX + 2*buttonSize + 2*5
				sellBtn.Y = buttonY
				sellBtn.Width = buttonSize
				sellBtn.Height = buttonSize
				sellBtn.States[PhaseBuild].Visible = true
			}
		} else {
			// Hide
			g.state.buttons["rotate_left"].States[PhaseBuild].Visible = false
			g.state.buttons["rotate_right"].States[PhaseBuild].Visible = false
			g.state.buttons["sell"].States[PhaseBuild].Visible = false
		}
	} else {
		// Hide rotate
		g.state.buttons["rotate_left"].States[PhaseBuild].Visible = false
		g.state.buttons["rotate_right"].States[PhaseBuild].Visible = false
		g.state.buttons["sell"].States[PhaseBuild].Visible = false
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
			// Check if picking from available first
			inventoryClicked := false
			for i, ms := range g.state.inventory {
				row := i / 7
				col := i % 7
				x := g.gridStartX + col*(g.cellSize+g.gridMargin)
				y := g.availableY + row*(g.cellSize+g.gridMargin)
				if cx >= x-10 && cx <= x+g.cellSize+10 && cy >= y-10 && cy <= y+g.cellSize+10 {
					// Deselect all grid machines
					for _, m := range g.state.machines {
						if m != nil {
							m.Selected = false
						}
					}
					g.state.inventorySelected[i] = !g.state.inventorySelected[i]
					ms.Selected = g.state.inventorySelected[i]
					inventoryClicked = true
					break
				}
			}

			if !inventoryClicked {
				// Deselect all inventory
				for i := range g.state.inventorySelected {
					g.state.inventorySelected[i] = false
				}
				for _, ms := range g.state.inventory {
					if ms != nil {
						ms.Selected = false
					}
				}

				// Deselect all grid machines
				for _, m := range g.state.machines {
					if m != nil {
						m.Selected = false
					}
				}

				// Check if picking placed machine
				ms := g.getMachineAt(cx, cy)
				if ms != nil {
					ms.Selected = true
				}
			}
		}

	}

	if g.lastInput.IsDragging {
		selected := g.getSelectedMachine()
		if selected != nil {
			if selected.IsPlaced && selected.RunAdded == g.state.runsLeft {
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
