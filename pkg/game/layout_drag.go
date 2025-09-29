package game

import (
	"bytes"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

func (g *Game) drawDragLayout(screen *ebiten.Image) {
	// Draw factory floor
	g.drawFactoryFloor(screen)

	// Draw available machines
	for i, ms := range g.state.inventory {
		if ms != nil && !ms.BeingDragged && ms.Machine != nil {
			row := i / 7
			col := i % 7
			x := g.gridStartX + col*(g.cellSize+g.gridMargin)
			y := g.availableY + row*(g.cellSize+g.gridMargin)
			vector.DrawFilledRect(screen, float32(x), float32(y), float32(g.cellSize), float32(g.cellSize), ms.Machine.GetColor(), false)
			if g.state.inventorySelected[i] {
				vector.StrokeRect(screen, float32(x), float32(y), float32(g.cellSize), float32(g.cellSize), 3, color.RGBA{R: 255, G: 0, B: 0, A: 255}, false)
			}
		}
	}

	// Draw placed machines
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

		g.drawArrow(screen, float32(x), float32(y), ms.Orientation)
		if ms.Selected {
			vector.StrokeRect(screen, float32(x), float32(y), float32(g.cellSize), float32(g.cellSize), 3, color.RGBA{R: 255, G: 255, B: 0, A: 255}, false)
		}
	}

	// Draw the dragging machine on top
	if dragging := g.getDraggingMachine(); dragging != nil {
		cx, cy := g.lastInput.X, g.lastInput.Y
		vector.DrawFilledRect(screen, float32(cx-g.cellSize/2), float32(cy-g.cellSize/2), float32(g.cellSize), float32(g.cellSize), dragging.Machine.GetColor(), false)
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

	// Current Run Score (centered in the middle) with boxes
	source, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}
	faceLarge := &text.GoTextFace{Source: source, Size: 24}
	smallBoxW := 60
	smallBoxH := 40
	gap := 10
	xW := 10
	totalW := smallBoxW + gap + xW + gap + smallBoxW
	startX := (g.screenWidth - totalW) / 2
	y := g.bottomY + 10
	baseBoxX := startX
	xPos := startX + smallBoxW + gap
	multBoxX := startX + smallBoxW + gap + xW + gap

	// Score label
	scoreWidth, _ := text.Measure("Run Score", faceLarge, 0)
	scoreX := startX - int(scoreWidth) - 20
	opScore := &text.DrawOptions{}
	opScore.GeoM.Translate(float64(scoreX), float64(y+8))
	opScore.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, "Run Score", faceLarge, opScore)

	// Base box
	vector.DrawFilledRect(screen, float32(baseBoxX), float32(y), float32(smallBoxW), float32(smallBoxH), color.RGBA{R: 0, G: 0, B: 0, A: 255}, false)

	// x
	opX := &text.DrawOptions{}
	opX.GeoM.Translate(float64(xPos), float64(y+8))
	opX.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, "x", faceLarge, opX)

	// Mult box
	vector.DrawFilledRect(screen, float32(multBoxX), float32(y), float32(smallBoxW), float32(smallBoxH), color.RGBA{R: 0, G: 0, B: 0, A: 255}, false)

	// Base text
	baseStr := fmt.Sprintf("%d", g.state.roundScore)
	opBase := &text.DrawOptions{}
	opBase.GeoM.Translate(float64(baseBoxX+10), float64(y+8))
	opBase.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, baseStr, faceLarge, opBase)

	// Mult text
	multStr := fmt.Sprintf("%d", g.state.multiplier)
	opMult := &text.DrawOptions{}
	opMult.GeoM.Translate(float64(multBoxX+10), float64(y+8))
	opMult.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, multStr, faceLarge, opMult)

	// Draw info bar at bottom
	g.drawInfoBar(screen, g.bottomY+g.bottomHeight)

}
