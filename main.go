package main

import (
	"log"

	"github/brensch/game/pkg/game"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(480, 800)
	ebiten.SetWindowTitle("Factory game")

	// Target the second monitor
	var monitors []*ebiten.MonitorType
	monitors = ebiten.AppendMonitors(monitors)
	if len(monitors) > 1 {
		ebiten.SetMonitor(monitors[1])
	}

	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
