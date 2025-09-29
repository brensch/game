package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) drawRoundEndLayout(screen *ebiten.Image) {
	// Clear screen or draw background
	vector.DrawFilledRect(screen, 0, 0, float32(g.screenWidth), float32(g.height), color.RGBA{R: 50, G: 50, B: 50, A: 255}, false)

	// Draw info bar at bottom
	g.drawInfoBar(screen, g.height-g.topPanelHeight)

}
