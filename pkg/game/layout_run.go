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
		size := float64(g.cellSize) / 4
		vector.DrawFilledRect(screen, float32(x-size/2), float32(y-size/2), float32(size), float32(size), anim.Color, false)
	}

	// Draw bottom panel
	vector.DrawFilledRect(screen, 0, float32(g.bottomY), float32(g.screenWidth), float32(g.bottomHeight), color.RGBA{R: 80, G: 80, B: 80, A: 255}, false)

	// Current Run Score (centered in the middle) with boxes
	source, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}
	faceLarge := &text.GoTextFace{Source: source, Size: 24}
	y := g.bottomY + 10

	if len(g.state.animations) == 0 && g.state.roundScore > 0 {
		// Run ended, show the total points earned
		total := g.state.roundScore * g.state.multiplier
		totalStr := fmt.Sprintf("%d", total)
		totalWidth, _ := text.Measure(totalStr, faceLarge, 0)
		totalBoxW := int(totalWidth) + 20
		totalBoxH := 40
		totalBoxX := (g.screenWidth - totalBoxW) / 2
		vector.DrawFilledRect(screen, float32(totalBoxX), float32(y), float32(totalBoxW), float32(totalBoxH), color.RGBA{R: 0, G: 0, B: 0, A: 255}, false)
		opTotal := &text.DrawOptions{}
		opTotal.GeoM.Translate(float64(totalBoxX+10), float64(y+8))
		opTotal.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, totalStr, faceLarge, opTotal)
	} else {
		// Running, show Run Score with boxes
		smallBoxW := 60
		smallBoxH := 40
		gap := 10
		xW := 10
		totalW := smallBoxW + gap + xW + gap + smallBoxW
		startX := (g.screenWidth - totalW) / 2
		baseBoxX := startX
		xPos := startX + smallBoxW + gap
		multBoxX := startX + smallBoxW + gap + xW + gap

		// Run Score label
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
	}

	// Draw info bar at bottom
	g.drawInfoBar(screen, g.bottomY+g.bottomHeight)

}
