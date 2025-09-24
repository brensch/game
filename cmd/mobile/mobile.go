package mobile

import (
	"image/color"

	"github/brensch/game/pkg/game"

	"github.com/hajimehoshi/ebiten/v2/mobile"
)

func init() {
	mobile.SetGame(&game.Game{
		Balls: []game.Ball{{
			X:     game.ScreenWidth / 2,
			Y:     game.ScreenHeight / 2,
			Color: color.RGBA{0, 0, 255, 255},
		}},
	})
}

// Dummy is a dummy exported function.
//
// gomobile doesn't compile a package that doesn't include any exported function.
// Dummy forces gomobile to compile this package.
func Dummy() {}
