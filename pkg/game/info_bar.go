package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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

	// Top row: Round | Runs Left | Money | Restocks Left
	colors := []color.RGBA{
		{R: 0, G: 100, B: 200, A: 255}, // Blue for Round
		{R: 0, G: 150, B: 0, A: 255},   // Green for Runs Left
		{R: 200, G: 150, B: 0, A: 255}, // Gold for Money
		{R: 200, G: 0, B: 0, A: 255},   // Red for Restocks
	}
	labels := []string{"Round", "Runs Left", "Money", "Restocks"}
	values := []string{
		fmt.Sprintf("%d", g.state.round),
		fmt.Sprintf("%d", g.state.runsLeft),
		fmt.Sprintf("$%d", g.state.money),
		fmt.Sprintf("%d", g.state.restocksLeft),
	}

	for i := 0; i < 4; i++ {
		x := buttonSpace + i*boxWidth
		// Outer colored rectangle
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(boxWidth), float32(rowHeight), colors[i], false)

		// Label: small text at top
		labelY := y + 5
		ebitenutil.DebugPrintAt(screen, labels[i], x+10, labelY)

		// Inner black rectangle for number
		innerWidth := 50
		innerHeight := 20
		innerX := x + (boxWidth-innerWidth)/2
		innerY := y + 25
		vector.DrawFilledRect(screen, float32(innerX), float32(innerY), float32(innerWidth), float32(innerHeight), color.RGBA{R: 0, G: 0, B: 0, A: 255}, false)

		// Number inside, centered
		numX := innerX + (innerWidth-len(values[i])*6)/2
		numY := innerY + 5
		ebitenutil.DebugPrintAt(screen, values[i], numX, numY)
	}

	// Bottom row: Full width progress bar
	bottomY := y + rowHeight
	progressText := fmt.Sprintf("Score %d / Target %d", g.state.totalScore, g.state.targetScore)
	textWidth := len(progressText) * 6
	textX := buttonSpace + (contentWidth-textWidth)/2
	ebitenutil.DebugPrintAt(screen, progressText, textX, bottomY+5)

	// Progress bar
	barMargin := 10
	barY := bottomY + 20
	barHeight2 := rowHeight - 30
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
