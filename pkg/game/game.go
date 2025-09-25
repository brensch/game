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

// Object represents an item moving through the factory.
type Object struct {
	GridPosition int
	X, Y         float64
	Type         ObjectType
	pathIndex    int
}

// MachineType represents the different kinds of machines.
type MachineType int

const (
	MachineConveyor MachineType = iota
	MachineProcessor
	MachineStart
	MachineEnd
)

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

// GamePhase represents the current state of the game (building or running).
type GamePhase int

const (
	PhaseBuild GamePhase = iota
	PhaseRun
)

// GameState holds all information about the current state of the game.
type GameState struct {
	phase                    GamePhase
	money                    int
	run                      int
	maxRuns                  int
	baseScore                int
	multiplier               int
	machines                 map[int]*MachineState
	availableMachines        []*Machine
	objects                  []*Object
	draggingMachine          *Machine
	dragOffsetX, dragOffsetY int
	startMachine             *Machine
	endMachine               *Machine
	round                    int
}

// Game implements ebiten.Game.
type Game struct {
	state                                                                    *GameState
	machines                                                                 []*Machine
	width, height                                                            int
	topPanelHeight, foremanHeight, gridHeight, availableHeight, bottomHeight int
	topPanelY, foremanY, gridStartY, availableY, bottomY                     int
	screenWidth, gridStartX                                                  int
}

// NewGame creates a new Game instance.
func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	state := &GameState{
		phase:    PhaseBuild,
		money:    7,
		run:      3,
		maxRuns:  6,
		machines: make(map[int]*MachineState),
		round:    1,
	}

	g := &Game{state: state}
	g.width = 480
	g.height = 800
	g.calculateLayout()

	// Place fixed Start and End machines
	startPos := 0
	endPos := gridCols*gridRows - 1

	state.startMachine = &Machine{Type: MachineStart, GridX: startPos % gridCols, GridY: startPos / gridCols, IsPlaced: true, RoundAdded: 0, Color: color.RGBA{R: 150, G: 255, B: 150, A: 255}}
	state.endMachine = &Machine{Type: MachineEnd, GridX: endPos % gridCols, GridY: endPos / gridCols, IsPlaced: true, RoundAdded: 0, Color: color.RGBA{R: 255, G: 150, B: 150, A: 255}}

	state.machines[startPos] = &MachineState{Machine: &Start{}, IsPlaced: true, RoundAdded: 0}
	state.machines[endPos] = &MachineState{Machine: &End{}, IsPlaced: true, RoundAdded: 0}

	g.machines = []*Machine{
		&Machine{Type: MachineConveyor, X: 0, Y: 0, IsDraggable: true, IsPlaced: false, Color: color.RGBA{R: 200, G: 200, B: 200, A: 255}},
		&Machine{Type: MachineProcessor, X: 0, Y: 0, IsDraggable: true, IsPlaced: false, Color: color.RGBA{R: 100, G: 200, B: 100, A: 255}},
		state.startMachine,
		state.endMachine,
	}

	state.availableMachines = []*Machine{g.machines[0], g.machines[1]}

	return g
}

func createMachine(mt MachineType) MachineInterface {
	switch mt {
	case MachineConveyor:
		return &Conveyor{}
	case MachineProcessor:
		return &Processor{}
	case MachineStart:
		return &Start{}
	case MachineEnd:
		return &End{}
	}
	return nil
}

func applyChanges(changes []*Change, objects *[]*Object) {
	for _, change := range changes {
		switch change.Type {
		case ChangeTypeCreate:
			*objects = append(*objects, &Object{GridPosition: change.GridPosition, Type: change.ObjectType})
		case ChangeTypeMove:
			for i, obj := range *objects {
				if obj.GridPosition == change.FromPosition {
					(*objects)[i].GridPosition = change.ToPosition
					break
				}
			}
		case ChangeTypeDelete:
			for i, obj := range *objects {
				if obj.GridPosition == change.GridPosition {
					*objects = append((*objects)[:i], (*objects)[i+1:]...)
					break
				}
			}
		}
	}
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
			} else {
				g.state.phase = PhaseBuild
				g.state.objects = nil // Clear objects
				g.state.baseScore = 0 // Reset score
				g.state.round++
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
			if !m.IsPlaced && cx >= m.X && cx <= m.X+cellSize && cy >= m.Y && cy <= m.Y+cellSize {
				g.state.draggingMachine = m
				g.state.dragOffsetX = cx - m.X
				g.state.dragOffsetY = cy - m.Y
				break
			}
		}

		// Check if picking up from grid (only current round machines)
		if g.state.draggingMachine == nil {
			for col := 0; col < gridCols; col++ {
				for row := 0; row < gridRows; row++ {
					position := col*gridRows + row
					if ms, ok := g.state.machines[position]; ok {
						if ms.RoundAdded == g.state.round {
							// Find the UI machine
							for _, m := range g.machines {
								if m.GridX == col && m.GridY == row && m.IsPlaced {
									g.state.draggingMachine = m
									g.state.dragOffsetX = cx - m.X
									g.state.dragOffsetY = cy - m.Y
									delete(g.state.machines, position)
									m.IsPlaced = false
									break
								}
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
		dm := g.state.draggingMachine
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

				if _, ok := g.state.machines[col*gridRows+row]; !ok {
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
				dm.IsPlaced = true
				dm.RoundAdded = g.state.round
				position := gridY*gridCols + gridX
				g.state.machines[position] = &MachineState{Machine: createMachine(dm.Type), IsPlaced: true, RoundAdded: g.state.round}
				g.state.draggingMachine = nil
			} else {
				// Dropped elsewhere, discard
				g.state.draggingMachine = nil
			}
		}
	}
}

func (g *Game) updateRun() {
	var changes []*Change
	for pos, ms := range g.state.machines {
		if change := ms.Machine.Process(pos, g.state.objects, g.state.round); change != nil {
			changes = append(changes, change)
		}
	}
	applyChanges(changes, &g.state.objects)

	// Remove invalid objects
	for i := 0; i < len(g.state.objects); i++ {
		obj := g.state.objects[i]
		if obj.GridPosition < 0 || obj.GridPosition >= gridCols*gridRows {
			g.state.objects = append(g.state.objects[:i], g.state.objects[i+1:]...)
			i--
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
		dm := g.state.draggingMachine
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
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Round: %d", g.state.round), 200, g.foremanY+50)

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
	for pos, ms := range g.state.machines {
		col := pos % gridCols
		row := pos / gridCols
		x := g.gridStartX + col*(cellSize+gridMargin)
		y := g.gridStartY + row*(cellSize+gridMargin)
		vector.DrawFilledRect(screen, float32(x), float32(y), cellSize, cellSize, ms.Machine.GetColor(), false)
		if ms.Machine.GetType() == MachineStart {
			ebitenutil.DebugPrintAt(screen, "Start", int(x)+10, int(y)+20)
		}
		if ms.Machine.GetType() == MachineEnd {
			ebitenutil.DebugPrintAt(screen, "End", int(x)+15, int(y)+20)
		}
	}

	// Available machines
	for _, m := range g.state.availableMachines {
		if !m.IsPlaced {
			vector.DrawFilledRect(screen, float32(m.X), float32(m.Y), cellSize, cellSize, m.Color, false)
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
		gridX := obj.GridPosition % gridCols
		gridY := obj.GridPosition / gridCols
		x := g.gridStartX + gridX*(cellSize+gridMargin) + cellSize/2
		y := g.gridStartY + gridY*(cellSize+gridMargin) + cellSize/2
		vector.DrawFilledCircle(screen, float32(x), float32(y), 10, objColor, false)
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.width = outsideWidth
	g.height = outsideHeight
	g.calculateLayout()

	// Update machine positions
	for _, m := range g.machines {
		if m.IsPlaced {
			m.X = g.gridStartX + m.GridX*(cellSize+gridMargin)
			m.Y = g.gridStartY + m.GridY*(cellSize+gridMargin)
		}
	}

	// Update available machines
	if len(g.state.availableMachines) > 0 {
		g.state.availableMachines[0].X = g.gridStartX
		g.state.availableMachines[0].Y = g.availableY
	}
	if len(g.state.availableMachines) > 1 {
		g.state.availableMachines[1].X = g.gridStartX + cellSize + gridMargin
		g.state.availableMachines[1].Y = g.availableY
	}

	return outsideWidth, outsideHeight
}
