package game

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

const (
	ScreenWidth     = 500
	ScreenHeight    = 500
	Gravity         = 0.3
	Damping         = 0.8
	BallRadius      = 20
	ExplosionRadius = 100
	ExplosionSpeed  = 15
)

type Ball struct {
	X, Y   float64
	VX, VY float64
	Color  color.Color
}

type Game struct {
	Balls        []Ball
	PrevTouchIDs []ebiten.TouchID
	MouseHeld    bool
}

func (g *Game) Update() error {
	var currentTouches []ebiten.TouchID
	currentTouches = ebiten.AppendTouchIDs(currentTouches)

	// Collect all input positions
	type Input struct {
		x, y  float64
		isNew bool
	}
	var inputs []Input

	// Handle touch input
	for _, id := range currentTouches {
		x, y := ebiten.TouchPosition(id)
		isNew := true
		for _, prevID := range g.PrevTouchIDs {
			if id == prevID {
				isNew = false
				break
			}
		}
		inputs = append(inputs, Input{x: float64(x), y: float64(y), isNew: isNew})
	}
	g.PrevTouchIDs = currentTouches

	// Handle mouse input
	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mouseX, mouseY := ebiten.CursorPosition()
		inputs = append(inputs, Input{x: float64(mouseX), y: float64(mouseY), isNew: true})
		g.MouseHeld = true
	} else if mousePressed && g.MouseHeld {
		mouseX, mouseY := ebiten.CursorPosition()
		inputs = append(inputs, Input{x: float64(mouseX), y: float64(mouseY), isNew: false})
	} else if !mousePressed {
		g.MouseHeld = false
	}

	// Process inputs
	for _, input := range inputs {
		if input.isNew {
			g.explodeAt(input.x, input.y)
		}
		g.addBallAt(input.x, input.y)
	}

	// Apply physics to all balls
	for i := range g.Balls {
		ball := &g.Balls[i]
		ball.VY += Gravity
		ball.X += ball.VX
		ball.Y += ball.VY

		// Bounce off walls
		if ball.X-BallRadius < 0 {
			ball.X = BallRadius
			ball.VX = -ball.VX * Damping
		} else if ball.X+BallRadius > ScreenWidth {
			ball.X = ScreenWidth - BallRadius
			ball.VX = -ball.VX * Damping
		}

		if ball.Y-BallRadius < 0 {
			ball.Y = BallRadius
			ball.VY = -ball.VY * Damping
		} else if ball.Y+BallRadius > ScreenHeight {
			ball.Y = ScreenHeight - BallRadius
			ball.VY = -ball.VY * Damping
		}
	}

	return nil
}

func (g *Game) explodeAt(x, y float64) {
	// Explode balls near the point
	for i := range g.Balls {
		ball := &g.Balls[i]
		dx := ball.X - x
		dy := ball.Y - y
		dist := dx*dx + dy*dy
		if dist < ExplosionRadius*ExplosionRadius && dist > 0 {
			// Normalize direction
			len := math.Sqrt(dist)
			dirX := dx / len
			dirY := dy / len
			// Set velocity away from explosion point
			ball.VX = dirX * ExplosionSpeed
			ball.VY = dirY * ExplosionSpeed
		}
	}

	// Add a new ball
	g.addBallAt(x, y)
}

func (g *Game) addBallAt(x, y float64) {
	newBall := Ball{
		X:     x,
		Y:     y,
		VX:    r.Float64()*4 - 2,
		VY:    r.Float64()*4 - 2,
		Color: color.RGBA{uint8(r.Intn(256)), uint8(r.Intn(256)), uint8(r.Intn(256)), 255},
	}
	g.Balls = append(g.Balls, newBall)
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, ball := range g.Balls {
		vector.DrawFilledCircle(screen, float32(ball.X), float32(ball.Y), BallRadius, ball.Color, false)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
