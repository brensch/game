package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var (
	lastTouchCount int
	touchPressed   bool
)

// InputState represents the current input state
type InputState struct {
	Pressed      bool
	JustPressed  bool
	JustReleased bool
	X, Y         int
}

// GetUnifiedInput returns a unified input state that works for both mouse and touch
func GetUnifiedInput() InputState {
	// Check mouse input first
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		return InputState{
			Pressed:      true,
			JustPressed:  true,
			JustReleased: false,
			X:            x,
			Y:            y,
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		return InputState{
			Pressed:      true,
			JustPressed:  false,
			JustReleased: false,
			X:            x,
			Y:            y,
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		return InputState{
			Pressed:      false,
			JustPressed:  false,
			JustReleased: true,
			X:            x,
			Y:            y,
		}
	}

	// Check touch input
	var touchIDs []ebiten.TouchID
	touchIDs = ebiten.AppendTouchIDs(touchIDs)
	currentTouchCount := len(touchIDs)

	justPressed := false
	justReleased := false

	if currentTouchCount > 0 && lastTouchCount == 0 {
		justPressed = true
		touchPressed = true
	} else if currentTouchCount == 0 && lastTouchCount > 0 {
		justReleased = true
		touchPressed = false
	}

	lastTouchCount = currentTouchCount

	if currentTouchCount > 0 {
		// Use the first touch (earliest touch)
		id := touchIDs[0]
		x, y := ebiten.TouchPosition(id)
		return InputState{
			Pressed:      true,
			JustPressed:  justPressed,
			JustReleased: false,
			X:            x,
			Y:            y,
		}
	}

	if justReleased {
		return InputState{
			Pressed:      false,
			JustPressed:  false,
			JustReleased: true,
			X:            0, // No position on release
			Y:            0,
		}
	}

	// No input
	return InputState{
		Pressed:      false,
		JustPressed:  false,
		JustReleased: false,
		X:            0,
		Y:            0,
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
