package game

func (g *Game) handleDragAndDrop() {
	cx, cy := g.lastInput.X, g.lastInput.Y

	if g.lastInput.JustPressed {

		// Check rotation buttons first
		if g.state.buttons["rotate_left"].IsClicked(g.lastInput, g.state.phase) {
			selected := g.getSelectedMachine()
			if selected != nil {
				selected.Orientation = (selected.Orientation + 3) % 4
			}
			return
		}
		if g.state.buttons["rotate_right"].IsClicked(g.lastInput, g.state.phase) {
			selected := g.getSelectedMachine()
			if selected != nil {
				selected.Orientation = (selected.Orientation + 1) % 4
			}
			return
		}

		// Check restock button
		if g.state.buttons["restock"].IsClicked(g.lastInput, g.state.phase) {
			if g.state.restocksLeft > 0 {
				selectedIndices := []int{}
				for i, sel := range g.state.inventorySelected {
					if sel {
						selectedIndices = append(selectedIndices, i)
					}
				}
				num := len(selectedIndices)
				if num > 0 {
					// Discard selected
					newInventory := []*MachineState{}
					newSelected := []bool{}
					for i, ms := range g.state.inventory {
						if !g.state.inventorySelected[i] {
							newInventory = append(newInventory, ms)
							newSelected = append(newSelected, false)
						}
					}
					g.state.inventory = newInventory
					g.state.inventorySelected = newSelected
					// Deal num new
					newMachines := dealMachines(g.state.catalogue, num, g.state.runsLeft)
					g.state.inventory = append(g.state.inventory, newMachines...)
					g.state.inventorySelected = append(g.state.inventorySelected, make([]bool, num)...)
					g.state.restocksLeft--
				}
			}
			return
		}

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
		for i := range g.state.inventorySelected {
			g.state.inventorySelected[i] = false
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
			} else {
				// Check if over sell area
				sellX := 10
				sellY := g.bottomY + 10
				sellWidth := 120
				sellHeight := g.bottomHeight - 20
				if cx >= sellX-10 && cx <= sellX+sellWidth+10 && cy >= sellY-10 && cy <= sellY+sellHeight+10 {
					if dragging.IsPlaced && dragging.RunAdded == g.state.runsLeft {
						// Sell the machine
						g.state.money += dragging.Machine.GetCost()
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
	}
}
