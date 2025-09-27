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
	boxWidth := contentWidth / 3
	rowHeight := barHeight / 2

	// Top row: Round | Runs Left | Money
	// Round box (left)
	roundX := buttonSpace
	vector.DrawFilledRect(screen, float32(roundX), float32(y), float32(boxWidth), float32(rowHeight), color.RGBA{R: 100, G: 100, B: 100, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Round: %d", g.state.round), roundX+10, y+15)

	// Runs Left box (middle)
	runsX := roundX + boxWidth
	vector.DrawFilledRect(screen, float32(runsX), float32(y), float32(boxWidth), float32(rowHeight), color.RGBA{R: 100, G: 100, B: 100, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Runs Left: %d", g.state.runsLeft), runsX+10, y+15)

	// Money box (right)
	moneyX := runsX + boxWidth
	vector.DrawFilledRect(screen, float32(moneyX), float32(y), float32(boxWidth), float32(rowHeight), color.RGBA{R: 100, G: 100, B: 100, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Money: $%d", g.state.money), moneyX+10, y+15)

	// Bottom row: Score | Progress Bar | Target
	bottomY := y + rowHeight

	// Score box (bottom left)
	vector.DrawFilledRect(screen, float32(roundX), float32(bottomY), float32(boxWidth), float32(rowHeight), color.RGBA{R: 100, G: 100, B: 100, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", g.state.totalScore), roundX+10, bottomY+15)

	// Progress bar box (bottom middle)
	vector.DrawFilledRect(screen, float32(runsX), float32(bottomY), float32(boxWidth), float32(rowHeight), color.RGBA{R: 100, G: 100, B: 100, A: 255}, false)
	// Progress bar
	barMargin := 10
	barY := bottomY + 10
	barHeight2 := rowHeight - 20
	barLeft := runsX + barMargin
	barRight := runsX + boxWidth - barMargin
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

	// Target box (bottom right)
	vector.DrawFilledRect(screen, float32(moneyX), float32(bottomY), float32(boxWidth), float32(rowHeight), color.RGBA{R: 100, G: 100, B: 100, A: 255}, false)
	targetStr := fmt.Sprintf("Target: %d", g.state.targetScore)
	targetX := moneyX + (boxWidth-len(targetStr)*6)/2 // Center the text
	ebitenutil.DebugPrintAt(screen, targetStr, targetX, bottomY+15)

	// Render info bar buttons
	for _, button := range g.state.buttons {
		if button.States[g.state.phase].Visible {
			// Check if button is in info bar
			if button.Y >= y && button.Y < y+barHeight {
				button.Render(screen, g.state.phase)
			}
		}
	}
}
