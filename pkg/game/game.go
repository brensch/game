package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
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
	Score        *Score
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
	PhaseRoundEnd
	PhaseGameOver
	PhaseInfo
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
	runsLeft          int
	baseScore         int
	multiplier        int
	multMult          int
	machines          []*MachineState
	inventory         []*MachineState
	catalogue         []MachineInterface
	inventorySize     int
	restocksLeft      int
	inventorySelected []bool
	objects           []*Object
	round             int
	animations        []*Animation
	animationTick     int
	animationSpeed    float64
	buttons           map[string]*Button
	allChanges        [][]*Change
	roundScore        int
	totalScore        int
	targetScore       int
	gameOver          bool
	endRunDelay       int
	previousPhase     GamePhase
}

// ButtonState represents the state of a button for a phase.
type ButtonState struct {
	Text     string
	Color    color.RGBA
	Disabled bool
	Visible  bool
}

// Button represents a clickable UI button.
type Button struct {
	X, Y, Width, Height int
	Text                string
	Disabled            bool
	Color               color.RGBA
	Font                font.Face
	States              map[GamePhase]ButtonState
	CustomRender        func(screen *ebiten.Image, b *Button, phase GamePhase)
	OnClick             func(g *Game, input InputState) // Click handler function
}

// Init initializes the button with dimensions and click handler.
func (b *Button) Init(x, y, width, height int, text string, onClick func(g *Game, input InputState)) {
	b.X = x
	b.Y = y
	b.Width = width
	b.Height = height
	b.Text = text
	b.Disabled = false
	b.Color = color.RGBA{R: 100, G: 200, B: 100, A: 255} // Default green
	b.States = make(map[GamePhase]ButtonState)
	b.OnClick = onClick
}

// Render draws the button on the screen.
func (b *Button) Render(screen *ebiten.Image, gameState *GameState) {
	state := b.States[gameState.phase]

	// Check if this button should be visible based on the game state
	visible := state.Visible

	// Special logic for machine-dependent buttons
	if b == gameState.buttons["rotate_left"] || b == gameState.buttons["rotate_right"] || b == gameState.buttons["sell"] {
		// These buttons are visible when a placeable, non-End machine is selected
		selected := getSelectedMachine(gameState)
		visible = selected != nil && selected.IsPlaced && selected.Machine.GetType() != MachineEnd
	}

	if !visible {
		return
	}

	btnColor := state.Color
	if state.Disabled {
		btnColor = color.RGBA{R: 100, G: 100, B: 100, A: 255}
	}
	vector.DrawFilledRect(screen, float32(b.X), float32(b.Y), float32(b.Width), float32(b.Height), btnColor, false)

	if b.CustomRender != nil {
		b.CustomRender(screen, b, gameState.phase)
	} else {
		text.Draw(screen, state.Text, b.Font, b.X+5, b.Y+b.Height/2+5, color.Black)
	}
}

// IsClicked checks if the button was clicked using the input state.
func (b *Button) IsClicked(input InputState, gameState *GameState) bool {
	state := b.States[gameState.phase]

	// Check if this button should be visible based on the game state
	visible := state.Visible

	// Special logic for machine-dependent buttons
	if b == gameState.buttons["rotate_left"] || b == gameState.buttons["rotate_right"] || b == gameState.buttons["sell"] {
		// These buttons are visible when a placeable, non-End machine is selected
		selected := getSelectedMachine(gameState)
		visible = selected != nil && selected.IsPlaced && selected.Machine.GetType() != MachineEnd
	}

	if !visible || state.Disabled {
		return false
	}
	if input.JustPressed {
		cx, cy := input.X, input.Y
		return cx >= b.X && cx <= b.X+b.Width && cy >= b.Y && cy <= b.Y+b.Height
	}
	return false
}

// HandleClick processes click events for this button.
func (b *Button) HandleClick(g *Game, input InputState) {
	if b.IsClicked(input, g.state) && b.OnClick != nil {
		b.OnClick(g, input)
	}
}

// Game implements ebiten.Game.
type Game struct {
	state                                                                       *GameState
	width, height                                                               int
	topPanelHeight, infoBarHeight, foremanHeight, availableHeight, bottomHeight int
	topPanelY, foremanY, gridStartY, availableY, bottomY                        int
	screenWidth, gridStartX                                                     int
	cellSize, gridMargin                                                        int
	lastSelected                                                                *MachineState

	vignetteImage *ebiten.Image
	font          font.Face
	lastInput     InputState
}

func (g *Game) getSelectedMachine() *MachineState {
	for _, ms := range g.state.machines {
		if ms != nil && ms.Selected {
			return ms
		}
	}
	for _, ms := range g.state.inventory {
		if ms != nil && ms.Selected {
			return ms
		}
	}
	return nil
}

// getSelectedMachine is a helper function to get selected machine from GameState
func getSelectedMachine(gameState *GameState) *MachineState {
	for _, ms := range gameState.machines {
		if ms != nil && ms.Selected {
			return ms
		}
	}
	for _, ms := range gameState.inventory {
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
	for _, ms := range g.state.inventory {
		if ms != nil && ms.BeingDragged {
			return ms
		}
	}
	return nil
}

func dealMachines(catalogue []MachineInterface, n int, runsLeft int) []*MachineState {
	result := make([]*MachineState, n)
	for i := 0; i < n; i++ {
		idx := rand.Intn(len(catalogue))
		machine := catalogue[idx]
		result[i] = &MachineState{Machine: machine, Orientation: OrientationEast, BeingDragged: false, IsPlaced: false, RunAdded: runsLeft}
	}
	return result
}

// NewGame creates a new Game instance.
func NewGame(width, height int) *Game {
	state := &GameState{
		phase:          PhaseBuild,
		money:          10,
		runsLeft:       6,
		machines:       make([]*MachineState, gridCols*gridRows),
		round:          1,
		animations:     []*Animation{},
		animationTick:  0,
		animationSpeed: 1.0,
		buttons:        make(map[string]*Button),
		multiplier:     1,
		multMult:       1,
		roundScore:     0,
		totalScore:     0,
		targetScore:    10, // round * 10
		gameOver:       false,
		endRunDelay:    0,
		previousPhase:  PhaseBuild,
	}

	g := &Game{state: state}
	g.width = width
	g.height = height
	fontData := gomono.TTF
	parsed, _ := opentype.Parse(fontData)
	g.font, _ = opentype.NewFace(parsed, &opentype.FaceOptions{Size: 16, DPI: 72})
	g.calculateLayout()

	// Initialize buttons
	g.initButtons()

	// Place random End machine
	endRow := 1 + rand.Intn(displayRows)
	endCol := 1 + rand.Intn(displayCols)
	endPos := endRow*gridCols + endCol

	state.machines[endPos] = &MachineState{Machine: &End{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: true, RunAdded: 0, OriginalPos: endPos}

	state.catalogue = []MachineInterface{
		&Conveyor{},
		&Processor{},
		&Miner{},
	}
	state.inventorySize = 5
	state.restocksLeft = 3
	state.inventory = dealMachines(state.catalogue, 5, 6)
	state.inventorySelected = make([]bool, len(state.inventory))

	return g
}

// Button click handlers
func handleRestartClick(g *Game, input InputState) {
	// Reset game state
	g.state = &GameState{
		phase:          PhaseBuild,
		money:          10,
		runsLeft:       6,
		machines:       make([]*MachineState, gridCols*gridRows),
		round:          1,
		animations:     []*Animation{},
		animationTick:  0,
		animationSpeed: 1.0,
		buttons:        make(map[string]*Button),
		allChanges:     nil,
		multiplier:     1,
		multMult:       1,
		roundScore:     0,
		totalScore:     0,
		targetScore:    10,
		gameOver:       false,
		endRunDelay:    0,
		previousPhase:  PhaseBuild,
	}
	g.initButtons()
	// Place random End machine
	endRow := 1 + rand.Intn(displayRows)
	endCol := 1 + rand.Intn(displayCols)
	endPos := endRow*gridCols + endCol
	g.state.machines[endPos] = &MachineState{Machine: &End{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: true, RunAdded: 0, OriginalPos: endPos}
	g.state.inventory = []*MachineState{
		{Machine: &Conveyor{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: false, RunAdded: 0},
		{Machine: &Processor{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: false, RunAdded: 0},
		{Machine: &Miner{}, Orientation: OrientationEast, BeingDragged: false, IsPlaced: false, RunAdded: 0},
	}
}

func handleRunClick(g *Game, input InputState) {
	if g.state.phase == PhaseBuild {
		g.state.phase = PhaseRun
		g.state.allChanges = nil
		g.state.animations = []*Animation{}
		g.state.animationTick = 0
		g.state.animationSpeed = 1.0
		g.state.endRunDelay = 0
		go func() {
			changes, _ := SimulateRun(g.state.machines)
			g.state.allChanges = changes
		}()
	}
}

func handleSellClick(g *Game, input InputState) {
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

func handleNextRoundClick(g *Game, input InputState) {
	if g.state.phase == PhaseRoundEnd {
		// Advance to next round
		g.state.phase = PhaseBuild
		g.state.runsLeft = 6
		g.state.round++
		g.state.targetScore = g.state.round * g.state.round * 10
		g.state.money += g.state.round * 10
		// Reset machines: keep only End, clear others
		var endMachine *MachineState
		for _, ms := range g.state.machines {
			if ms != nil && ms.Machine.GetType() == MachineEnd {
				endMachine = ms
				endMachine.IsPlaced = true
				endMachine.RunAdded = g.state.runsLeft
				break
			}
		}
		g.state.machines = make([]*MachineState, gridCols*gridRows)
		if endMachine != nil {
			// Place End at a random position
			endRow := 1 + rand.Intn(displayRows)
			endCol := 1 + rand.Intn(displayCols)
			endPos := endRow*gridCols + endCol
			g.state.machines[endPos] = endMachine
		}
		// Reset available machines
		g.state.inventory = dealMachines(g.state.catalogue, g.state.inventorySize, g.state.runsLeft)
		g.state.inventorySelected = make([]bool, len(g.state.inventory))
		g.state.restocksLeft = 3
	}
}

func handleInfoClick(g *Game, input InputState) {
	if g.state.phase == PhaseBuild || g.state.phase == PhaseRoundEnd {
		g.state.previousPhase = g.state.phase
		g.state.phase = PhaseInfo
	}
}

func handleCloseInfoClick(g *Game, input InputState) {
	g.state.phase = g.state.previousPhase
}

func handleRotateLeftClick(g *Game, input InputState) {
	// TODO: Implement rotate left functionality
}

func handleRotateRightClick(g *Game, input InputState) {
	// TODO: Implement rotate right functionality
}

func handleRestockClick(g *Game, input InputState) {
	// TODO: Implement restock functionality
}

func (g *Game) initButtons() {
	// Restart button
	restartBtn := &Button{}
	infoBarY := g.height - g.infoBarHeight
	restartBtn.Init(10, infoBarY+5, 80, 30, "Restart", handleRestartClick)
	restartBtn.Color = color.RGBA{R: 200, G: 100, B: 100, A: 255} // Red
	restartBtn.States[PhaseBuild] = ButtonState{Text: "Restart", Color: color.RGBA{R: 200, G: 100, B: 100, A: 255}, Disabled: false, Visible: true}
	restartBtn.States[PhaseRun] = ButtonState{Text: "Restart", Color: color.RGBA{R: 200, G: 100, B: 100, A: 255}, Disabled: false, Visible: true}
	restartBtn.States[PhaseRoundEnd] = ButtonState{Text: "Restart", Color: color.RGBA{R: 200, G: 100, B: 100, A: 255}, Disabled: false, Visible: true}
	restartBtn.Font = g.font
	g.state.buttons["restart"] = restartBtn

	// Start Run button
	runBtn := &Button{}
	runBtn.Init(g.screenWidth-10-buttonWidth, g.bottomY+10, buttonWidth, g.bottomHeight-20, "Start Run", handleRunClick)
	runBtn.Color = color.RGBA{R: 100, G: 200, B: 100, A: 255} // Green
	runBtn.States[PhaseBuild] = ButtonState{Text: "Start Run", Color: color.RGBA{R: 100, G: 200, B: 100, A: 255}, Disabled: false, Visible: true}
	runBtn.States[PhaseRun] = ButtonState{Text: "Running", Color: color.RGBA{R: 200, G: 200, B: 100, A: 255}, Disabled: true, Visible: true}
	runBtn.Font = g.font
	g.state.buttons["run"] = runBtn

	// Rotate counterclockwise button
	rotateLeftBtn := &Button{}
	gridRightEdge := g.gridStartX + displayCols*g.cellSize + (displayCols-1)*g.gridMargin
	counterclockwiseX := gridRightEdge - 2*g.cellSize - g.gridMargin
	counterclockwiseY := g.availableY + g.cellSize + g.gridMargin
	rotateLeftBtn.Init(counterclockwiseX, counterclockwiseY, g.cellSize, g.cellSize, "<-", handleRotateLeftClick)
	rotateLeftBtn.Color = color.RGBA{R: 200, G: 100, B: 100, A: 255} // Red
	rotateLeftBtn.States[PhaseBuild] = ButtonState{Text: "<-", Color: color.RGBA{R: 200, G: 100, B: 100, A: 255}, Disabled: false, Visible: false}
	rotateLeftBtn.States[PhaseRun] = ButtonState{Text: "<-", Color: color.RGBA{R: 200, G: 100, B: 100, A: 255}, Disabled: false, Visible: false}
	rotateLeftBtn.Font = g.font
	rotateLeftBtn.CustomRender = func(screen *ebiten.Image, b *Button, phase GamePhase) {
		g.drawRotateArrow(screen, b.X, b.Y, b.Width, b.Height, true)
	}
	g.state.buttons["rotate_left"] = rotateLeftBtn

	// Rotate clockwise button
	rotateRightBtn := &Button{}
	clockwiseX := gridRightEdge - g.cellSize
	clockwiseY := g.availableY + g.cellSize + g.gridMargin
	rotateRightBtn.Init(clockwiseX, clockwiseY, g.cellSize, g.cellSize, "->", handleRotateRightClick)
	rotateRightBtn.Color = color.RGBA{R: 100, G: 100, B: 200, A: 255} // Blue
	rotateRightBtn.States[PhaseBuild] = ButtonState{Text: "->", Color: color.RGBA{R: 100, G: 100, B: 200, A: 255}, Disabled: false, Visible: false}
	rotateRightBtn.States[PhaseRun] = ButtonState{Text: "->", Color: color.RGBA{R: 100, G: 100, B: 200, A: 255}, Disabled: false, Visible: false}
	rotateRightBtn.Font = g.font
	rotateRightBtn.CustomRender = func(screen *ebiten.Image, b *Button, phase GamePhase) {
		g.drawRotateArrow(screen, b.X, b.Y, b.Width, b.Height, false)
	}
	g.state.buttons["rotate_right"] = rotateRightBtn

	// Next Round button
	nextRoundBtn := &Button{}
	nextRoundBtn.Init(g.screenWidth/2-50, g.height/2+50, 100, 30, "Next Round", handleNextRoundClick)
	nextRoundBtn.Color = color.RGBA{R: 100, G: 200, B: 100, A: 255} // Green
	nextRoundBtn.States[PhaseRoundEnd] = ButtonState{Text: "Next Round", Color: color.RGBA{R: 100, G: 200, B: 100, A: 255}, Disabled: false, Visible: true}
	nextRoundBtn.Font = g.font
	g.state.buttons["next_round"] = nextRoundBtn

	// Info button
	infoBtn := &Button{}
	infoBtn.Init(10, infoBarY+45, 80, 30, "Info", handleInfoClick)
	infoBtn.Color = color.RGBA{R: 100, G: 100, B: 200, A: 255} // Blue
	infoBtn.States[PhaseBuild] = ButtonState{Text: "Info", Color: color.RGBA{R: 100, G: 100, B: 200, A: 255}, Disabled: false, Visible: true}
	infoBtn.States[PhaseRoundEnd] = ButtonState{Text: "Info", Color: color.RGBA{R: 100, G: 100, B: 200, A: 255}, Disabled: false, Visible: true}
	infoBtn.Font = g.font
	g.state.buttons["info"] = infoBtn // Close info button
	closeInfoBtn := &Button{}
	closeInfoBtn.Init(g.screenWidth/2-50, g.height/2+50, 100, 30, "Close", handleCloseInfoClick)
	closeInfoBtn.Color = color.RGBA{R: 100, G: 200, B: 100, A: 255} // Green
	closeInfoBtn.States[PhaseInfo] = ButtonState{Text: "Close", Color: color.RGBA{R: 100, G: 200, B: 100, A: 255}, Disabled: false, Visible: true}
	closeInfoBtn.Font = g.font
	g.state.buttons["close_info"] = closeInfoBtn

	// Restock button
	restockBtn := &Button{}
	restockX := gridRightEdge - 2*g.cellSize - g.gridMargin - 80 - g.gridMargin - 80 - g.gridMargin
	restockBtn.Init(restockX, g.availableY+g.cellSize+g.gridMargin, 80, 30, "Restock", handleRestockClick)
	restockBtn.Color = color.RGBA{R: 200, G: 100, B: 200, A: 255} // Purple
	restockBtn.States[PhaseBuild] = ButtonState{Text: "Restock", Color: color.RGBA{R: 200, G: 100, B: 200, A: 255}, Disabled: false, Visible: false}
	restockBtn.Font = g.font
	g.state.buttons["restock"] = restockBtn

	// Sell button
	sellBtn := &Button{}
	sellX := gridRightEdge - 2*g.cellSize - g.gridMargin - 80 - g.gridMargin
	sellBtn.Init(sellX, g.availableY+g.cellSize+g.gridMargin, 80, 30, "Sell", handleSellClick)
	sellBtn.Color = color.RGBA{R: 200, G: 100, B: 100, A: 255} // Red
	sellBtn.States[PhaseBuild] = ButtonState{Text: "Sell", Color: color.RGBA{R: 200, G: 100, B: 100, A: 255}, Disabled: false, Visible: false}
	sellBtn.Font = g.font
	g.state.buttons["sell"] = sellBtn
	popupRestartBtn := &Button{}
	popupRestartBtn.Init(g.screenWidth/2-50, g.height/2+50, 100, 30, "Restart", handleRestartClick)
	popupRestartBtn.Color = color.RGBA{R: 200, G: 100, B: 100, A: 255} // Red
	popupRestartBtn.States[PhaseGameOver] = ButtonState{Text: "Restart", Color: color.RGBA{R: 200, G: 100, B: 100, A: 255}, Disabled: false, Visible: true}
	popupRestartBtn.Font = g.font
	g.state.buttons["popup_restart"] = popupRestartBtn
}

func (g *Game) calculateLayout() {
	g.topPanelHeight = topPanelHeight
	g.infoBarHeight = topPanelHeight
	g.foremanHeight = foremanHeight
	g.availableHeight = availableHeight
	g.bottomHeight = bottomHeight

	marginRatio := 1.0 / 6.0
	availableWidth := g.screenWidth - 40
	availableHeight := g.height - (g.foremanHeight + g.availableHeight + g.bottomHeight + g.infoBarHeight + 4*minGap)
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
	totalFixedHeight := gridHeight + g.availableHeight + g.bottomHeight + g.infoBarHeight
	gap := (g.height - totalFixedHeight) / 4
	if gap < minGap {
		gap = minGap
	}
	g.topPanelY = 0
	g.gridStartY = g.topPanelY + gap
	g.availableY = g.gridStartY + gridHeight + gap
	g.bottomY = g.height - g.bottomHeight - g.infoBarHeight
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

// updateButtonPositions updates dynamic button positions based on selected machines.
func (g *Game) updateButtonPositions() {
	selected := getSelectedMachine(g.state)
	if selected == nil || !selected.IsPlaced || selected.Machine.GetType() == MachineEnd {
		return
	}

	// Find the position of the selected machine
	for pos, ms := range g.state.machines {
		if ms == selected {
			col := pos % gridCols
			row := pos / gridCols
			if row >= 1 && row <= displayRows && col >= 1 && col <= displayCols {
				// Calculate screen position of the selected machine
				machineX := g.gridStartX + (col-1)*(g.cellSize+g.gridMargin)
				machineY := g.gridStartY + (row-1)*(g.cellSize+g.gridMargin)

				// Position buttons below the selected machine
				buttonY := machineY + g.cellSize + g.gridMargin

				// Update rotate left button
				if rotateLeft, exists := g.state.buttons["rotate_left"]; exists {
					rotateLeft.X = machineX
					rotateLeft.Y = buttonY
				}

				// Update rotate right button
				if rotateRight, exists := g.state.buttons["rotate_right"]; exists {
					rotateRight.X = machineX + g.cellSize + g.gridMargin
					rotateRight.Y = buttonY
				}

				// Update sell button
				if sellBtn, exists := g.state.buttons["sell"]; exists {
					sellBtn.X = machineX + 2*(g.cellSize+g.gridMargin)
					sellBtn.Y = buttonY
				}
			}
			break
		}
	}
}

// processButtons processes all button clicks using their individual handlers.
func (g *Game) processButtons() {
	for _, button := range g.state.buttons {
		button.HandleClick(g, g.lastInput)
	}
}

// Update proceeds the game state.
func (g *Game) Update() error {
	g.GetInput()

	switch g.state.phase {
	case PhaseBuild:
		g.handleDragAndDrop()
	case PhaseRun:
		g.handleRunPhase()
	case PhaseRoundEnd:
		// Handle round end phase
	}

	// Update button positions based on current state
	g.updateButtonPositions()

	// Process all button clicks
	g.processButtons()

	return nil
} // Draw draws the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 40, G: 40, B: 40, A: 255})

	switch g.state.phase {
	case PhaseBuild:
		g.drawDragLayout(screen)
	case PhaseRun:
		g.drawRunLayout(screen)
	case PhaseRoundEnd:
		g.drawRoundEndLayout(screen)
	case PhaseGameOver:
		// Draw a simple game over screen
		g.drawRoundEndLayout(screen) // or something
	case PhaseInfo:
		// Draw the layout of the previous phase
		switch g.state.previousPhase {
		case PhaseBuild:
			g.drawDragLayout(screen)
		case PhaseRoundEnd:
			g.drawRoundEndLayout(screen)
		}
	}

	// Apply CRT effects
	g.drawScanlines(screen)
	// Now, draw the vignette overlay on top of everything.
	if g.vignetteImage != nil {
		op := &ebiten.DrawImageOptions{}
		screen.DrawImage(g.vignetteImage, op)
	}

	// Render all buttons
	for _, button := range g.state.buttons {
		button.Render(screen, g.state)
	}
	// Draw game over popup
	if g.state.phase == PhaseGameOver {
		popupX := g.screenWidth/2 - 150
		popupY := g.height/2 - 100
		popupW := 300
		popupH := 200
		vector.DrawFilledRect(screen, float32(popupX), float32(popupY), float32(popupW), float32(popupH), color.RGBA{R: 50, G: 50, B: 50, A: 200}, false)
		vector.DrawFilledRect(screen, float32(popupX), float32(popupY), float32(popupW), float32(popupH), color.RGBA{R: 0, G: 0, B: 0, A: 0}, true) // Border
		text.Draw(screen, "Game Over", g.font, popupX+20, popupY+30, color.White)
		text.Draw(screen, fmt.Sprintf("Final Score: %d", g.state.totalScore), g.font, popupX+20, popupY+60, color.White)
		g.state.buttons["popup_restart"].Render(screen, g.state)
	}

	// Draw info popup
	if g.state.phase == PhaseInfo {
		popupX := g.screenWidth/2 - 150
		popupY := g.height/2 - 100
		popupW := 300
		popupH := 200
		vector.DrawFilledRect(screen, float32(popupX), float32(popupY), float32(popupW), float32(popupH), color.RGBA{R: 50, G: 50, B: 50, A: 200}, false)
		vector.DrawFilledRect(screen, float32(popupX), float32(popupY), float32(popupW), float32(popupH), color.RGBA{R: 0, G: 0, B: 0, A: 0}, true) // Border
		text.Draw(screen, "Game Info", g.font, popupX+20, popupY+30, color.White)
		text.Draw(screen, "This is a factory automation game.", g.font, popupX+20, popupY+60, color.White)
		text.Draw(screen, "Build machines to process objects.", g.font, popupX+20, popupY+80, color.White)
		g.state.buttons["close_info"].Render(screen, g.state)
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

	// Only initialize buttons if they don't exist yet
	if len(g.state.buttons) == 0 {
		g.initButtons()
	}

	if g.vignetteImage == nil || g.vignetteImage.Bounds().Dx() != outsideWidth || g.vignetteImage.Bounds().Dy() != outsideHeight {
		// Create the vignette with 50% strength and a falloff of 1.5
		g.vignetteImage = createVignetteImage(outsideWidth, outsideHeight, 0.5, 1.5)
	}

	return outsideWidth, outsideHeight
}
