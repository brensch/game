package game

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func getMachineName(mt MachineType) string {
	switch mt {
	case MachineConveyor:
		return "Conveyor"
	case MachineProcessor:
		return "Processor"
	case MachineMiner:
		return "Miner"
	case MachineEnd:
		return "End"
	case MachineSplitter:
		return "Splitter"
	default:
		return "Unknown"
	}
}

func wrapText(text string, maxLen int) []string {
	words := strings.Fields(text)
	var lines []string
	current := ""
	for _, word := range words {
		if len(current)+len(word)+1 > maxLen {
			if current != "" {
				lines = append(lines, current)
				current = word
			} else {
				lines = append(lines, word)
				current = ""
			}
		} else {
			if current != "" {
				current += " " + word
			} else {
				current = word
			}
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func (g *Game) drawScanlines(screen *ebiten.Image) {
	bounds := screen.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	for y := 0; y < h; y += 2 {
		vector.DrawFilledRect(screen, 0, float32(y), float32(w), 1, color.RGBA{R: 0, G: 0, B: 0, A: 30}, false)
	}
}

// func (g *Game) drawUI(screen *ebiten.Image) {
// 	// Top panel - Total Score, Money, Round Target, Run, Round
// 	vector.DrawFilledRect(screen, 0, float32(g.topPanelY), float32(g.screenWidth), float32(g.topPanelHeight), color.RGBA{R: 80, G: 80, B: 80, A: 255}, false)
// 	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Total Score: %d", g.state.totalScore), 20, g.topPanelY+10)
// 	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Money: $%d", g.state.money), 20, g.topPanelY+30)
// 	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Round Target: %d", g.state.targetScore), 200, g.topPanelY+10)
// 	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Runs Left: %d", g.state.runsLeft), 200, g.topPanelY+30)
// 	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Round: %d", g.state.round), 200, g.topPanelY+50)

// 	// Bottom Panel
// 	vector.DrawFilledRect(screen, 0, float32(g.bottomY), float32(g.screenWidth), float32(g.bottomHeight), color.RGBA{R: 80, G: 80, B: 80, A: 255}, false)

// 	// Current Run Score (centered in the middle)
// 	scoreText := fmt.Sprintf("Run Score: %d x %d", g.state.roundScore, g.state.multiplier)
// 	scoreX := (g.screenWidth - len(scoreText)*6) / 2 // Approximate centering, assuming ~6px per char
// 	ebitenutil.DebugPrintAt(screen, scoreText, scoreX, g.bottomY+20)

// 	// Render all buttons
// 	for _, button := range g.state.buttons {
// 		button.Render(screen, g.state)
// 	}
// }

func (g *Game) drawFactoryFloor(screen *ebiten.Image) {
	for r := 0; r < displayRows; r++ {
		for c := 0; c < displayCols; c++ {
			x := g.gridStartX + c*(g.cellSize+g.gridMargin)
			y := g.gridStartY + r*(g.cellSize+g.gridMargin)
			vector.DrawFilledRect(screen, float32(x), float32(y), float32(g.cellSize), float32(g.cellSize), color.RGBA{R: 60, G: 60, B: 60, A: 255}, false)
		}
	}
}

func (g *Game) drawArrow(screen *ebiten.Image, x, y float32, orientation Orientation) {
	arrowColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	arrowSize := float32(g.cellSize / 6)
	centerX := x + float32(g.cellSize)/2
	shaftStartY := y + arrowSize
	shaftEndY := y + float32(g.cellSize) - arrowSize
	arrowY := y + 2*arrowSize
	leftX := centerX - arrowSize
	rightX := centerX + arrowSize
	shaftY := y + float32(g.cellSize)/2
	shaftLeft := x + arrowSize
	shaftRight := x + float32(g.cellSize) - arrowSize
	arrowX := x + float32(g.cellSize) - 2*arrowSize
	topY := shaftY - arrowSize
	bottomY := shaftY + arrowSize
	switch orientation {
	case OrientationNorth:
		vector.StrokeLine(screen, centerX, shaftEndY, centerX, shaftStartY, 1, arrowColor, false)
		vector.StrokeLine(screen, leftX, arrowY, centerX, shaftStartY, 1, arrowColor, false)
		vector.StrokeLine(screen, rightX, arrowY, centerX, shaftStartY, 1, arrowColor, false)
	case OrientationEast:
		vector.StrokeLine(screen, shaftLeft, shaftY, shaftRight, shaftY, 1, arrowColor, false)
		vector.StrokeLine(screen, arrowX, topY, shaftRight, shaftY, 1, arrowColor, false)
		vector.StrokeLine(screen, arrowX, bottomY, shaftRight, shaftY, 1, arrowColor, false)
	case OrientationSouth:
		vector.StrokeLine(screen, centerX, shaftStartY, centerX, shaftEndY, 1, arrowColor, false)
		vector.StrokeLine(screen, leftX, y+float32(g.cellSize)-2*arrowSize, centerX, shaftEndY, 1, arrowColor, false)
		vector.StrokeLine(screen, rightX, y+float32(g.cellSize)-2*arrowSize, centerX, shaftEndY, 1, arrowColor, false)
	case OrientationWest:
		vector.StrokeLine(screen, shaftRight, shaftY, shaftLeft, shaftY, 1, arrowColor, false)
		vector.StrokeLine(screen, x+2*arrowSize, topY, shaftLeft, shaftY, 1, arrowColor, false)
		vector.StrokeLine(screen, x+2*arrowSize, bottomY, shaftLeft, shaftY, 1, arrowColor, false)
	}
}

func (g *Game) drawRotateArrow(screen *ebiten.Image, x, y, width, height int, left bool) {
	arrowColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	margin := float32(4)
	lineLength := float32(width) * 0.25
	centerX := float32(x) + float32(width)/2
	centerY := float32(y) + float32(height)/2

	if left {
		// Counterclockwise: L shape centered, arrow pointing right
		vertTop := centerY - lineLength/2
		vertBottom := centerY + lineLength/2

		// Vertical line
		vector.StrokeLine(screen, centerX, vertTop, centerX, vertBottom, 3, arrowColor, false)

		// Horizontal line from bottom to right
		horizRight := centerX + lineLength
		vector.StrokeLine(screen, centerX, vertBottom, horizRight, vertBottom, 3, arrowColor, false)

		// Arrowhead at the right end: two lines diagonally backwards from tip
		arrowTipX := horizRight
		vector.StrokeLine(screen, arrowTipX, vertBottom, arrowTipX-margin, vertBottom-margin, 3, arrowColor, false)
		vector.StrokeLine(screen, arrowTipX, vertBottom, arrowTipX-margin, vertBottom+margin, 3, arrowColor, false)
	} else {
		// Clockwise: L shape centered, arrow pointing left
		vertTop := centerY - lineLength/2
		vertBottom := centerY + lineLength/2

		// Vertical line
		vector.StrokeLine(screen, centerX, vertTop, centerX, vertBottom, 3, arrowColor, false)

		// Horizontal line from bottom to left
		horizLeft := centerX - lineLength
		vector.StrokeLine(screen, centerX, vertBottom, horizLeft, vertBottom, 3, arrowColor, false)

		// Arrowhead at the left end: two lines diagonally backwards from tip
		arrowTipX := horizLeft
		vector.StrokeLine(screen, arrowTipX, vertBottom, arrowTipX+margin, vertBottom-margin, 3, arrowColor, false)
		vector.StrokeLine(screen, arrowTipX, vertBottom, arrowTipX+margin, vertBottom+margin, 3, arrowColor, false)
	}
}

// func (g *Game) drawMachines(screen *ebiten.Image) {
// 	// Machines on the grid
// 	for pos, ms := range g.state.machines {
// 		if ms == nil || ms.Machine == nil || ms.BeingDragged {
// 			continue
// 		}
// 		col := pos % gridCols
// 		row := pos / gridCols
// 		if row < 1 || row > displayRows || col < 1 || col > displayCols {
// 			continue
// 		}
// 		x := g.gridStartX + (col-1)*(g.cellSize+g.gridMargin)
// 		y := g.gridStartY + (row-1)*(g.cellSize+g.gridMargin)
// 		vector.DrawFilledRect(screen, float32(x), float32(y), float32(g.cellSize), float32(g.cellSize), ms.Machine.GetColor(), false)

// 		if ms.Machine.GetType() == MachineEnd {
// 			ebitenutil.DebugPrintAt(screen, "End", int(x)+15, int(y)+20)
// 		}
// 		g.drawArrow(screen, float32(x), float32(y), ms.Orientation)
// 		if ms.Selected {
// 			vector.StrokeRect(screen, float32(x), float32(y), float32(g.cellSize), float32(g.cellSize), 3, color.RGBA{R: 255, G: 255, B: 0, A: 255}, false)
// 		}
// 	}

// 	// Available machines
// 	for i, ms := range g.state.inventory {
// 		if ms != nil && !ms.BeingDragged && ms.Machine != nil {
// 			row := i / 7
// 			col := i % 7
// 			x := g.gridStartX + col*(g.cellSize+g.gridMargin)
// 			y := g.availableY + row*(g.cellSize+g.gridMargin)
// 			vector.DrawFilledRect(screen, float32(x), float32(y), float32(g.cellSize), float32(g.cellSize), ms.Machine.GetColor(), false)
// 			if g.state.inventorySelected[i] {
// 				vector.StrokeRect(screen, float32(x), float32(y), float32(g.cellSize), float32(g.cellSize), 3, color.RGBA{R: 255, G: 255, B: 0, A: 255}, false)
// 			}
// 		}
// 	}

// 	// // Rotation buttons
// 	// g.state.buttons["rotate_left"].Render(screen, g.state)
// 	// g.state.buttons["rotate_right"].Render(screen, g.state)
// }

// func (g *Game) drawObjects(screen *ebiten.Image) {
// 	for _, obj := range g.state.objects {
// 		objColor := color.RGBA{R: 255, A: 255}
// 		switch obj.Type {
// 		case ObjectGreen:
// 			objColor.G = 255
// 		case ObjectBlue:
// 			objColor.B = 255
// 		}
// 		gridX := obj.GridPosition % gridCols
// 		gridY := obj.GridPosition / gridCols
// 		if gridY < 1 || gridY > displayRows || gridX < 1 || gridX > displayCols {
// 			continue
// 		}
// 		x := g.gridStartX + (gridX-1)*(g.cellSize+g.gridMargin) + g.cellSize/2
// 		y := g.gridStartY + (gridY-1)*(g.cellSize+g.gridMargin) + g.cellSize/2
// 		vector.DrawFilledCircle(screen, float32(x), float32(y), 10, objColor, false)
// 	}

// 	// Draw animations
// 	for _, anim := range g.state.animations {
// 		progress := anim.Elapsed / anim.Duration
// 		if progress > 1 {
// 			progress = 1
// 		}
// 		x := anim.StartX + (anim.EndX-anim.StartX)*progress
// 		y := anim.StartY + (anim.EndY-anim.StartY)*progress
// 		vector.DrawFilledCircle(screen, float32(x), float32(y), 10, anim.Color, false)
// 	}
// }

func (g *Game) drawTooltip(screen *ebiten.Image) {
	var tooltipMachine MachineInterface
	var tooltipX, tooltipY int
	var isHover bool

	// Use current cursor position for hover detection
	cx, cy := ebiten.CursorPosition()

	// Check for hover over grid machines
	for pos := 0; pos < gridCols*gridRows; pos++ {
		ms := g.state.machines[pos]
		if ms != nil && !ms.BeingDragged && ms.Machine != nil {
			col := pos % gridCols
			row := pos / gridCols
			if row >= 1 && row <= displayRows && col >= 1 && col <= displayCols {
				x := g.gridStartX + (col-1)*(g.cellSize+g.gridMargin)
				y := g.gridStartY + (row-1)*(g.cellSize+g.gridMargin)
				if cx >= x-15 && cx <= x+g.cellSize+15 && cy >= y-15 && cy <= y+g.cellSize+15 {
					tooltipMachine = ms.Machine
					tooltipX = x + g.cellSize/2 - 200
					tooltipY = y - 80
					isHover = true
					break
				}
			}
		}
	}

	// If not hovering over grid, check inventory
	if tooltipMachine == nil {
		for i, ms := range g.state.inventory {
			if ms != nil && !ms.BeingDragged && ms.Machine != nil {
				row := i / 7
				col := i % 7
				x := g.gridStartX + col*(g.cellSize+g.gridMargin)
				y := g.availableY + row*(g.cellSize+g.gridMargin)
				if cx >= x-15 && cx <= x+g.cellSize+15 && cy >= y-15 && cy <= y+g.cellSize+15 {
					tooltipMachine = ms.Machine
					tooltipX = x + g.cellSize/2 - 200
					tooltipY = y - 80
					isHover = true
					break
				}
			}
		}
	}

	// If not hovering, show for selected machine
	if tooltipMachine == nil {
		selected := g.getSelectedMachine()
		if selected != nil && selected.Machine != nil {
			tooltipMachine = selected.Machine
			if selected.IsPlaced {
				// Grid machine
				for pos, ms := range g.state.machines {
					if ms == selected {
						col := pos % gridCols
						row := pos / gridCols
						tooltipX = g.gridStartX + (col-1)*(g.cellSize+g.gridMargin) + g.cellSize/2 - 200
						tooltipY = g.gridStartY + (row-1)*(g.cellSize+g.gridMargin) - 80
						break
					}
				}
			} else {
				// Inventory machine
				for i, ms := range g.state.inventory {
					if ms == selected {
						row := i / 7
						col := i % 7
						tooltipX = g.gridStartX + col*(g.cellSize+g.gridMargin) + g.cellSize/2 - 200
						tooltipY = g.availableY + row*(g.cellSize+g.gridMargin) - 80
						break
					}
				}
			}
			isHover = false
		}
	}

	if tooltipMachine != nil {
		name := tooltipMachine.GetName()
		description := tooltipMachine.GetDescription()
		cost := tooltipMachine.GetCost()
		roles := tooltipMachine.GetRoles()
		lines := wrapText(description, 60)

		// Build roles string
		var rolesStr string
		if len(roles) > 0 {
			rolesStr = "Roles: "
			for i, role := range roles {
				if i > 0 {
					rolesStr += ", "
				}
				rolesStr += getMachineRoleName(role)
			}
		}

		// Calculate height
		nameHeight := 15
		lineHeight := 15
		rolesHeight := 15
		costHeight := 15
		totalHeight := 20 + nameHeight + len(lines)*lineHeight + rolesHeight + costHeight

		// Ensure tooltip stays on screen
		if tooltipX < 5 {
			tooltipX = 5
		}
		if tooltipX > g.screenWidth-405 {
			tooltipX = g.screenWidth - 405
		}
		if tooltipY < 5 {
			tooltipY = 5
		}
		if tooltipY > g.height-totalHeight-5 {
			tooltipY = g.height - totalHeight - 5
		}

		// Draw tooltip background
		var bgColor color.RGBA
		if isHover {
			bgColor = color.RGBA{R: 255, G: 255, B: 0, A: 255} // Yellow for hover
		} else {
			bgColor = color.RGBA{R: 255, G: 255, B: 255, A: 255} // White for selected
		}
		vector.DrawFilledRect(screen, float32(tooltipX-5), float32(tooltipY-5), 400, float32(totalHeight), bgColor, false)
		vector.StrokeRect(screen, float32(tooltipX-5), float32(tooltipY-5), 400, float32(totalHeight), 1, color.RGBA{R: 0, G: 0, B: 0, A: 255}, false)

		// Draw tooltip text
		y := tooltipY + 10
		text.Draw(screen, name, g.font, tooltipX, y, color.Black)
		y += nameHeight
		for _, line := range lines {
			text.Draw(screen, line, g.font, tooltipX, y, color.Black)
			y += lineHeight
		}
		if rolesStr != "" {
			text.Draw(screen, rolesStr, g.font, tooltipX, y, color.Black)
			y += rolesHeight
		}
		text.Draw(screen, fmt.Sprintf("Cost: $%d", cost), g.font, tooltipX, y, color.Black)
	}
}
