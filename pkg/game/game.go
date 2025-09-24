package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	ScreenWidth  = 500
	ScreenHeight = 500
	Gravity      = 0.3
	Damping      = 0.8
	ThrowFactor  = 0.1
	BallRadius   = 20
)

type Game struct {
	BallX, BallY   float64
	BallVX, BallVY float64
	Dragging       bool
	PrevX, PrevY   float64
}

func (g *Game) Update() error {
	mouseX, mouseY := ebiten.CursorPosition()
	dx := mouseX - int(g.BallX)
	dy := mouseY - int(g.BallY)
	distance := dx*dx + dy*dy

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && distance < BallRadius*BallRadius && !g.Dragging {
		g.Dragging = true
		g.PrevX = float64(mouseX)
		g.PrevY = float64(mouseY)
	}

	if g.Dragging {
		g.BallX = float64(mouseX)
		g.BallY = float64(mouseY)
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			g.BallVX = (float64(mouseX) - g.PrevX) * ThrowFactor
			g.BallVY = (float64(mouseY) - g.PrevY) * ThrowFactor
			g.Dragging = false
		}
	} else {
		// Apply physics
		g.BallVY += Gravity
		g.BallX += g.BallVX
		g.BallY += g.BallVY

		// Bounce off walls
		if g.BallX-BallRadius < 0 {
			g.BallX = BallRadius
			g.BallVX = -g.BallVX * Damping
		} else if g.BallX+BallRadius > ScreenWidth {
			g.BallX = ScreenWidth - BallRadius
			g.BallVX = -g.BallVX * Damping
		}

		if g.BallY-BallRadius < 0 {
			g.BallY = BallRadius
			g.BallVY = -g.BallVY * Damping
		} else if g.BallY+BallRadius > ScreenHeight {
			g.BallY = ScreenHeight - BallRadius
			g.BallVY = -g.BallVY * Damping
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	vector.DrawFilledCircle(screen, float32(g.BallX), float32(g.BallY), BallRadius, color.RGBA{0, 0, 255, 255}, false)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
