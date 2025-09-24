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
	Balls                  []Ball
	PrevTouchIDs           []ebiten.TouchID
	MouseHeld              bool
	MouseHoldFrames        int
	MouseFrameCounter      int
	MouseCurrentThreshold  int
	TouchHoldFrames        map[ebiten.TouchID]int
	TouchFrameCounters     map[ebiten.TouchID]int
	TouchCurrentThresholds map[ebiten.TouchID]int
}

func (g *Game) Update() error {
	var currentTouches []ebiten.TouchID
	currentTouches = ebiten.AppendTouchIDs(currentTouches)

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
		if isNew {
			// New touch: explode
			g.explodeAt(float64(x), float64(y))
			g.TouchFrameCounters[id] = 0
			g.TouchCurrentThresholds[id] = 30
		} else {
			// Continuing touch: add ball
			g.addBallAt(float64(x), float64(y))
		}
	}
	// Remove old touch frames
	for id := range g.TouchHoldFrames {
		found := false
		for _, currID := range currentTouches {
			if id == currID {
				found = true
				break
			}
		}
		if !found {
			delete(g.TouchHoldFrames, id)
			delete(g.TouchFrameCounters, id)
			delete(g.TouchCurrentThresholds, id)
		}
	}
	g.PrevTouchIDs = currentTouches

	// Handle mouse input
	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Just pressed: explode
		mouseX, mouseY := ebiten.CursorPosition()
		g.explodeAt(float64(mouseX), float64(mouseY))
		g.MouseHeld = true
		g.MouseFrameCounter = 0
		g.MouseCurrentThreshold = 30
	} else if mousePressed && g.MouseHeld {
		// Held: add ball
		mouseX, mouseY := ebiten.CursorPosition()
		g.addBallAt(float64(mouseX), float64(mouseY))
	} else if !mousePressed {
		g.MouseHeld = false
		g.MouseFrameCounter = 0
		g.MouseCurrentThreshold = 30
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
