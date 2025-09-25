package mobile

import (
	"github/brensch/game/pkg/game"

	"github.com/hajimehoshi/ebiten/v2/mobile"
)

func InitGame(w, h int) {
	mobile.SetGame(game.NewGame(w, h))
}
