package main

import (
	"log"

	"github/brensch/game/pkg/game"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("Bouncing Ball Physics Game")

	// Target the second monitor
	var monitors []*ebiten.MonitorType
	monitors = ebiten.AppendMonitors(monitors)
	if len(monitors) > 1 {
		ebiten.SetMonitor(monitors[1])
	}

	g := &game.Game{
		BallX: game.ScreenWidth / 2,
		BallY: game.ScreenHeight / 2,
	}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
