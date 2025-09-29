package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// InputState represents the current input state
type InputState struct {
	Pressed            bool
	JustPressed        bool
	JustReleased       bool
	X, Y               int
	IsDragging         bool
	DragStartX         int
	DragStartY         int
	LastTouchCount     int
	TouchPressed       bool
	LastTouchX         int
	LastTouchY         int
	ClickStartFrame    int
	LongClickedMachine *MachineState
}

// getUnifiedInput returns a unified input state that works for both mouse and touch
func getUnifiedInput(prev InputState, currentFrame int) InputState {
	// Check mouse input first
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		return InputState{
			Pressed:            true,
			JustPressed:        true,
			JustReleased:       false,
			X:                  x,
			Y:                  y,
			IsDragging:         false,
			DragStartX:         x,
			DragStartY:         y,
			LastTouchCount:     prev.LastTouchCount,
			TouchPressed:       prev.TouchPressed,
			LastTouchX:         prev.LastTouchX,
			LastTouchY:         prev.LastTouchY,
			ClickStartFrame:    currentFrame,
			LongClickedMachine: nil,
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		dx := x - prev.DragStartX
		dy := y - prev.DragStartY
		newIsDragging := prev.IsDragging
		if !newIsDragging && dx*dx+dy*dy > 1000 {
			newIsDragging = true
		}
		return InputState{
			Pressed:            true,
			JustPressed:        false,
			JustReleased:       false,
			X:                  x,
			Y:                  y,
			IsDragging:         newIsDragging,
			DragStartX:         prev.DragStartX,
			DragStartY:         prev.DragStartY,
			LastTouchCount:     prev.LastTouchCount,
			TouchPressed:       prev.TouchPressed,
			LastTouchX:         prev.LastTouchX,
			LastTouchY:         prev.LastTouchY,
			ClickStartFrame:    prev.ClickStartFrame,
			LongClickedMachine: prev.LongClickedMachine,
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		return InputState{
			Pressed:            false,
			JustPressed:        false,
			JustReleased:       true,
			X:                  x,
			Y:                  y,
			IsDragging:         false,
			DragStartX:         prev.DragStartX,
			DragStartY:         prev.DragStartY,
			LastTouchCount:     prev.LastTouchCount,
			TouchPressed:       prev.TouchPressed,
			LastTouchX:         prev.LastTouchX,
			LastTouchY:         prev.LastTouchY,
			ClickStartFrame:    prev.ClickStartFrame,
			LongClickedMachine: prev.LongClickedMachine,
		}
	}

	// Check touch input
	var touchIDs []ebiten.TouchID
	touchIDs = ebiten.AppendTouchIDs(touchIDs)
	currentTouchCount := len(touchIDs)

	justPressed := false
	justReleased := false

	if currentTouchCount > 0 && prev.LastTouchCount == 0 {
		justPressed = true
	} else if currentTouchCount == 0 && prev.LastTouchCount > 0 {
		justReleased = true
	}

	if currentTouchCount > 0 {
		// Use the first touch (earliest touch)
		id := touchIDs[0]
		x, y := ebiten.TouchPosition(id)
		newIsDragging := prev.IsDragging
		newDragStartX := prev.DragStartX
		newDragStartY := prev.DragStartY
		if justPressed {
			newIsDragging = false
			newDragStartX = x
			newDragStartY = y
		}
		dx := x - newDragStartX
		dy := y - newDragStartY
		if !newIsDragging && dx*dx+dy*dy > 2500 {
			newIsDragging = true
		}
		newClickStartFrame := prev.ClickStartFrame
		if justPressed {
			newClickStartFrame = currentFrame
		}
		return InputState{
			Pressed:            true,
			JustPressed:        justPressed,
			JustReleased:       false,
			X:                  x,
			Y:                  y,
			IsDragging:         newIsDragging,
			DragStartX:         newDragStartX,
			DragStartY:         newDragStartY,
			LastTouchCount:     currentTouchCount,
			TouchPressed:       true,
			LastTouchX:         x,
			LastTouchY:         y,
			ClickStartFrame:    newClickStartFrame,
			LongClickedMachine: prev.LongClickedMachine,
		}
	}

	if justReleased {
		return InputState{
			Pressed:            false,
			JustPressed:        false,
			JustReleased:       true,
			X:                  prev.LastTouchX,
			Y:                  prev.LastTouchY,
			IsDragging:         false,
			DragStartX:         prev.DragStartX,
			DragStartY:         prev.DragStartY,
			LastTouchCount:     currentTouchCount,
			TouchPressed:       false,
			LastTouchX:         prev.LastTouchX,
			LastTouchY:         prev.LastTouchY,
			ClickStartFrame:    prev.ClickStartFrame,
			LongClickedMachine: prev.LongClickedMachine,
		}
	}

	// No input
	return InputState{
		Pressed:            false,
		JustPressed:        false,
		JustReleased:       false,
		X:                  0,
		Y:                  0,
		IsDragging:         prev.IsDragging,
		DragStartX:         prev.DragStartX,
		DragStartY:         prev.DragStartY,
		LastTouchCount:     prev.LastTouchCount,
		TouchPressed:       prev.TouchPressed,
		LastTouchX:         prev.LastTouchX,
		LastTouchY:         prev.LastTouchY,
		ClickStartFrame:    prev.ClickStartFrame,
		LongClickedMachine: prev.LongClickedMachine,
	}
}

// GetCursorPosition returns the current cursor/touch position
func GetCursorPosition() (int, int) {
	// Check for touch first
	var touchIDs []ebiten.TouchID
	touchIDs = ebiten.AppendTouchIDs(touchIDs)
	if len(touchIDs) > 0 {
		return ebiten.TouchPosition(touchIDs[0])
	}

	// Fall back to mouse
	return ebiten.CursorPosition()
}

func (g *Game) GetInput() {
	g.lastInput = getUnifiedInput(g.lastInput, g.frameCount)

	// Clear long clicked machine when dragging starts
	if g.lastInput.IsDragging && g.state.longClickedMachine != nil {
		g.state.longClickedMachine = nil
	}

	// Check if we need to clear the long clicked machine on new click
	if g.lastInput.JustPressed {
		// Find which machine is being clicked now
		cx, cy := g.lastInput.X, g.lastInput.Y
		var clickedMachine *MachineState

		// Check grid machines
		for pos := 0; pos < gridCols*gridRows; pos++ {
			ms := g.state.machines[pos]
			if ms != nil && !ms.BeingDragged && ms.Machine != nil {
				col := pos % gridCols
				row := pos / gridCols
				if row >= 1 && row <= displayRows && col >= 1 && col <= displayCols {
					x := g.gridStartX + (col-1)*(g.cellSize+g.gridMargin)
					y := g.gridStartY + (row-1)*(g.cellSize+g.gridMargin)
					if cx >= x-15 && cx <= x+g.cellSize+15 && cy >= y-15 && cy <= y+g.cellSize+15 {
						clickedMachine = ms
						break
					}
				}
			}
		}

		// If not on grid, check inventory
		if clickedMachine == nil {
			for i, ms := range g.state.inventory {
				if ms != nil && !ms.BeingDragged && ms.Machine != nil {
					row := i / 7
					col := i % 7
					x := g.gridStartX + col*(g.cellSize+g.gridMargin)
					y := g.availableY + row*(g.cellSize+g.gridMargin)
					if cx >= x-15 && cx <= x+g.cellSize+15 && cy >= y-15 && cy <= y+g.cellSize+15 {
						clickedMachine = ms
						break
					}
				}
			}
		}

		// If clicking on a different machine or empty space, clear the long clicked machine
		if clickedMachine != g.state.longClickedMachine {
			g.state.longClickedMachine = nil
		}
	}

	// Check for long click on machines
	if g.lastInput.Pressed && !g.lastInput.IsDragging {
		framesHeld := g.frameCount - g.lastInput.ClickStartFrame
		if framesHeld >= longClickThreshold {
			// Find which machine is being clicked
			cx, cy := g.lastInput.X, g.lastInput.Y

			// Check grid machines
			for pos := 0; pos < gridCols*gridRows; pos++ {
				ms := g.state.machines[pos]
				if ms != nil && ms.Machine != nil {
					col := pos % gridCols
					row := pos / gridCols
					if row >= 1 && row <= displayRows && col >= 1 && col <= displayCols {
						x := g.gridStartX + (col-1)*(g.cellSize+g.gridMargin)
						y := g.gridStartY + (row-1)*(g.cellSize+g.gridMargin)
						if cx >= x-15 && cx <= x+g.cellSize+15 && cy >= y-15 && cy <= y+g.cellSize+15 {
							g.lastInput.LongClickedMachine = ms
							g.state.longClickedMachine = ms
							break
						}
					}
				}
			}

			// If not on grid, check inventory
			if g.lastInput.LongClickedMachine == nil {
				for i, ms := range g.state.inventory {
					if ms != nil && ms.Machine != nil {
						row := i / 7
						col := i % 7
						x := g.gridStartX + col*(g.cellSize+g.gridMargin)
						y := g.availableY + row*(g.cellSize+g.gridMargin)
						if cx >= x-15 && cx <= x+g.cellSize+15 && cy >= y-15 && cy <= y+g.cellSize+15 {
							g.lastInput.LongClickedMachine = ms
							g.state.longClickedMachine = ms
							break
						}
					}
				}
			}
		}
	}
}
