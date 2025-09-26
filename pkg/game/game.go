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
	buttonWidth     = 100

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
	animations        []*Animation
	animationTick     int
	animationSpeed    float64
	buttons           map[string]*Button
	inputPressed      bool
	pressX, pressY    int
}

// Button represents a clickable UI button.
type Button struct {
	X, Y, Width, Height int
	Text                string
	Disabled            bool
	Color               color.RGBA
}

// Init initializes the button with dimensions.
func (b *Button) Init(x, y, width, height int, text string) {
	b.X = x
	b.Y = y
	b.Width = width
	b.Height = height
	b.Text = text
	b.Disabled = false
	b.Color = color.RGBA{R: 100, G: 200, B: 100, A: 255} // Default green
}

// Render draws the button on the screen.
func (b *Button) Render(screen *ebiten.Image) {
	btnColor := b.Color
	if b.Disabled {
		btnColor = color.RGBA{R: 100, G: 100, B: 100, A: 255}
	}
	vector.DrawFilledRect(screen, float32(b.X), float32(b.Y), float32(b.Width), float32(b.Height), btnColor, false)
	ebitenutil.DebugPrintAt(screen, b.Text, b.X+5, b.Y+b.Height/2+5)
}

// IsClicked checks if the button was clicked using the input state.
func (b *Button) IsClicked(input InputState) bool {
	if b.Disabled {
		return false
	}
	if input.JustPressed {
		cx, cy := input.X, input.Y
		return cx >= b.X && cx <= b.X+b.Width && cy >= b.Y && cy <= b.Y+b.Height
	}
	return false
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
	lastInput     InputState
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
		buttons:        make(map[string]*Button),
		inputPressed:   false,
		pressX:         0,
		pressY:         0,
	}

	g := &Game{state: state}
	g.width = width
	g.height = height
	g.calculateLayout()

	// Initialize buttons
	g.initButtons()

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

func (g *Game) initButtons() {
	// Restart button
	restartBtn := &Button{}
	restartBtn.Init(g.screenWidth-100, g.topPanelY+10, 80, g.topPanelHeight-20, "Restart")
	restartBtn.Color = color.RGBA{R: 200, G: 100, B: 100, A: 255} // Red
	g.state.buttons["restart"] = restartBtn

	// Sell button
	sellBtn := &Button{}
	sellBtn.Init(10, g.bottomY+10, buttonWidth, g.bottomHeight-20, "Sell")
	sellBtn.Color = color.RGBA{R: 255, G: 100, B: 100, A: 255} // Red
	g.state.buttons["sell"] = sellBtn

	// Start Run button
	runBtn := &Button{}
	runBtn.Init(g.screenWidth-10-buttonWidth, g.bottomY+10, buttonWidth, g.bottomHeight-20, "Start Run")
	runBtn.Color = color.RGBA{R: 100, G: 200, B: 100, A: 255} // Green
	g.state.buttons["run"] = runBtn

	// Rotate counterclockwise button
	rotateLeftBtn := &Button{}
	counterclockwiseX := g.screenWidth - 2*g.cellSize - g.gridMargin
	counterclockwiseY := g.availableY
	rotateLeftBtn.Init(counterclockwiseX, counterclockwiseY, g.cellSize, g.cellSize, "<-")
	rotateLeftBtn.Color = color.RGBA{R: 200, G: 100, B: 100, A: 255} // Red
	g.state.buttons["rotate_left"] = rotateLeftBtn

	// Rotate clockwise button
	rotateRightBtn := &Button{}
	clockwiseX := g.screenWidth - g.cellSize
	clockwiseY := g.availableY
	rotateRightBtn.Init(clockwiseX, clockwiseY, g.cellSize, g.cellSize, "->")
	rotateRightBtn.Color = color.RGBA{R: 100, G: 100, B: 200, A: 255} // Blue
	g.state.buttons["rotate_right"] = rotateRightBtn
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
	g.GetInput()

	switch g.state.phase {
	case PhaseBuild:
		g.handleDragAndDrop()
	case PhaseRun:
		g.handleRunPhase()
	}

	// Check button clicks
	if g.state.buttons["run"].IsClicked(g.lastInput) {
		if g.state.phase == PhaseBuild {
			g.state.phase = PhaseRun
			g.state.animations = []*Animation{}
			g.state.animationTick = 0
			g.state.animationSpeed = 1.0
		}
	}
	if g.state.buttons["restart"].IsClicked(g.lastInput) {
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
			buttons:        make(map[string]*Button),
			inputPressed:   false,
			pressX:         0,
			pressY:         0,
		}
		g.initButtons()
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
	if g.state.buttons["sell"].IsClicked(g.lastInput) {
		selected := g.getSelectedMachine()
		if selected != nil && selected.IsPlaced && selected.Machine.GetType() != MachineEnd {
			// Remove from grid
			for pos, ms := range g.state.machines {
				if ms == selected {
					g.state.machines[pos] = nil
					break
				}
			}
			// Deselect
			selected.Selected = false
		}
	}

	// Update button states
	if g.state.phase == PhaseRun {
		g.state.buttons["run"].Text = "Running"
		g.state.buttons["run"].Color = color.RGBA{R: 200, G: 200, B: 100, A: 255} // Yellow
	} else {
		g.state.buttons["run"].Text = "Start Run"
		g.state.buttons["run"].Color = color.RGBA{R: 100, G: 200, B: 100, A: 255} // Green
	}
	return nil
}// Draw draws the game screen.
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

	// Debug input state
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Input: P:%v JP:%v JR:%v X:%d Y:%d", g.lastInput.Pressed, g.lastInput.JustPressed, g.lastInput.JustReleased, g.lastInput.X, g.lastInput.Y), 10, 10)
	runButtonX := g.screenWidth - 10 - buttonWidth
	runButtonY := g.bottomY + 10
	runButtonWidth := buttonWidth
	runButtonHeight := g.bottomHeight - 20
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("RunBtn: X:%d-%d Y:%d-%d", runButtonX-10, runButtonX+runButtonWidth+10, runButtonY-10, runButtonY+runButtonHeight+10), 10, 30)

	// Apply CRT effects
	g.drawScanlines(screen)
	// Now, draw the vignette overlay on top of everything.
	if g.vignetteImage != nil {
		op := &ebiten.DrawImageOptions{}
		screen.DrawImage(g.vignetteImage, op)
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
