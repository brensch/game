package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) drawScanlines(screen *ebiten.Image) {
	bounds := screen.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	for y := 0; y < h; y += 2 {
		vector.DrawFilledRect(screen, 0, float32(y), float32(w), 1, color.RGBA{R: 0, G: 0, B: 0, A: 30}, false)
	}
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Top panel - Total Score
	vector.DrawFilledRect(screen, 10, float32(g.topPanelY), float32(g.screenWidth-20), float32(g.topPanelHeight), color.RGBA{R: 80, G: 80, B: 80, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Total Score: %d x %d = %d", g.state.baseScore, g.state.multiplier, g.state.baseScore*g.state.multiplier), 20, g.topPanelY+20)

	// Restart button
	vector.DrawFilledRect(screen, float32(g.screenWidth-100), float32(g.topPanelY+10), 80, float32(g.topPanelHeight-20), color.RGBA{R: 200, G: 100, B: 100, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, "Restart", g.screenWidth-90, g.topPanelY+30)

	// Foreman panel - Money and Run
	vector.DrawFilledRect(screen, 10, float32(g.foremanY), float32(g.screenWidth-20), float32(g.foremanHeight), color.RGBA{R: 100, G: 100, B: 100, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Money: $%d", g.state.money), 20, g.foremanY+20)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Run: %d/%d", g.state.run, g.state.maxRuns), 200, g.foremanY+20)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Round: %d", g.state.round), 200, g.foremanY+50)

	// Bottom Panel
	vector.DrawFilledRect(screen, 10, float32(g.bottomY), float32(g.screenWidth-20), float32(g.bottomHeight), color.RGBA{R: 80, G: 80, B: 80, A: 255}, false)

	// Sell Button (100px)
	vector.DrawFilledRect(screen, 10, float32(g.bottomY+10), buttonWidth, float32(g.bottomHeight-20), color.RGBA{R: 255, G: 100, B: 100, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, "Sell", 20, g.bottomY+20)

	// Current Round Score (centered in the middle)
	scoreText := fmt.Sprintf("Round Score: %d", g.state.baseScore)
	scoreX := (g.screenWidth - len(scoreText)*6) / 2  // Approximate centering, assuming ~6px per char
	ebitenutil.DebugPrintAt(screen, scoreText, scoreX, g.bottomY+20)

	// Start/Stop Run Button (100px on the right)
	runButtonX := float32(g.screenWidth - 10 - buttonWidth)
	runButtonColor := color.RGBA{R: 100, G: 200, B: 100, A: 255}
	runButtonText := "Start Run"
	if g.state.running {
		runButtonColor = color.RGBA{R: 200, G: 200, B: 100, A: 255}
		runButtonText = "Running"
	}
	vector.DrawFilledRect(screen, runButtonX, float32(g.bottomY+10), buttonWidth, float32(g.bottomHeight-20), runButtonColor, false)
	ebitenutil.DebugPrintAt(screen, runButtonText, int(runButtonX)+5, g.bottomY+20)
}

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

func (g *Game) drawMachines(screen *ebiten.Image) {
	// Machines on the grid
	for pos, ms := range g.state.machines {
		if ms == nil || ms.Machine == nil || ms.BeingDragged {
			continue
		}
		col := pos % gridCols
		row := pos / gridCols
		if row < 1 || row > displayRows || col < 1 || col > displayCols {
			continue
		}
		x := g.gridStartX + (col-1)*(g.cellSize+g.gridMargin)
		y := g.gridStartY + (row-1)*(g.cellSize+g.gridMargin)
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(g.cellSize), float32(g.cellSize), ms.Machine.GetColor(), false)

		if ms.Machine.GetType() == MachineEnd {
			ebitenutil.DebugPrintAt(screen, "End", int(x)+15, int(y)+20)
		}
		g.drawArrow(screen, float32(x), float32(y), ms.Orientation)
		if ms.Selected {
			vector.StrokeRect(screen, float32(x), float32(y), float32(g.cellSize), float32(g.cellSize), 3, color.RGBA{R: 255, G: 255, B: 0, A: 255}, false)
		}
	}

	// Available machines
	for i, ms := range g.state.availableMachines {
		if ms != nil && !ms.BeingDragged && ms.Machine != nil {
			x := g.gridStartX + i*(g.cellSize+g.gridMargin)
			y := g.availableY
			vector.DrawFilledRect(screen, float32(x), float32(y), float32(g.cellSize), float32(g.cellSize), ms.Machine.GetColor(), false)
		}
	}

	// Rotation buttons
	counterclockwiseX := g.screenWidth - 2*g.cellSize - g.gridMargin
	counterclockwiseY := g.availableY
	vector.DrawFilledCircle(screen, float32(counterclockwiseX+g.cellSize/2), float32(counterclockwiseY+g.cellSize/2), float32(g.cellSize/2), color.RGBA{R: 200, G: 100, B: 100, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, "<-", counterclockwiseX+22, counterclockwiseY+26)

	clockwiseX := g.screenWidth - g.cellSize
	clockwiseY := g.availableY
	vector.DrawFilledCircle(screen, float32(clockwiseX+g.cellSize/2), float32(clockwiseY+g.cellSize/2), float32(g.cellSize/2), color.RGBA{R: 100, G: 100, B: 200, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, "->", clockwiseX+22, clockwiseY+26)
}

func (g *Game) drawObjects(screen *ebiten.Image) {
	for _, obj := range g.state.objects {
		objColor := color.RGBA{R: 255, A: 255}
		switch obj.Type {
		case ObjectGreen:
			objColor.G = 255
		case ObjectBlue:
			objColor.B = 255
		}
		gridX := obj.GridPosition % gridCols
		gridY := obj.GridPosition / gridCols
		if gridY < 1 || gridY > displayRows || gridX < 1 || gridX > displayCols {
			continue
		}
		x := g.gridStartX + (gridX-1)*(g.cellSize+g.gridMargin) + g.cellSize/2
		y := g.gridStartY + (gridY-1)*(g.cellSize+g.gridMargin) + g.cellSize/2
		vector.DrawFilledCircle(screen, float32(x), float32(y), 10, objColor, false)
	}

	// Draw animations
	for _, anim := range g.state.animations {
		progress := anim.Elapsed / anim.Duration
		if progress > 1 {
			progress = 1
		}
		x := anim.StartX + (anim.EndX-anim.StartX)*progress
		y := anim.StartY + (anim.EndY-anim.StartY)*progress
		vector.DrawFilledCircle(screen, float32(x), float32(y), 10, anim.Color, false)
	}
}

func (g *Game) drawTooltips(screen *ebiten.Image) {
	cx, cy := GetCursorPosition()

	// Check for selected machine first
	selected := g.getSelectedMachine()
	if selected != nil {
		// Find position of selected machine
		for pos, ms := range g.state.machines {
			if ms == selected {
				col := pos % gridCols
				row := pos / gridCols
				if row >= 1 && row <= displayRows && col >= 1 && col <= displayCols {
					x := g.gridStartX + (col-1)*(g.cellSize+g.gridMargin) + g.cellSize/2
					y := g.gridStartY + (row-1)*(g.cellSize+g.gridMargin)
					g.drawTooltip(screen, selected.Machine.GetDescription(), x, y-10)
					return
				}
			}
		}
		// Check available machines
		for i, ms := range g.state.availableMachines {
			if ms == selected {
				x := g.gridStartX + i*(g.cellSize+g.gridMargin) + g.cellSize/2
				y := g.availableY + g.cellSize
				g.drawTooltip(screen, selected.Machine.GetDescription(), x, y+10)
				return
			}
		}
	}

	// Check for hover on grid machines
	if ms := g.getMachineAt(cx, cy); ms != nil {
		col := (cx - g.gridStartX) / (g.cellSize + g.gridMargin)
		row := (cy - g.gridStartY) / (g.cellSize + g.gridMargin)
		x := g.gridStartX + col*(g.cellSize+g.gridMargin) + g.cellSize/2
		y := g.gridStartY + row*(g.cellSize+g.gridMargin)
		g.drawTooltip(screen, ms.Machine.GetDescription(), x, y-10)
		return
	}

	// Check for hover on available machines
	for i, ms := range g.state.availableMachines {
		x := g.gridStartX + i*(g.cellSize+g.gridMargin)
		y := g.availableY
		if cx >= x && cx <= x+g.cellSize && cy >= y && cy <= y+g.cellSize {
			g.drawTooltip(screen, ms.Machine.GetDescription(), x+g.cellSize/2, y+g.cellSize+10)
			return
		}
	}
}

func (g *Game) drawTooltip(screen *ebiten.Image, text string, x, y int) {
	// Measure text width (approximate)
	textWidth := len(text) * 6 // rough estimate
	boxWidth := textWidth + 20
	boxHeight := 30

	// Position box above the point
	boxX := x - boxWidth/2
	boxY := y - boxHeight - 5

	// Ensure box stays on screen
	if boxX < 10 {
		boxX = 10
	}
	if boxX+boxWidth > g.screenWidth-10 {
		boxX = g.screenWidth - boxWidth - 10
	}
	if boxY < 10 {
		boxY = y + 15 // show below if above goes off screen
	}

	// Draw background
	vector.DrawFilledRect(screen, float32(boxX), float32(boxY), float32(boxWidth), float32(boxHeight), color.RGBA{R: 0, G: 0, B: 0, A: 200}, false)
	// Draw border
	vector.StrokeRect(screen, float32(boxX), float32(boxY), float32(boxWidth), float32(boxHeight), 2, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)

	// Draw text
	ebitenutil.DebugPrintAt(screen, text, boxX+10, boxY+8)
}
