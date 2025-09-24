package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 500
	screenHeight = 500
	gridSize     = 5
	cellSize     = screenWidth / gridSize
)

type Game struct {
	grid [gridSize][gridSize]bool
}

func (g *Game) Update() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		col := x / cellSize
		row := y / cellSize
		if col >= 0 && col < gridSize && row >= 0 && row < gridSize {
			g.grid[row][col] = !g.grid[row][col]
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			var clr color.Color
			if g.grid[i][j] {
				clr = color.RGBA{255, 0, 0, 255} // Red
			} else {
				clr = color.RGBA{255, 255, 255, 255} // White
			}
			ebitenutil.DrawRect(screen, float64(j*cellSize), float64(i*cellSize), cellSize, cellSize, clr)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("5x5 Grid Game which is cool")

	// Target the second monitor
	var monitors []*ebiten.MonitorType
	monitors = ebiten.AppendMonitors(monitors)
	if len(monitors) > 1 {
		ebiten.SetMonitor(monitors[1])
	}

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
