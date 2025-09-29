package game

import (
	"bytes"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

func (g *Game) drawInfoBar(screen *ebiten.Image, y int) {
	barHeight := g.infoBarHeight
	// Draw background
	vector.DrawFilledRect(screen, 0, float32(y), float32(g.screenWidth), float32(barHeight), color.RGBA{R: 80, G: 80, B: 80, A: 255}, false)

	// Define layout: leave space on left for buttons
	buttonSpace := 100
	contentWidth := g.screenWidth - buttonSpace
	boxWidth := contentWidth / 4
	rowHeight := barHeight / 2
	topRowHeight := rowHeight + 5
	bottomRowHeight := rowHeight - 5

	// Top row: Round | Runs Left | Money | Restocks Left
	colors := []color.RGBA{
		{R: 135, G: 206, B: 235, A: 255}, // Sky blue for Round
		{R: 144, G: 238, B: 144, A: 255}, // Light green for Runs Left
		{R: 255, G: 215, B: 0, A: 255},   // Gold for Money
		{R: 255, G: 99, B: 71, A: 255},   // Tomato red for Restocks
	}
	labels := []string{"Round", "Runs Left", "Money", "Restocks"}
	values := []string{
		fmt.Sprintf("%d", g.state.round),
		fmt.Sprintf("%d", g.state.runsLeft),
		fmt.Sprintf("$%d", g.state.money),
		fmt.Sprintf("%d", g.state.restocksLeft),
	}

	source, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}
	face := &text.GoTextFace{Source: source, Size: 20}

	for i := 0; i < 4; i++ {
		x := buttonSpace + i*boxWidth
		// Outer colored rectangle
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(boxWidth), float32(topRowHeight), colors[i], false)

		// Label: small text at top
		labelY := y + 2
		ebitenutil.DebugPrintAt(screen, labels[i], x+10, labelY)

		// Inner black rectangle for number
		innerWidth := 60
		innerHeight := 25
		innerX := x + (boxWidth-innerWidth)/2
		innerY := y + 15
		vector.DrawFilledRect(screen, float32(innerX), float32(innerY), float32(innerWidth), float32(innerHeight), color.RGBA{R: 0, G: 0, B: 0, A: 255}, false)

		// Number inside, centered, using larger font
		width, _ := text.Measure(values[i], face, 0)
		pixelWidth := int(width)
		numX := innerX + (innerWidth-pixelWidth)/2
		numY := innerY + innerHeight - 2
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(numX), float64(numY))
		op.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, values[i], face, op)
	}

	// Bottom row: Full width progress bar
	bottomY := y + topRowHeight
	progressText := fmt.Sprintf("Score %d / Target %d", g.state.totalScore, g.state.targetScore)
	textWidth := len(progressText) * 6
	textX := buttonSpace + (contentWidth-textWidth)/2
	ebitenutil.DebugPrintAt(screen, progressText, textX, bottomY+5)

	// Progress bar
	barMargin := 10
	barY := bottomY + 20
	barHeight2 := bottomRowHeight - 30
	barLeft := buttonSpace + barMargin
	barRight := g.screenWidth - barMargin
	barWidth := barRight - barLeft
	progress := float64(g.state.totalScore) / float64(g.state.targetScore)
	if progress > 1.0 {
		progress = 1.0
	}
	fillWidth := int(float64(barWidth) * progress)
	// Background
	vector.DrawFilledRect(screen, float32(barLeft), float32(barY), float32(barWidth), float32(barHeight2), color.RGBA{R: 50, G: 50, B: 50, A: 255}, false)
	// Fill
	vector.DrawFilledRect(screen, float32(barLeft), float32(barY), float32(fillWidth), float32(barHeight2), color.RGBA{R: 0, G: 200, B: 0, A: 255}, false)

	// Render info bar buttons
	for _, button := range g.state.buttons {
		if state := button.States[g.state.phase]; state != nil && state.Visible {
			// Check if button is in info bar
			if button.Y >= y && button.Y < y+barHeight {
				button.Render(screen, g.state)
			}
		}
	}
}
