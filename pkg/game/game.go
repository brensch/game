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
	Type         ObjectType
	Effects      []EffectInterface
}

// MachineType represents the different kinds of machines.
type MachineType int

const (
	MachineConveyor MachineType = iota
	MachineProcessor
	MachineStart
	MachineEnd
)

// Orientation represents the direction a machine is facing.
type Orientation int

const (
	OrientationNorth Orientation = iota
	OrientationEast
	OrientationSouth
	OrientationWest
)

// GamePhase represents the current state of the game (building or running).
type GamePhase int

const (
	PhaseBuild GamePhase = iota
	PhaseRun
)

// Animation represents a moving object animation.
type Animation struct {
	StartX, StartY float64
	EndX, EndY     float64
	Color          color.RGBA
	Duration       float64
	Elapsed        float64
}

// GameState holds all information about the current state of the game.
type GameState struct {
	phase             GamePhase
	money             int
	run               int
	maxRuns           int
	baseScore         int
	multiplier        int
	machines          []*MachineState
	availableMachines []*MachineState
	objects           []*Object
	round             int
	mousePressed      bool
	pressX, pressY    int
	running           bool
	animations        []*Animation
	animationTick     int
	animationSpeed    float64
}

// Game implements ebiten.Game.
type Game struct {
	state                                                                    *GameState
	draggingMachine                                                          *MachineState
	width, height                                                            int
	topPanelHeight, foremanHeight, gridHeight, availableHeight, bottomHeight int
	topPanelY, foremanY, gridStartY, availableY, bottomY                     int
	screenWidth, gridStartX                                                  int
}

func (g *Game) getSelectedMachine() *MachineState {
	for _, ms := range g.state.machines {
		if ms != nil && ms.Selected {
			return ms
		}
	}
	for _, ms := range g.state.availableMachines {
		if ms != nil && ms.Selected {
			return ms
		}
	}
	return nil
}

// NewGame creates a new Game instance.
func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	state := &GameState{
		phase:          PhaseBuild,
		money:          7,
		run:            3,
		maxRuns:        6,
		machines:       make([]*MachineState, gridCols*gridRows),
		round:          1,
		animations:     []*Animation{},
		animationTick:  0,
		animationSpeed: 1.0,
	}

	g := &Game{state: state}
	g.width = 480
	g.height = 800
	g.calculateLayout()

	// Place fixed Start and End machines
	startPos := 0
	endPos := gridCols*gridRows - 1

	state.machines[startPos] = &MachineState{Machine: &Start{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: true, RoundAdded: 0}
	state.machines[endPos] = &MachineState{Machine: &End{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: true, RoundAdded: 0}

	state.availableMachines = []*MachineState{
		{Machine: &Conveyor{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: false, RoundAdded: 0},
		{Machine: &Processor{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: false, RoundAdded: 0},
	}

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

func (g *Game) getMachineAt(cx, cy int) *MachineState {
	for pos, ms := range g.state.machines {
		if ms == nil {
			continue
		}
		col := pos % gridCols
		row := pos / gridCols
		x := g.gridStartX + col*(cellSize+gridMargin)
		y := g.gridStartY + row*(cellSize+gridMargin)
		if cx >= x && cx <= x+cellSize && cy >= y && cy <= y+cellSize {
			return ms
		}
	}
	return nil
}

func (g *Game) getPos(ms *MachineState) int {
	for pos, m := range g.state.machines {
		if m == ms {
			return pos
		}
	}
	return -1
}

// Update proceeds the game state.
func (g *Game) Update() error {
	switch g.state.phase {
	case PhaseBuild:
		g.handleDragAndDrop()
	case PhaseRun:
		if len(g.state.animations) == 0 {
			// Start new tick
			changes, _ := SimulateRun(g.state.machines)
			if g.state.animationTick >= len(changes) {
				// All ticks done
				g.state.phase = PhaseBuild
				g.state.animationTick = 0
				g.state.animationSpeed = 1.0
				break
			}
			tickChanges := changes[g.state.animationTick]
			g.state.animations = []*Animation{}
			for _, ch := range tickChanges {
				if ch.StartObject == nil || ch.EndObject == nil {
					continue
				}
				startGridX := ch.StartObject.GridPosition % gridCols
				startGridY := ch.StartObject.GridPosition / gridCols
				endGridX := ch.EndObject.GridPosition % gridCols
				endGridY := ch.EndObject.GridPosition / gridCols
				startX := float64(g.gridStartX + startGridX*(cellSize+gridMargin) + cellSize/2)
				startY := float64(g.gridStartY + startGridY*(cellSize+gridMargin) + cellSize/2)
				endX := float64(g.gridStartX + endGridX*(cellSize+gridMargin) + cellSize/2)
				endY := float64(g.gridStartY + endGridY*(cellSize+gridMargin) + cellSize/2)
				objColor := color.RGBA{R: 255, A: 255}
				switch ch.StartObject.Type {
				case ObjectGreen:
					objColor.G = 255
				case ObjectBlue:
					objColor.B = 255
				}
				duration := 30.0 / g.state.animationSpeed // frames, decrease over time
				g.state.animations = append(g.state.animations, &Animation{
					StartX: startX, StartY: startY,
					EndX: endX, EndY: endY,
					Color: objColor, Duration: duration, Elapsed: 0,
				})
			}
			g.state.animationTick++
			g.state.animationSpeed += 0.3 // speed up significantly each tick
		}
		// Update animations
		for _, anim := range g.state.animations {
			anim.Elapsed++
		}
		// Remove completed animations
		newAnims := []*Animation{}
		for _, anim := range g.state.animations {
			if anim.Elapsed < anim.Duration {
				newAnims = append(newAnims, anim)
			}
		}
		g.state.animations = newAnims
	}

	// Check for "Start Run" button click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		// Simple button detection for "Start Run"
		if cx > 250 && cx < g.screenWidth-30 && cy > g.bottomY+10 && cy < g.bottomY+10+g.bottomHeight-20 {
			if g.state.phase == PhaseBuild {
				g.state.phase = PhaseRun
				g.state.animations = []*Animation{}
				g.state.animationTick = 0
				g.state.animationSpeed = 1.0
			}
		}
	}
	return nil
}

func (g *Game) handleDragAndDrop() {
	cx, cy := ebiten.CursorPosition()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.state.mousePressed = true
		g.state.pressX, g.state.pressY = cx, cy

		// Check rotation buttons first
		counterclockwiseX := g.screenWidth - 2*cellSize - gridMargin
		counterclockwiseY := g.availableY
		if cx >= counterclockwiseX && cx <= counterclockwiseX+cellSize && cy >= counterclockwiseY && cy <= counterclockwiseY+cellSize {
			selected := g.getSelectedMachine()
			if selected != nil {
				selected.Orientation = (selected.Orientation + 3) % 4
			}
			return
		}
		clockwiseX := g.screenWidth - cellSize
		clockwiseY := g.availableY
		if cx >= clockwiseX && cx <= clockwiseX+cellSize && cy >= clockwiseY && cy <= clockwiseY+cellSize {
			selected := g.getSelectedMachine()
			if selected != nil {
				selected.Orientation = (selected.Orientation + 1) % 4
			}
			return
		}

		// Deselect all
		for _, m := range g.state.machines {
			if m != nil {
				m.Selected = false
			}
		}
		for _, m := range g.state.availableMachines {
			if m != nil {
				m.Selected = false
			}
		}

		// Check if picking from available
		for i, ms := range g.state.availableMachines {
			x := g.gridStartX + i*(cellSize+gridMargin)
			y := g.availableY
			if cx >= x && cx <= x+cellSize && cy >= y && cy <= y+cellSize {
				ms.Selected = true
				break
			}
		}

		// Check if picking placed machine
		ms := g.getMachineAt(cx, cy)
		if ms != nil {
			ms.Selected = true
		}

		// Check run button
		runButtonX := 250
		runButtonY := g.bottomY + 10
		runButtonWidth := g.screenWidth - 30 - 250
		runButtonHeight := g.bottomHeight - 20
		if cx >= runButtonX && cx <= runButtonX+runButtonWidth && cy >= runButtonY && cy <= runButtonY+runButtonHeight {
			if !g.state.running {
				g.state.running = true
				changes, _ := SimulateRun(g.state.machines)
				for i, tickChanges := range changes {
					fmt.Printf("Tick %d: %d changes\n", i, len(tickChanges))
					for _, ch := range tickChanges {
						startStr := "nil"
						if ch.StartObject != nil {
							startStr = fmt.Sprintf("pos %d type %d", ch.StartObject.GridPosition, ch.StartObject.Type)
						}
						endStr := "nil"
						if ch.EndObject != nil {
							endStr = fmt.Sprintf("pos %d type %d", ch.EndObject.GridPosition, ch.EndObject.Type)
						}
						fmt.Printf("  Change: Start %s -> End %s\n", startStr, endStr)
					}
				}
				g.state.running = false
				g.state.round++
				g.state.run++
			}
		}
	}

	if g.state.mousePressed {
		dx := cx - g.state.pressX
		dy := cy - g.state.pressY
		if dx*dx+dy*dy > 1000 { // threshold
			selected := g.getSelectedMachine()
			if selected != nil {
				if selected.IsPlaced {
					g.draggingMachine = selected
					pos := g.getPos(g.draggingMachine)
					g.state.machines[pos] = nil
				} else {
					// from available
					g.draggingMachine = &MachineState{Machine: selected.Machine, Orientation: selected.Orientation, BeingDragged: true, IsPlaced: false, RoundAdded: g.state.round, Selected: true}
					selected.Selected = false
				}
				g.draggingMachine.BeingDragged = true
			}
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if g.draggingMachine != nil {
			// Place at cursor position
			gridX, gridY := -1, -1
			for r := 0; r < gridRows; r++ {
				for c := 0; c < gridCols; c++ {
					x := g.gridStartX + c*(cellSize+gridMargin)
					y := g.gridStartY + r*(cellSize+gridMargin)
					if cx >= x && cx <= x+cellSize && cy >= y && cy <= y+cellSize {
						position := r*gridCols + c
						if g.state.machines[position] == nil {
							gridX, gridY = c, r
						}
						break
					}
				}
				if gridX != -1 {
					break
				}
			}
			if gridX != -1 {
				if !g.draggingMachine.IsPlaced {
					g.state.money -= 1
				}
				g.draggingMachine.IsPlaced = true
				g.draggingMachine.RoundAdded = g.state.round
				position := gridY*gridCols + gridX
				g.state.machines[position] = g.draggingMachine
			}
			g.draggingMachine.BeingDragged = false
			g.draggingMachine = nil
		}
		g.state.mousePressed = false
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
	if g.draggingMachine != nil {
		cx, cy := ebiten.CursorPosition()
		vector.DrawFilledRect(screen, float32(cx-cellSize/2), float32(cy-cellSize/2), cellSize, cellSize, g.draggingMachine.Machine.GetColor(), false)
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
	if g.state.running {
		runButtonColor = color.RGBA{R: 200, G: 200, B: 100, A: 255}
		runButtonText = "Running"
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

func (g *Game) drawArrow(screen *ebiten.Image, x, y float32, orientation Orientation) {
	arrowColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	switch orientation {
	case OrientationNorth:
		vector.StrokeLine(screen, x+30, y+50, x+30, y+10, 1, arrowColor, false)
		vector.StrokeLine(screen, x+20, y+20, x+30, y+10, 1, arrowColor, false)
		vector.StrokeLine(screen, x+40, y+20, x+30, y+10, 1, arrowColor, false)
	case OrientationEast:
		vector.StrokeLine(screen, x+10, y+30, x+50, y+30, 1, arrowColor, false)
		vector.StrokeLine(screen, x+40, y+20, x+50, y+30, 1, arrowColor, false)
		vector.StrokeLine(screen, x+40, y+40, x+50, y+30, 1, arrowColor, false)
	case OrientationSouth:
		vector.StrokeLine(screen, x+30, y+10, x+30, y+50, 1, arrowColor, false)
		vector.StrokeLine(screen, x+20, y+40, x+30, y+50, 1, arrowColor, false)
		vector.StrokeLine(screen, x+40, y+40, x+30, y+50, 1, arrowColor, false)
	case OrientationWest:
		vector.StrokeLine(screen, x+50, y+30, x+10, y+30, 1, arrowColor, false)
		vector.StrokeLine(screen, x+20, y+20, x+10, y+30, 1, arrowColor, false)
		vector.StrokeLine(screen, x+20, y+40, x+10, y+30, 1, arrowColor, false)
	}
}

func (g *Game) drawMachines(screen *ebiten.Image) {
	// Machines on the grid
	for pos, ms := range g.state.machines {
		if ms == nil || ms.Machine == nil {
			continue
		}
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
		g.drawArrow(screen, float32(x), float32(y), ms.Orientation)
		if ms.Selected {
			vector.StrokeRect(screen, float32(x), float32(y), cellSize, cellSize, 3, color.RGBA{R: 255, G: 255, B: 0, A: 255}, false)
		}
	}

	// Available machines
	for i, ms := range g.state.availableMachines {
		if ms != nil && !ms.BeingDragged && ms.Machine != nil {
			x := g.gridStartX + i*(cellSize+gridMargin)
			y := g.availableY
			vector.DrawFilledRect(screen, float32(x), float32(y), cellSize, cellSize, ms.Machine.GetColor(), false)
		}
	}

	// Rotation buttons
	counterclockwiseX := g.screenWidth - 2*cellSize - gridMargin
	counterclockwiseY := g.availableY
	vector.DrawFilledCircle(screen, float32(counterclockwiseX+cellSize/2), float32(counterclockwiseY+cellSize/2), cellSize/2, color.RGBA{R: 200, G: 100, B: 100, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, "<-", counterclockwiseX+22, counterclockwiseY+26)

	clockwiseX := g.screenWidth - cellSize
	clockwiseY := g.availableY
	vector.DrawFilledCircle(screen, float32(clockwiseX+cellSize/2), float32(clockwiseY+cellSize/2), cellSize/2, color.RGBA{R: 100, G: 100, B: 200, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, "->", clockwiseX+22, clockwiseY+26)
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

	// Draw animations
	for _, anim := range g.state.animations {
		progress := anim.Elapsed / anim.Duration
		if progress > 1 {
			progress = 1
		}
		x := anim.StartX + (anim.EndX-anim.StartX)*progress
		y := anim.StartY + (anim.EndY-anim.StartY)*progress
		vector.DrawFilledCircle(screen, float32(x), float32(y), 10, anim.Color, false)
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.width = outsideWidth
	g.height = outsideHeight
	g.calculateLayout()

	return outsideWidth, outsideHeight
}
