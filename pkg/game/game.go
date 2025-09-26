package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	topPanelHeight  = 80
	foremanHeight   = 80
	availableHeight = 60
	bottomHeight    = 60
	minGap          = 10

	gridCols    = 9
	gridRows    = 9
	displayCols = 7
	displayRows = 7
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
	MachineMiner
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func manhattan(p1, p2 int) int {
	c1 := p1 % gridCols
	r1 := p1 / gridCols
	c2 := p2 % gridCols
	r2 := p2 / gridCols
	return abs(r1-r2) + abs(c1-c2)
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
	inputPressed      bool
	pressX, pressY    int
	running           bool
	animations        []*Animation
	animationTick     int
	animationSpeed    float64
}

// Game implements ebiten.Game.
type Game struct {
	state                                                        *GameState
	width, height                                                int
	topPanelHeight, foremanHeight, availableHeight, bottomHeight int
	topPanelY, foremanY, gridStartY, availableY, bottomY         int
	screenWidth, gridStartX                                      int
	cellSize, gridMargin                                         int

	vignetteImage *ebiten.Image
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

func (g *Game) getDraggingMachine() *MachineState {
	for _, ms := range g.state.machines {
		if ms != nil && ms.BeingDragged {
			return ms
		}
	}
	for _, ms := range g.state.availableMachines {
		if ms != nil && ms.BeingDragged {
			return ms
		}
	}
	return nil
}

// NewGame creates a new Game instance.
func NewGame(width, height int) *Game {
	state := &GameState{
		phase:          PhaseBuild,
		money:          7,
		run:            1,
		maxRuns:        6,
		machines:       make([]*MachineState, gridCols*gridRows),
		round:          1,
		animations:     []*Animation{},
		animationTick:  0,
		animationSpeed: 1.0,
	}

	g := &Game{state: state}
	g.width = width
	g.height = height
	g.calculateLayout()

	// Place random End machine
	endRow := 1 + rand.Intn(displayRows)
	endCol := 1 + rand.Intn(displayCols)
	endPos := endRow*gridCols + endCol

	state.machines[endPos] = &MachineState{Machine: &End{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: true, RunAdded: 0, OriginalPos: endPos}

	state.availableMachines = []*MachineState{
		{Machine: &Conveyor{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: false, RunAdded: 0},
		{Machine: &Processor{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: false, RunAdded: 0},
		{Machine: &Miner{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: false, RunAdded: 0},
	}

	return g
}

func (g *Game) calculateLayout() {
	g.topPanelHeight = topPanelHeight
	g.foremanHeight = foremanHeight
	g.availableHeight = availableHeight
	g.bottomHeight = bottomHeight

	marginRatio := 1.0 / 6.0
	availableWidth := g.screenWidth - 40
	availableHeight := g.height - (g.topPanelHeight + g.foremanHeight + g.availableHeight + g.bottomHeight + 4*minGap)
	widthFactor := float64(displayCols) + float64(displayCols-1)*marginRatio
	heightFactor := float64(displayRows) + float64(displayRows-1)*marginRatio
	cellSizeW := int(float64(availableWidth) / widthFactor)
	cellSizeH := int(float64(availableHeight) / heightFactor)
	g.cellSize = cellSizeW
	if cellSizeH < g.cellSize {
		g.cellSize = cellSizeH
	}
	if g.cellSize < 10 {
		g.cellSize = 10
	}
	g.gridMargin = int(float64(g.cellSize) * marginRatio)

	gridHeight := displayRows*g.cellSize + (displayRows-1)*g.gridMargin
	totalFixedHeight := g.topPanelHeight + g.foremanHeight + gridHeight + g.availableHeight + g.bottomHeight
	gap := (g.height - totalFixedHeight) / 5
	if gap < minGap {
		gap = minGap
	}
	g.topPanelY = gap
	g.foremanY = g.topPanelY + g.topPanelHeight + gap
	g.gridStartY = g.foremanY + g.foremanHeight + gap
	g.availableY = g.gridStartY + gridHeight + gap
	g.bottomY = g.availableY + g.availableHeight + gap
	g.screenWidth = g.width
	g.gridStartX = (g.screenWidth - (displayCols*g.cellSize + (displayCols-1)*g.gridMargin)) / 2
}

func (g *Game) getMachineAt(cx, cy int) *MachineState {
	col := (cx - g.gridStartX) / (g.cellSize + g.gridMargin)
	row := (cy - g.gridStartY) / (g.cellSize + g.gridMargin)
	if col < 0 || col >= displayCols || row < 0 || row >= displayRows {
		return nil
	}
	internalCol := col + 1
	internalRow := row + 1
	pos := internalRow*gridCols + internalCol
	if pos < len(g.state.machines) && g.state.machines[pos] != nil {
		return g.state.machines[pos]
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
				g.state.run++
				if g.state.run > g.state.maxRuns {
					g.state.run = 1
					g.state.round++
				}
				// Move end to random location up to 2 squares away
				for pos, ms := range g.state.machines {
					if ms != nil && ms.Machine.GetType() == MachineEnd {
						currentPos := pos
						var candidates []int
						cr := currentPos / gridCols
						cc := currentPos % gridCols
						for dr := -2; dr <= 2; dr++ {
							for dc := -2; dc <= 2; dc++ {
								if abs(dr)+abs(dc) > 2 || (dr == 0 && dc == 0) {
									continue
								}
								nr := cr + dr
								nc := cc + dc
								if nr >= 1 && nr <= displayRows && nc >= 1 && nc <= displayCols {
									npos := nr*gridCols + nc
									if g.state.machines[npos] == nil {
										candidates = append(candidates, npos)
									}
								}
							}
						}
						if len(candidates) > 0 {
							newPos := candidates[rand.Intn(len(candidates))]
							g.state.machines[newPos] = ms
							g.state.machines[currentPos] = nil
						}
						break
					}
				}
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
				startX := float64(g.gridStartX + (startGridX-1)*(g.cellSize+g.gridMargin) + g.cellSize/2)
				startY := float64(g.gridStartY + (startGridY-1)*(g.cellSize+g.gridMargin) + g.cellSize/2)
				endX := float64(g.gridStartX + (endGridX-1)*(g.cellSize+g.gridMargin) + g.cellSize/2)
				endY := float64(g.gridStartY + (endGridY-1)*(g.cellSize+g.gridMargin) + g.cellSize/2)
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
	input := GetUnifiedInput()
	if input.JustPressed {
		cx, cy := input.X, input.Y
		// Simple button detection for "Start Run"
		if cx > 250 && cx < g.screenWidth-30 && cy > g.bottomY+10 && cy < g.bottomY+10+g.bottomHeight-20 {
			if g.state.phase == PhaseBuild {
				g.state.phase = PhaseRun
				g.state.animations = []*Animation{}
				g.state.animationTick = 0
				g.state.animationSpeed = 1.0
			}
		}
		// Check for "Restart" button click
		if cx >= g.screenWidth-100 && cx <= g.screenWidth-20 && cy >= g.topPanelY+10 && cy <= g.topPanelY+10+g.topPanelHeight-20 {
			// Reset game state
			g.state = &GameState{
				phase:          PhaseBuild,
				money:          7,
				run:            1,
				maxRuns:        6,
				machines:       make([]*MachineState, gridCols*gridRows),
				round:          1,
				animations:     []*Animation{},
				animationTick:  0,
				animationSpeed: 1.0,
			}
			// Place random End machine
			endRow := 1 + rand.Intn(displayRows)
			endCol := 1 + rand.Intn(displayCols)
			endPos := endRow*gridCols + endCol
			g.state.machines[endPos] = &MachineState{Machine: &End{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: true, RunAdded: 0, OriginalPos: endPos}
			g.state.availableMachines = []*MachineState{
				{Machine: &Conveyor{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: false, RunAdded: 0},
				{Machine: &Processor{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: false, RunAdded: 0},
				{Machine: &Miner{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: false, RunAdded: 0},
			}
		}
	}
	return nil
}

func (g *Game) handleDragAndDrop() {
	input := GetUnifiedInput()
	cx, cy := input.X, input.Y

	if input.JustPressed {
		g.state.inputPressed = true
		g.state.pressX, g.state.pressY = cx, cy

		// Check rotation buttons first
		counterclockwiseX := g.screenWidth - 2*g.cellSize - g.gridMargin
		counterclockwiseY := g.availableY
		if cx >= counterclockwiseX && cx <= counterclockwiseX+g.cellSize && cy >= counterclockwiseY && cy <= counterclockwiseY+g.cellSize {
			selected := g.getSelectedMachine()
			if selected != nil {
				selected.Orientation = (selected.Orientation + 3) % 4
			}
			return
		}
		clockwiseX := g.screenWidth - g.cellSize
		clockwiseY := g.availableY
		if cx >= clockwiseX && cx <= clockwiseX+g.cellSize && cy >= clockwiseY && cy <= clockwiseY+g.cellSize {
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
			x := g.gridStartX + i*(g.cellSize+g.gridMargin)
			y := g.availableY
			if cx >= x && cx <= x+g.cellSize && cy >= y && cy <= y+g.cellSize {
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
			}
		}
	}

	if g.state.inputPressed && input.Pressed {
		dx := cx - g.state.pressX
		dy := cy - g.state.pressY
		if dx*dx+dy*dy > 1000 { // threshold
			selected := g.getSelectedMachine()
			if selected != nil {
				if selected.IsPlaced && selected.Machine.GetType() != MachineEnd {
					selected.BeingDragged = true
					pos := g.getPos(selected)
					selected.OriginalPos = pos
				} else if !selected.IsPlaced {
					// from available
					selected.BeingDragged = true
					selected.RunAdded = g.state.run
				}
			}
		}
	}

	if input.JustReleased {
		dragging := g.getDraggingMachine()
		if dragging != nil {
			// Place at cursor position
			gridX, gridY := -1, -1
			for r := 0; r < displayRows; r++ {
				for c := 0; c < displayCols; c++ {
					x := g.gridStartX + c*(g.cellSize+g.gridMargin)
					y := g.gridStartY + r*(g.cellSize+g.gridMargin)
					if cx >= x && cx <= x+g.cellSize && cy >= y && cy <= y+g.cellSize {
						position := (r+1)*gridCols + (c + 1)
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
				var placedMS *MachineState
				if !dragging.IsPlaced {
					g.state.money -= 1
					// Create a new instance for placed machine
					newMS := &MachineState{
						Machine:      dragging.Machine,
						Orientation:  dragging.Orientation,
						BeingDragged: false,
						IsPlaced:     true,
						RunAdded:     g.state.run,
					}
					position := (gridY+1)*gridCols + (gridX + 1)
					g.state.machines[position] = newMS
					placedMS = newMS
				} else {
					// Moving existing placed machine
					position := (gridY+1)*gridCols + (gridX + 1)
					g.state.machines[position] = dragging
					if position != dragging.OriginalPos {
						g.state.machines[dragging.OriginalPos] = nil
					}
					placedMS = dragging
				}
				// Select the newly placed machine
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
				// Select the placed one
				placedMS.Selected = true
			} else {
				// Check if over sell area
				sellX := 10
				sellY := g.bottomY + 10
				sellWidth := 120
				sellHeight := g.bottomHeight - 20
				if cx >= sellX && cx <= sellX+sellWidth && cy >= sellY && cy <= sellY+sellHeight {
					if dragging.IsPlaced && dragging.Machine.GetType() != MachineMiner && dragging.Machine.GetType() != MachineEnd {
						// Sell the machine
						g.state.money += 1
						// Remove from grid
						for pos, ms := range g.state.machines {
							if ms == dragging {
								g.state.machines[pos] = nil
								break
							}
						}
					}
				}
			}
			dragging.BeingDragged = false
		}
		g.state.inputPressed = false
	}
}

// Draw draws the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 40, G: 40, B: 40, A: 255})
	g.drawUI(screen)
	g.drawFactoryFloor(screen)
	g.drawMachines(screen)
	g.drawObjects(screen)
	g.drawTooltips(screen)

	// Draw the dragging machine on top
	dragging := g.getDraggingMachine()
	if dragging != nil {
		cx, cy := GetCursorPosition()
		vector.DrawFilledRect(screen, float32(cx-g.cellSize/2), float32(cy-g.cellSize/2), float32(g.cellSize), float32(g.cellSize), dragging.Machine.GetColor(), false)
	}

	// Apply CRT effects
	g.drawScanlines(screen)
	// Now, draw the vignette overlay on top of everything.
	if g.vignetteImage != nil {
		op := &ebiten.DrawImageOptions{}
		screen.DrawImage(g.vignetteImage, op)
	}
}

func (g *Game) drawScanlines(screen *ebiten.Image) {
	bounds := screen.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	for y := 0; y < h; y += 2 {
		vector.DrawFilledRect(screen, 0, float32(y), float32(w), 1, color.RGBA{R: 0, G: 0, B: 0, A: 30}, false)
	}
}

// createVignetteImage generates a new image containing a smooth vignette effect.
// w, h: The dimensions of your screen.
// strength: How dark the vignette is (e.g., 0.5 for a moderate effect).
// falloff: How sharply the vignette fades (e.g., 1.5 for a tighter fade).
func createVignetteImage(w, h int, strength, falloff float64) *ebiten.Image {
	// Create a new blank image to draw our vignette on.
	vignette := ebiten.NewImage(w, h)

	bounds := vignette.Bounds()
	centerX := float64(bounds.Dx()) / 2.0
	centerY := float64(bounds.Dy()) / 2.0

	// The max distance is from the center to a corner.
	maxDist := math.Hypot(centerX, centerY)

	// We'll modify the raw pixel data directly for performance.
	pixels := make([]byte, 4*w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Calculate the distance of the current pixel from the center.
			dist := math.Hypot(float64(x)-centerX, float64(y)-centerY)

			// Normalize the distance to a 0.0-1.0 range.
			ratio := dist / maxDist

			// Use Pow to create a smoother, more natural falloff.
			// The falloff parameter controls the curve of the gradient.
			ratio = math.Pow(ratio, falloff)

			// Calculate the alpha based on the ratio and desired strength.
			// Clamp the alpha value to a max of 255.
			alpha := math.Min(ratio*255*strength, 255)

			// The vignette color is black (0, 0, 0). We only modify the alpha.
			// The pixel array is a flat slice: [R, G, B, A, R, G, B, A, ...]
			idx := (y*w + x) * 4
			pixels[idx+0] = 0           // R
			pixels[idx+1] = 0           // G
			pixels[idx+2] = 0           // B
			pixels[idx+3] = byte(alpha) // A
		}
	}

	// Apply the calculated pixel data to our image.
	vignette.WritePixels(pixels)

	return vignette
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Top panel - Total Score
	vector.DrawFilledRect(screen, 10, float32(g.topPanelY), float32(g.screenWidth-20), float32(g.topPanelHeight), color.RGBA{R: 80, G: 80, B: 80, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Total Score: %d x %d = %d", g.state.baseScore, g.state.multiplier, g.state.baseScore*g.state.multiplier), 20, g.topPanelY+20)

	// Restart button
	vector.DrawFilledRect(screen, float32(g.screenWidth-100), float32(g.topPanelY+10), 80, float32(g.topPanelHeight-20), color.RGBA{R: 200, G: 100, B: 100, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, "Restart", g.screenWidth-90, g.topPanelY+30)

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
	for r := 0; r < displayRows; r++ {
		for c := 0; c < displayCols; c++ {
			x := g.gridStartX + c*(g.cellSize+g.gridMargin)
			y := g.gridStartY + r*(g.cellSize+g.gridMargin)
			vector.DrawFilledRect(screen, float32(x), float32(y), float32(g.cellSize), float32(g.cellSize), color.RGBA{R: 60, G: 60, B: 60, A: 255}, false)
		}
	}
}

func (g *Game) drawArrow(screen *ebiten.Image, x, y float32, orientation Orientation) {
	arrowColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	arrowSize := float32(g.cellSize / 6)
	centerX := x + float32(g.cellSize)/2
	shaftStartY := y + arrowSize
	shaftEndY := y + float32(g.cellSize) - arrowSize
	arrowY := y + 2*arrowSize
	leftX := centerX - arrowSize
	rightX := centerX + arrowSize
	shaftY := y + float32(g.cellSize)/2
	shaftLeft := x + arrowSize
	shaftRight := x + float32(g.cellSize) - arrowSize
	arrowX := x + float32(g.cellSize) - 2*arrowSize
	topY := shaftY - arrowSize
	bottomY := shaftY + arrowSize
	switch orientation {
	case OrientationNorth:
		vector.StrokeLine(screen, centerX, shaftEndY, centerX, shaftStartY, 1, arrowColor, false)
		vector.StrokeLine(screen, leftX, arrowY, centerX, shaftStartY, 1, arrowColor, false)
		vector.StrokeLine(screen, rightX, arrowY, centerX, shaftStartY, 1, arrowColor, false)
	case OrientationEast:
		vector.StrokeLine(screen, shaftLeft, shaftY, shaftRight, shaftY, 1, arrowColor, false)
		vector.StrokeLine(screen, arrowX, topY, shaftRight, shaftY, 1, arrowColor, false)
		vector.StrokeLine(screen, arrowX, bottomY, shaftRight, shaftY, 1, arrowColor, false)
	case OrientationSouth:
		vector.StrokeLine(screen, centerX, shaftStartY, centerX, shaftEndY, 1, arrowColor, false)
		vector.StrokeLine(screen, leftX, y+float32(g.cellSize)-2*arrowSize, centerX, shaftEndY, 1, arrowColor, false)
		vector.StrokeLine(screen, rightX, y+float32(g.cellSize)-2*arrowSize, centerX, shaftEndY, 1, arrowColor, false)
	case OrientationWest:
		vector.StrokeLine(screen, shaftRight, shaftY, shaftLeft, shaftY, 1, arrowColor, false)
		vector.StrokeLine(screen, x+2*arrowSize, topY, shaftLeft, shaftY, 1, arrowColor, false)
		vector.StrokeLine(screen, x+2*arrowSize, bottomY, shaftLeft, shaftY, 1, arrowColor, false)
	}
}

func (g *Game) drawMachines(screen *ebiten.Image) {
	// Machines on the grid
	for pos, ms := range g.state.machines {
		if ms == nil || ms.Machine == nil || ms.BeingDragged {
			continue
		}
		col := pos % gridCols
		row := pos / gridCols
		if row < 1 || row > displayRows || col < 1 || col > displayCols {
			continue
		}
		x := g.gridStartX + (col-1)*(g.cellSize+g.gridMargin)
		y := g.gridStartY + (row-1)*(g.cellSize+g.gridMargin)
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(g.cellSize), float32(g.cellSize), ms.Machine.GetColor(), false)

		if ms.Machine.GetType() == MachineEnd {
			ebitenutil.DebugPrintAt(screen, "End", x+g.cellSize/4, y+g.cellSize/3)
		}
		g.drawArrow(screen, float32(x), float32(y), ms.Orientation)
		if ms.Selected {
			vector.StrokeRect(screen, float32(x), float32(y), float32(g.cellSize), float32(g.cellSize), 3, color.RGBA{R: 255, G: 255, B: 0, A: 255}, false)
		}
	}

	// Available machines
	for i, ms := range g.state.availableMachines {
		if ms != nil && !ms.BeingDragged && ms.Machine != nil {
			x := g.gridStartX + i*(g.cellSize+g.gridMargin)
			y := g.availableY
			vector.DrawFilledRect(screen, float32(x), float32(y), float32(g.cellSize), float32(g.cellSize), ms.Machine.GetColor(), false)
		}
	}

	// Rotation buttons
	counterclockwiseX := g.screenWidth - 2*g.cellSize - g.gridMargin
	counterclockwiseY := g.availableY
	vector.DrawFilledCircle(screen, float32(counterclockwiseX+g.cellSize/2), float32(counterclockwiseY+g.cellSize/2), float32(g.cellSize/2), color.RGBA{R: 200, G: 100, B: 100, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, "<-", counterclockwiseX+g.cellSize/4, counterclockwiseY+g.cellSize/2-3)

	clockwiseX := g.screenWidth - g.cellSize
	clockwiseY := g.availableY
	vector.DrawFilledCircle(screen, float32(clockwiseX+g.cellSize/2), float32(clockwiseY+g.cellSize/2), float32(g.cellSize/2), color.RGBA{R: 100, G: 100, B: 200, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, "->", clockwiseX+g.cellSize/4, clockwiseY+g.cellSize/2-3)
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
		if gridY < 1 || gridY > displayRows || gridX < 1 || gridX > displayCols {
			continue
		}
		x := g.gridStartX + (gridX-1)*(g.cellSize+g.gridMargin) + g.cellSize/2
		y := g.gridStartY + (gridY-1)*(g.cellSize+g.gridMargin) + g.cellSize/2
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

func (g *Game) drawTooltips(screen *ebiten.Image) {
	cx, cy := GetCursorPosition()

	// Check for selected machine first
	selected := g.getSelectedMachine()
	if selected != nil {
		// Find position of selected machine
		for pos, ms := range g.state.machines {
			if ms == selected {
				col := pos % gridCols
				row := pos / gridCols
				if row >= 1 && row <= displayRows && col >= 1 && col <= displayCols {
					x := g.gridStartX + (col-1)*(g.cellSize+g.gridMargin) + g.cellSize/2
					y := g.gridStartY + (row-1)*(g.cellSize+g.gridMargin)
					g.drawTooltip(screen, selected.Machine.GetDescription(), x, y-10)
					return
				}
			}
		}
		// Check available machines
		for i, ms := range g.state.availableMachines {
			if ms == selected {
				x := g.gridStartX + i*(g.cellSize+g.gridMargin) + g.cellSize/2
				y := g.availableY + g.cellSize
				g.drawTooltip(screen, selected.Machine.GetDescription(), x, y+10)
				return
			}
		}
	}

	// Check for hover on grid machines
	if ms := g.getMachineAt(cx, cy); ms != nil {
		col := (cx - g.gridStartX) / (g.cellSize + g.gridMargin)
		row := (cy - g.gridStartY) / (g.cellSize + g.gridMargin)
		x := g.gridStartX + col*(g.cellSize+g.gridMargin) + g.cellSize/2
		y := g.gridStartY + row*(g.cellSize+g.gridMargin)
		g.drawTooltip(screen, ms.Machine.GetDescription(), x, y-10)
		return
	}

	// Check for hover on available machines
	for i, ms := range g.state.availableMachines {
		x := g.gridStartX + i*(g.cellSize+g.gridMargin)
		y := g.availableY
		if cx >= x && cx <= x+g.cellSize && cy >= y && cy <= y+g.cellSize {
			g.drawTooltip(screen, ms.Machine.GetDescription(), x+g.cellSize/2, y+g.cellSize+10)
			return
		}
	}
}

func (g *Game) drawTooltip(screen *ebiten.Image, text string, x, y int) {
	// Measure text width (approximate)
	textWidth := len(text) * 6 // rough estimate
	boxWidth := textWidth + 20
	boxHeight := 30

	// Position box above the point
	boxX := x - boxWidth/2
	boxY := y - boxHeight - 5

	// Ensure box stays on screen
	if boxX < 10 {
		boxX = 10
	}
	if boxX+boxWidth > g.screenWidth-10 {
		boxX = g.screenWidth - boxWidth - 10
	}
	if boxY < 10 {
		boxY = y + 15 // show below if above goes off screen
	}

	// Draw background
	vector.DrawFilledRect(screen, float32(boxX), float32(boxY), float32(boxWidth), float32(boxHeight), color.RGBA{R: 0, G: 0, B: 0, A: 200}, false)
	// Draw border
	vector.StrokeRect(screen, float32(boxX), float32(boxY), float32(boxWidth), float32(boxHeight), 2, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)

	// Draw text
	ebitenutil.DebugPrintAt(screen, text, boxX+10, boxY+8)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.width = outsideWidth
	g.height = outsideHeight
	g.calculateLayout()

	if g.vignetteImage == nil || g.vignetteImage.Bounds().Dx() != outsideWidth || g.vignetteImage.Bounds().Dy() != outsideHeight {
		// Create the vignette with 50% strength and a falloff of 1.5
		g.vignetteImage = createVignetteImage(outsideWidth, outsideHeight, 0.5, 1.5)
	}

	return outsideWidth, outsideHeight
}
