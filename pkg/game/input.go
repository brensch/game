package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// InputState represents the current input state
type InputState struct {
	Pressed        bool
	JustPressed    bool
	JustReleased   bool
	X, Y           int
	IsDragging     bool
	DragStartX     int
	DragStartY     int
	LastTouchCount int
	TouchPressed   bool
	LastTouchX     int
	LastTouchY     int
}

// getUnifiedInput returns a unified input state that works for both mouse and touch
func getUnifiedInput(prev InputState) InputState {
	// Check mouse input first
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		return InputState{
			Pressed:        true,
			JustPressed:    true,
			JustReleased:   false,
			X:              x,
			Y:              y,
			IsDragging:     false,
			DragStartX:     x,
			DragStartY:     y,
			LastTouchCount: prev.LastTouchCount,
			TouchPressed:   prev.TouchPressed,
			LastTouchX:     prev.LastTouchX,
			LastTouchY:     prev.LastTouchY,
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
			Pressed:        true,
			JustPressed:    false,
			JustReleased:   false,
			X:              x,
			Y:              y,
			IsDragging:     newIsDragging,
			DragStartX:     prev.DragStartX,
			DragStartY:     prev.DragStartY,
			LastTouchCount: prev.LastTouchCount,
			TouchPressed:   prev.TouchPressed,
			LastTouchX:     prev.LastTouchX,
			LastTouchY:     prev.LastTouchY,
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		return InputState{
			Pressed:        false,
			JustPressed:    false,
			JustReleased:   true,
			X:              x,
			Y:              y,
			IsDragging:     false,
			DragStartX:     prev.DragStartX,
			DragStartY:     prev.DragStartY,
			LastTouchCount: prev.LastTouchCount,
			TouchPressed:   prev.TouchPressed,
			LastTouchX:     prev.LastTouchX,
			LastTouchY:     prev.LastTouchY,
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
		if !newIsDragging && dx*dx+dy*dy > 1000 {
			newIsDragging = true
		}
		return InputState{
			Pressed:        true,
			JustPressed:    justPressed,
			JustReleased:   false,
			X:              x,
			Y:              y,
			IsDragging:     newIsDragging,
			DragStartX:     newDragStartX,
			DragStartY:     newDragStartY,
			LastTouchCount: currentTouchCount,
			TouchPressed:   true,
			LastTouchX:     x,
			LastTouchY:     y,
		}
	}

	if justReleased {
		return InputState{
			Pressed:        false,
			JustPressed:    false,
			JustReleased:   true,
			X:              prev.LastTouchX,
			Y:              prev.LastTouchY,
			IsDragging:     false,
			DragStartX:     prev.DragStartX,
			DragStartY:     prev.DragStartY,
			LastTouchCount: currentTouchCount,
			TouchPressed:   false,
			LastTouchX:     prev.LastTouchX,
			LastTouchY:     prev.LastTouchY,
		}
	}

	// No input
	return InputState{
		Pressed:        false,
		JustPressed:    false,
		JustReleased:   false,
		X:              0,
		Y:              0,
		IsDragging:     prev.IsDragging,
		DragStartX:     prev.DragStartX,
		DragStartY:     prev.DragStartY,
		LastTouchCount: prev.LastTouchCount,
		TouchPressed:   prev.TouchPressed,
		LastTouchX:     prev.LastTouchX,
		LastTouchY:     prev.LastTouchY,
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
	g.lastInput = getUnifiedInput(g.lastInput)
}
