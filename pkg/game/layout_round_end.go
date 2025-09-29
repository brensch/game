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

func (g *Game) drawRoundEndLayout(screen *ebiten.Image) {
	// Clear screen or draw background
	vector.DrawFilledRect(screen, 0, 0, float32(g.screenWidth), float32(g.height), color.RGBA{R: 50, G: 50, B: 50, A: 255}, false)

	// Points earned this run
	source, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}
	faceLarge := &text.GoTextFace{Source: source, Size: 24}
	y := g.height/2 - 20
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

	// Draw info bar at bottom
	g.drawInfoBar(screen, g.height-g.topPanelHeight)

}
