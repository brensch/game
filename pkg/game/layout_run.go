package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) drawRunLayout(screen *ebiten.Image) {
	// Draw factory floor
	g.drawFactoryFloor(screen)

	// Draw placed machines
	for pos, ms := range g.state.machines {
		if ms == nil || ms.Machine == nil {
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

		g.drawArrow(screen, float32(x), float32(y), ms.Orientation)
	}

	// Draw animations
	for _, anim := range g.state.animations {
		progress := anim.Elapsed / anim.Duration
		x := anim.StartX + (anim.EndX-anim.StartX)*progress
		y := anim.StartY + (anim.EndY-anim.StartY)*progress
		vector.DrawFilledRect(screen, float32(x-2), float32(y-2), 4, 4, anim.Color, false)
	}

	// Draw bottom panel
	vector.DrawFilledRect(screen, 0, float32(g.bottomY), float32(g.screenWidth), float32(g.bottomHeight), color.RGBA{R: 80, G: 80, B: 80, A: 255}, false)

	// Current Run Score (centered in the middle)
	scoreText := fmt.Sprintf("Run Score: %d x %d", g.state.roundScore, g.state.multiplier)
	scoreX := (g.screenWidth - len(scoreText)*6) / 2 // Approximate centering, assuming ~6px per char
	ebitenutil.DebugPrintAt(screen, scoreText, scoreX, g.bottomY+20)

	// Draw info bar at bottom
	g.drawInfoBar(screen, g.bottomY+g.bottomHeight)

}
