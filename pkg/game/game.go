package game

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	topPanelHeight  = 80
	foremanHeight   = 80
	availableHeight = 60
	bottomHeight    = 60
	minGap          = 10

	gridCols   = 7
	gridRows   = 7
	cellSize   = 60
	gridMargin = 10
)

// ObjectType represents the different kinds of items that can move through the factory.
type ObjectType int

const (
	ObjectRed ObjectType = iota
	ObjectGreen
	ObjectBlue
)

// MachineType represents the different kinds of machines.
type MachineType int

const (
	MachineConveyor MachineType = iota
	MachineProcessor
	MachineStart
	MachineEnd
)

// GamePhase represents the current state of the game (building or running).
type GamePhase int

const (
	PhaseBuild GamePhase = iota
	PhaseRun
)

// EffectType represents different effects machines can have.
type EffectType int

const (
	EffectHolographic EffectType = iota
	EffectShiny
	EffectBuffSpeed
)

// DurationType represents how effect duration is measured.
type DurationType int

const (
	DurationTick DurationType = iota
	DurationRun
	DurationRound
)

// Object represents an item moving through the factory.
type Object struct {
	X, Y      float64
	Type      ObjectType
	pathIndex int
}

// MachineInterface defines the behavior for different machine types.
type MachineInterface interface {
	GetType() MachineType
	GetColor() color.RGBA
	CanMove(round int) bool
	Process(obj *Object, game *Game) bool
	EmitEffects(game *Game) []EffectEmission
	ReceiveEffect(effect Effect)
}

// Machine represents a machine on the factory floor.
type Machine struct {
	X, Y         int
	GridX, GridY int
	Type         MachineType
	Color        color.Color
	IsDraggable  bool
	IsPlaced     bool
	RoundAdded   int
	Effects      []Effect
}

// Effect represents an effect applied to a machine.
type Effect struct {
	Type         EffectType
	Duration     int
	DurationType DurationType
}

// EffectEmission represents an effect emitted by a machine to other machines.
type EffectEmission struct {
	TargetGridX int
	TargetGridY int
	Effect      Effect
}

// GetType returns the machine type.
func (m *Machine) GetType() MachineType {
	return m.Type
}

// GetColor returns the machine color.
func (m *Machine) GetColor() color.RGBA {
	if c, ok := m.Color.(color.RGBA); ok {
		return c
	}
	return color.RGBA{R: 128, G: 128, B: 128, A: 255} // default
}

// CanMove checks if the machine can be moved based on the current round.
func (m *Machine) CanMove(round int) bool {
	return round > m.RoundAdded
}

// Process handles object interaction based on machine type.
func (m *Machine) Process(obj *Object, game *Game) bool {
	switch m.Type {
	case MachineStart:
		// Start machine emits objects, doesn't process them
		return false
	case MachineEnd:
		// End machine consumes objects, adds score
		game.state.baseScore++
		return true // consumed
	case MachineConveyor:
		// Move object to next position
		// For now, simple move down
		obj.Y += 5
		return false
	case MachineProcessor:
		// Change object type or something
		obj.Type = (obj.Type + 1) % 3
		return false
	default:
		return false
	}
}

// EmitEffects emits effects to other machines.
func (m *Machine) EmitEffects(game *Game) []EffectEmission {
	// For now, no effects
	return nil
}

// ReceiveEffect applies an effect to this machine.
func (m *Machine) ReceiveEffect(effect Effect) {
	m.Effects = append(m.Effects, effect)
}

// GameState holds all information about the current state of the game.
type GameState struct {
	phase                    GamePhase
	money                    int
	run                      int
	maxRuns                  int
	baseScore                int
	multiplier               int
	machinesOnGrid           []MachineInterface
	availableMachines        []MachineInterface
	objects                  []*Object
	draggingMachine          MachineInterface
	dragOffsetX, dragOffsetY int
	startMachine             MachineInterface
	endMachine               MachineInterface
	currentRound             int
}

// Game implements ebiten.Game.
type Game struct {
	state                                                                    *GameState
	width, height                                                            int
	topPanelHeight, foremanHeight, gridHeight, availableHeight, bottomHeight int
	topPanelY, foremanY, gridStartY, availableY, bottomY                     int
	screenWidth, gridStartX                                                  int
}

// NewGame creates a new Game instance.
func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	state := &GameState{
		phase:          PhaseBuild,
		money:          7,
		run:            3,
		maxRuns:        6,
		baseScore:      0,
		multiplier:     1,
		machinesOnGrid: make([]MachineInterface, gridCols*gridRows),
		currentRound:   1,
	}

	g := &Game{state: state}
	g.width = 480
	g.height = 800
	g.calculateLayout()

	state.availableMachines = []MachineInterface{
		&Machine{X: g.gridStartX, Y: g.availableY, Type: MachineConveyor, Color: color.RGBA{R: 200, G: 200, B: 200, A: 255}, IsDraggable: true},
		&Machine{X: g.gridStartX + cellSize + gridMargin, Y: g.availableY, Type: MachineProcessor, Color: color.RGBA{R: 100, G: 200, B: 100, A: 255}, IsDraggable: true},
	}

	// Place fixed Start and End machines
	startX, startY := 3, 5
	endX, endY := 1, 3

	state.startMachine = &Machine{
		GridX: startX, GridY: startY,
		X: g.gridStartX + startX*(cellSize+gridMargin), Y: g.gridStartY + startY*(cellSize+gridMargin),
		Type: MachineStart, Color: color.RGBA{R: 150, G: 255, B: 150, A: 255}, RoundAdded: 0,
	}
	state.machinesOnGrid[startX*gridRows+startY] = state.startMachine

	state.endMachine = &Machine{
		GridX: endX, GridY: endY,
		X: g.gridStartX + endX*(cellSize+gridMargin), Y: g.gridStartY + endY*(cellSize+gridMargin),
		Type: MachineEnd, Color: color.RGBA{R: 255, G: 150, B: 150, A: 255}, RoundAdded: 0,
	}
	state.machinesOnGrid[endX*gridRows+endY] = state.endMachine

	return g
}

func (g *Game) calculateLayout() {
	g.topPanelHeight = topPanelHeight
	g.foremanHeight = foremanHeight
	g.availableHeight = availableHeight
	g.bottomHeight = bottomHeight

	gridHeight := gridRows*cellSize + (gridRows-1)*gridMargin
	gap := (g.height - 760) / 5
	if gap < 0 {
		gap = 0
	}
	g.topPanelY = gap
	g.foremanY = g.topPanelY + g.topPanelHeight + gap
	g.gridStartY = g.foremanY + g.foremanHeight + gap
	g.availableY = g.gridStartY + gridHeight + gap
	g.bottomY = g.availableY + g.availableHeight + gap
	g.screenWidth = g.width
	g.gridStartX = (g.screenWidth - (gridCols*cellSize + (gridCols-1)*gridMargin)) / 2
}

// Update proceeds the game state.
func (g *Game) Update() error {
	switch g.state.phase {
	case PhaseBuild:
		g.handleDragAndDrop()
	case PhaseRun:
		g.updateRun()
	}

	// Check for "Start Run" button click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		// Simple button detection for "Start Run"
		if cx > 250 && cx < g.screenWidth-30 && cy > g.bottomY+10 && cy < g.bottomY+10+g.bottomHeight-20 {
			if g.state.phase == PhaseBuild {
				g.state.phase = PhaseRun
				// Spawn a test object
				g.state.objects = append(g.state.objects, &Object{
					X:    float64(g.state.startMachine.(*Machine).X + cellSize/2),
					Y:    float64(g.state.startMachine.(*Machine).Y + cellSize/2),
					Type: ObjectType(rand.Intn(3)),
				})
			} else {
				g.state.phase = PhaseBuild
				g.state.objects = nil // Clear objects
				g.state.baseScore = 0 // Reset score
				g.state.currentRound++
			}
		}
	}
	return nil
}

func (g *Game) handleDragAndDrop() {
	cx, cy := ebiten.CursorPosition()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Check if we are picking up a new machine
		for _, m := range g.state.availableMachines {
			mm := m.(*Machine)
			if !mm.IsPlaced && cx >= mm.X && cx <= mm.X+cellSize && cy >= mm.Y && cy <= mm.Y+cellSize {
				g.state.draggingMachine = &Machine{
					Type:        mm.Type,
					Color:       mm.Color,
					IsDraggable: true,
				}
				g.state.dragOffsetX = cx - mm.X
				g.state.dragOffsetY = cy - mm.Y
				break
			}
		}

		// Check if picking up from grid (only current round machines)
		if g.state.draggingMachine == nil {
			for col := 0; col < gridCols; col++ {
				for row := 0; row < gridRows; row++ {
					if m := g.state.machinesOnGrid[col*gridRows+row]; m != nil {
						mm := m.(*Machine)
						if mm.RoundAdded == g.state.currentRound {
							if cx >= mm.X && cx <= mm.X+cellSize && cy >= mm.Y && cy <= mm.Y+cellSize {
								g.state.draggingMachine = m
								g.state.dragOffsetX = cx - mm.X
								g.state.dragOffsetY = cy - mm.Y
								// Remove from grid temporarily
								g.state.machinesOnGrid[col*gridRows+row] = nil
								break
							}
						}
					}
				}
				if g.state.draggingMachine != nil {
					break
				}
			}
		}
	}

	if g.state.draggingMachine != nil {
		dm := g.state.draggingMachine.(*Machine)
		dm.X = cx - g.state.dragOffsetX
		dm.Y = cy - g.state.dragOffsetY

		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			gridX, gridY := -1, -1
			// Check if dropped on the grid
			if cx > g.gridStartX && cx < g.gridStartX+gridCols*(cellSize+gridMargin) &&
				cy > g.gridStartY && cy < g.gridStartY+gridRows*(cellSize+gridMargin) {

				// Snap to grid
				col := (cx - g.gridStartX) / (cellSize + gridMargin)
				row := (cy - g.gridStartY) / (cellSize + gridMargin)

				if g.state.machinesOnGrid[col*gridRows+row] == nil {
					gridX, gridY = col, row
				}
			}

			// Check if dropped on sell area
			sellX, sellY, sellW, sellH := 10, g.bottomY+10, 120, g.bottomHeight-20
			if cx >= sellX && cx <= sellX+sellW && cy >= sellY && cy <= sellY+sellH {
				// Sell the machine, refund money
				g.state.money += 1
				g.state.draggingMachine = nil
			} else if gridX != -1 {
				// Place on grid
				if !dm.IsPlaced {
					g.state.money -= 1 // Deduct for buying
				}
				dm.GridX = gridX
				dm.GridY = gridY
				dm.X = g.gridStartX + gridX*(cellSize+gridMargin)
				dm.Y = g.gridStartY + gridY*(cellSize+gridMargin)
				dm.IsPlaced = true
				dm.RoundAdded = g.state.currentRound
				g.state.machinesOnGrid[gridX*gridRows+gridY] = g.state.draggingMachine
				g.state.draggingMachine = nil
			} else {
				// Dropped elsewhere, discard
				g.state.draggingMachine = nil
			}
		}
	}
}

func (g *Game) updateRun() {
	// Basic logic: move objects towards the end machine
	for _, obj := range g.state.objects {
		if obj.X < float64(g.state.endMachine.(*Machine).X+cellSize/2) {
			obj.X += 1
		} else if obj.X > float64(g.state.endMachine.(*Machine).X+cellSize/2) {
			obj.X -= 1
		}
		if obj.Y < float64(g.state.endMachine.(*Machine).Y+cellSize/2) {
			obj.Y += 1
		} else if obj.Y > float64(g.state.endMachine.(*Machine).Y+cellSize/2) {
			obj.Y -= 1
		}

		// Check if object reached the end
		distX := obj.X - float64(g.state.endMachine.(*Machine).X+cellSize/2)
		distY := obj.Y - float64(g.state.endMachine.(*Machine).Y+cellSize/2)
		if distX*distX+distY*distY < 10*10 { // Within 10 pixels
			g.state.baseScore += 1 // Score!
			// For now, just remove the object. In a real game, you'd handle it differently.
			g.state.objects = g.state.objects[1:]
		}
	}
}

// Draw draws the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 40, G: 40, B: 40, A: 255})
	g.drawUI(screen)
	g.drawFactoryFloor(screen)
	g.drawMachines(screen)
	g.drawObjects(screen)

	// Draw the dragging machine on top
	if g.state.draggingMachine != nil {
		dm := g.state.draggingMachine.(*Machine)
		vector.DrawFilledRect(screen, float32(dm.X), float32(dm.Y), cellSize, cellSize, dm.Color, false)
	}
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Top panel - Total Score
	vector.DrawFilledRect(screen, 10, float32(g.topPanelY), float32(g.screenWidth-20), float32(g.topPanelHeight), color.RGBA{R: 80, G: 80, B: 80, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Total Score: %d x %d = %d", g.state.baseScore, g.state.multiplier, g.state.baseScore*g.state.multiplier), 20, g.topPanelY+20)

	// Foreman panel - Money and Run
	vector.DrawFilledRect(screen, 10, float32(g.foremanY), float32(g.screenWidth-20), float32(g.foremanHeight), color.RGBA{R: 100, G: 100, B: 100, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Money: $%d", g.state.money), 20, g.foremanY+20)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Run: %d/%d", g.state.run, g.state.maxRuns), 200, g.foremanY+20)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Round: %d", g.state.currentRound), 200, g.foremanY+50)

	// Bottom Panel
	vector.DrawFilledRect(screen, 10, float32(g.bottomY), float32(g.screenWidth-20), float32(g.bottomHeight), color.RGBA{R: 80, G: 80, B: 80, A: 255}, false)

	// Sell Area
	vector.DrawFilledRect(screen, 10, float32(g.bottomY+10), 120, float32(g.bottomHeight-20), color.RGBA{R: 255, G: 100, B: 100, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, "Sell", 30, g.bottomY+20)

	// Current Round Score
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Round Score: %d", g.state.baseScore), 140, g.bottomY+20)

	// Start/Stop Run Button
	runButtonColor := color.RGBA{R: 100, G: 200, B: 100, A: 255}
	runButtonText := "Start Run"
	if g.state.phase == PhaseRun {
		runButtonColor = color.RGBA{R: 200, G: 100, B: 100, A: 255}
		runButtonText = "Stop Run"
	}
	vector.DrawFilledRect(screen, 250, float32(g.bottomY+10), float32(g.screenWidth-30-250), float32(g.bottomHeight-20), runButtonColor, false)
	ebitenutil.DebugPrintAt(screen, runButtonText, 260, g.bottomY+20)
}

func (g *Game) drawFactoryFloor(screen *ebiten.Image) {
	for r := 0; r < gridRows; r++ {
		for c := 0; c < gridCols; c++ {
			x := g.gridStartX + c*(cellSize+gridMargin)
			y := g.gridStartY + r*(cellSize+gridMargin)
			vector.DrawFilledRect(screen, float32(x), float32(y), cellSize, cellSize, color.RGBA{R: 60, G: 60, B: 60, A: 255}, false)
		}
	}
}

func (g *Game) drawMachines(screen *ebiten.Image) {
	// Machines on the grid
	for col := 0; col < gridCols; col++ {
		for row := 0; row < gridRows; row++ {
			if m := g.state.machinesOnGrid[col*gridRows+row]; m != nil {
				mm := m.(*Machine)
				vector.DrawFilledRect(screen, float32(mm.X), float32(mm.Y), cellSize, cellSize, mm.Color, false)
				if mm.Type == MachineStart {
					ebitenutil.DebugPrintAt(screen, "Start", mm.X+10, mm.Y+20)
				}
				if mm.Type == MachineEnd {
					ebitenutil.DebugPrintAt(screen, "End", mm.X+15, mm.Y+20)
				}
			}
		}
	}

	// Available machines
	for _, m := range g.state.availableMachines {
		mm := m.(*Machine)
		if !mm.IsPlaced {
			vector.DrawFilledRect(screen, float32(mm.X), float32(mm.Y), cellSize, cellSize, mm.Color, false)
		}
	}
}

func (g *Game) drawObjects(screen *ebiten.Image) {
	for _, obj := range g.state.objects {
		objColor := color.RGBA{R: 255, A: 255}
		switch obj.Type {
		case ObjectGreen:
			objColor.G = 255
		case ObjectBlue:
			objColor.B = 255
		}
		vector.DrawFilledCircle(screen, float32(obj.X), float32(obj.Y), 10, objColor, false)
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.width = outsideWidth
	g.height = outsideHeight
	g.calculateLayout()

	// Update machine positions
	startX, startY := 3, 5
	endX, endY := 1, 3

	if g.state.startMachine != nil {
		g.state.startMachine.(*Machine).X = g.gridStartX + startX*(cellSize+gridMargin)
		g.state.startMachine.(*Machine).Y = g.gridStartY + startY*(cellSize+gridMargin)
	}

	if g.state.endMachine != nil {
		g.state.endMachine.(*Machine).X = g.gridStartX + endX*(cellSize+gridMargin)
		g.state.endMachine.(*Machine).Y = g.gridStartY + endY*(cellSize+gridMargin)
	}

	// Update grid machines
	for i := 0; i < gridCols*gridRows; i++ {
		if m := g.state.machinesOnGrid[i]; m != nil {
			col := i / gridRows
			row := i % gridRows
			m.(*Machine).X = g.gridStartX + col*(cellSize+gridMargin)
			m.(*Machine).Y = g.gridStartY + row*(cellSize+gridMargin)
		}
	}

	// Update available machines
	if len(g.state.availableMachines) > 0 {
		g.state.availableMachines[0].(*Machine).X = g.gridStartX
		g.state.availableMachines[0].(*Machine).Y = g.availableY
	}
	if len(g.state.availableMachines) > 1 {
		g.state.availableMachines[1].(*Machine).X = g.gridStartX + cellSize + gridMargin
		g.state.availableMachines[1].(*Machine).Y = g.availableY
	}

	return outsideWidth, outsideHeight
}
