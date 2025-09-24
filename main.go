package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 500
	screenHeight = 500
	gravity      = 0.3
	damping      = 0.8
	throwFactor  = 0.1
	ballRadius   = 20
)

type Game struct {
	ballX, ballY           float64
	ballVX, ballVY         float64
	dragging               bool
	prevMouseX, prevMouseY float64
}

func (g *Game) Update() error {
	mouseX, mouseY := ebiten.CursorPosition()
	dx := mouseX - int(g.ballX)
	dy := mouseY - int(g.ballY)
	distance := dx*dx + dy*dy

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && distance < ballRadius*ballRadius && !g.dragging {
		g.dragging = true
		g.prevMouseX = float64(mouseX)
		g.prevMouseY = float64(mouseY)
	}

	if g.dragging {
		g.ballX = float64(mouseX)
		g.ballY = float64(mouseY)
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			g.ballVX = (float64(mouseX) - g.prevMouseX) * throwFactor
			g.ballVY = (float64(mouseY) - g.prevMouseY) * throwFactor
			g.dragging = false
		}
	} else {
		// Apply physics
		g.ballVY += gravity
		g.ballX += g.ballVX
		g.ballY += g.ballVY

		// Bounce off walls
		if g.ballX-ballRadius < 0 {
			g.ballX = ballRadius
			g.ballVX = -g.ballVX * damping
		} else if g.ballX+ballRadius > screenWidth {
			g.ballX = screenWidth - ballRadius
			g.ballVX = -g.ballVX * damping
		}

		if g.ballY-ballRadius < 0 {
			g.ballY = ballRadius
			g.ballVY = -g.ballVY * damping
		} else if g.ballY+ballRadius > screenHeight {
			g.ballY = screenHeight - ballRadius
			g.ballVY = -g.ballVY * damping
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	vector.DrawFilledCircle(screen, float32(g.ballX), float32(g.ballY), ballRadius, color.RGBA{0, 0, 255, 255}, false)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Bouncing Ball Physics Game")

	// Target the second monitor
	var monitors []*ebiten.MonitorType
	monitors = ebiten.AppendMonitors(monitors)
	if len(monitors) > 1 {
		ebiten.SetMonitor(monitors[1])
	}

	game := &Game{
		ballX: screenWidth / 2,
		ballY: screenHeight / 2,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
