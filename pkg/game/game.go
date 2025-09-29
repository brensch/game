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
	MachineGeneralConsumer
	MachineSplitter
)

// MachineRole represents the roles a machine can have.
type MachineRole int

const (
	RoleProducer MachineRole = iota
	RoleConsumer
	RoleMover
	RoleUpgrader
)

// getMachineRoleName returns the name of a machine role.
func getMachineRoleName(role MachineRole) string {
	switch role {
	case RoleProducer:
		return "Producer"
	case RoleConsumer:
		return "Consumer"
	case RoleMover:
		return "Mover"
	case RoleUpgrader:
		return "Upgrader"
	default:
		return "Unknown"
	}
}

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

// Game implements ebiten.Game.
type Game struct {
	state                                                                       *GameState
	width, height                                                               int
	topPanelHeight, infoBarHeight, foremanHeight, availableHeight, bottomHeight int
	topPanelY, foremanY, gridStartY, availableY, bottomY                        int
	screenWidth, gridStartX                                                     int
	cellSize, gridMargin                                                        int
	// lastSelected                                                                *MachineState

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

	state.catalogue = []MachineInterface{
		&Conveyor{},
		&Processor{},
		&Miner{},
		&Splitter{},
		&GeneralConsumer{},
	}
	state.inventorySize = 5
	state.restocksLeft = 3
	state.inventory = dealMachines(state.catalogue, 5, 6)
	state.inventorySelected = make([]bool, len(state.inventory))

	return g
}

func (g *Game) initButtons() {
	// Restart button
	restartBtn := &Button{}
	infoBarY := g.height - g.infoBarHeight
	restartBtn.Init(10, infoBarY+5, 80, 30, "Restart", handleRestartClick)
	restartBtn.Color = color.RGBA{R: 200, G: 100, B: 100, A: 255} // Red
	restartBtn.States[PhaseBuild] = &ButtonState{Text: "Restart", Color: color.RGBA{R: 200, G: 100, B: 100, A: 255}, Disabled: false, Visible: true}
	restartBtn.States[PhaseRun] = &ButtonState{Text: "Restart", Color: color.RGBA{R: 200, G: 100, B: 100, A: 255}, Disabled: false, Visible: true}
	restartBtn.States[PhaseRoundEnd] = &ButtonState{Text: "Restart", Color: color.RGBA{R: 200, G: 100, B: 100, A: 255}, Disabled: false, Visible: true}
	restartBtn.Font = g.font
	g.state.buttons["restart"] = restartBtn

	// Start Run button
	runBtn := &Button{}
	runBtn.Init(g.screenWidth-10-buttonWidth, g.bottomY+10, buttonWidth, g.bottomHeight-20, "Start Run", handleRunClick)
	runBtn.Color = color.RGBA{R: 100, G: 200, B: 100, A: 255} // Green
	runBtn.States[PhaseBuild] = &ButtonState{Text: "Start Run", Color: color.RGBA{R: 100, G: 200, B: 100, A: 255}, Disabled: false, Visible: true}
	runBtn.States[PhaseRun] = &ButtonState{Text: "Running", Color: color.RGBA{R: 200, G: 200, B: 100, A: 255}, Disabled: true, Visible: true}
	runBtn.Font = g.font
	g.state.buttons["run"] = runBtn

	// Rotate counterclockwise button
	rotateLeftBtn := &Button{}
	buttonSize := g.cellSize
	gridRightEdge := g.gridStartX + displayCols*g.cellSize + (displayCols-1)*g.gridMargin
	gap := 30
	counterclockwiseX := gridRightEdge - 2*buttonSize - gap
	counterclockwiseY := g.availableY + g.cellSize + g.gridMargin + 10
	rotateLeftBtn.Init(counterclockwiseX, counterclockwiseY, buttonSize, buttonSize, "<-", handleRotateLeftClick)
	rotateLeftBtn.Color = color.RGBA{R: 200, G: 100, B: 100, A: 255} // Red
	rotateLeftBtn.States[PhaseBuild] = &ButtonState{Text: "<-", Color: color.RGBA{R: 200, G: 100, B: 100, A: 255}, Disabled: false, Visible: false}
	rotateLeftBtn.Font = g.font
	rotateLeftBtn.CustomRender = func(screen *ebiten.Image, b *Button, phase GamePhase) {
		g.drawRotateArrow(screen, b.X, b.Y, b.Width, b.Height, true)
	}
	g.state.buttons["rotate_left"] = rotateLeftBtn

	// Rotate clockwise button
	rotateRightBtn := &Button{}
	clockwiseX := gridRightEdge - buttonSize
	clockwiseY := g.availableY + g.cellSize + g.gridMargin + 10
	rotateRightBtn.Init(clockwiseX, clockwiseY, buttonSize, buttonSize, "->", handleRotateRightClick)
	rotateRightBtn.Color = color.RGBA{R: 100, G: 100, B: 200, A: 255} // Blue
	rotateRightBtn.States[PhaseBuild] = &ButtonState{Text: "->", Color: color.RGBA{R: 100, G: 100, B: 200, A: 255}, Disabled: false, Visible: false}
	rotateRightBtn.Font = g.font
	rotateRightBtn.CustomRender = func(screen *ebiten.Image, b *Button, phase GamePhase) {
		g.drawRotateArrow(screen, b.X, b.Y, b.Width, b.Height, false)
	}
	g.state.buttons["rotate_right"] = rotateRightBtn

	// Next Round button
	nextRoundBtn := &Button{}
	nextRoundBtn.Init(g.screenWidth/2-50, g.height/2+50, 100, 30, "Next Round", handleNextRoundClick)
	nextRoundBtn.Color = color.RGBA{R: 100, G: 200, B: 100, A: 255} // Green
	nextRoundBtn.States[PhaseRoundEnd] = &ButtonState{Text: "Next Round", Color: color.RGBA{R: 100, G: 200, B: 100, A: 255}, Disabled: false, Visible: true}
	nextRoundBtn.Font = g.font
	g.state.buttons["next_round"] = nextRoundBtn

	// Info button
	infoBtn := &Button{}
	infoBtn.Init(10, infoBarY+45, 80, 30, "Info", handleInfoClick)
	infoBtn.Color = color.RGBA{R: 100, G: 100, B: 200, A: 255} // Blue
	infoBtn.States[PhaseBuild] = &ButtonState{Text: "Info", Color: color.RGBA{R: 100, G: 100, B: 200, A: 255}, Disabled: false, Visible: true}
	infoBtn.States[PhaseRoundEnd] = &ButtonState{Text: "Info", Color: color.RGBA{R: 100, G: 100, B: 200, A: 255}, Disabled: false, Visible: true}
	infoBtn.Font = g.font
	g.state.buttons["info"] = infoBtn // Close info button
	closeInfoBtn := &Button{}
	closeInfoBtn.Init(g.screenWidth/2-50, g.height/2+50, 100, 30, "Close", handleCloseInfoClick)
	closeInfoBtn.Color = color.RGBA{R: 100, G: 200, B: 100, A: 255} // Green
	closeInfoBtn.States[PhaseInfo] = &ButtonState{Text: "Close", Color: color.RGBA{R: 100, G: 200, B: 100, A: 255}, Disabled: false, Visible: true}
	closeInfoBtn.Font = g.font
	g.state.buttons["close_info"] = closeInfoBtn

	// Restock button
	restockBtn := &Button{}
	restockX := gridRightEdge - 2*g.cellSize - g.gridMargin - 80 - g.gridMargin - 80 - g.gridMargin
	restockBtn.Init(restockX, g.availableY+g.cellSize+g.gridMargin, 80, 30, "Restock", handleRestockClick)
	restockBtn.Color = color.RGBA{R: 200, G: 100, B: 200, A: 255} // Purple
	restockBtn.States[PhaseBuild] = &ButtonState{Text: "Restock", Color: color.RGBA{R: 200, G: 100, B: 200, A: 255}, Disabled: false, Visible: false}
	restockBtn.Font = g.font
	g.state.buttons["restock"] = restockBtn

	// Sell button
	sellBtn := &Button{}
	sellWidth := 2*buttonSize + gap
	sellX := gridRightEdge - sellWidth
	sellY := counterclockwiseY + buttonSize + gap
	sellBtn.Init(sellX, sellY, sellWidth, buttonSize, "Sell", handleSellClick)
	sellBtn.Color = color.RGBA{R: 200, G: 100, B: 100, A: 255} // Red
	sellBtn.States[PhaseBuild] = &ButtonState{Text: "Sell", Color: color.RGBA{R: 200, G: 100, B: 100, A: 255}, Disabled: false, Visible: false}
	sellBtn.Font = g.font
	g.state.buttons["sell"] = sellBtn
	popupRestartBtn := &Button{}
	popupRestartBtn.Init(g.screenWidth/2-50, g.height/2+50, 100, 30, "Restart", handleRestartClick)
	popupRestartBtn.Color = color.RGBA{R: 200, G: 100, B: 100, A: 255} // Red
	popupRestartBtn.States[PhaseGameOver] = &ButtonState{Text: "Restart", Color: color.RGBA{R: 200, G: 100, B: 100, A: 255}, Disabled: false, Visible: true}
	popupRestartBtn.Font = g.font
	g.state.buttons["popup_restart"] = popupRestartBtn
}

func (g *Game) repositionButtons() {
	infoBarY := g.height - g.infoBarHeight

	// Restart button
	if restartBtn, exists := g.state.buttons["restart"]; exists {
		restartBtn.X = 10
		restartBtn.Y = infoBarY + 5
	}

	// Start Run button
	if runBtn, exists := g.state.buttons["run"]; exists {
		runBtn.X = g.screenWidth - 10 - buttonWidth
		runBtn.Y = g.bottomY + 10
		runBtn.Width = buttonWidth
		runBtn.Height = g.bottomHeight - 20
	}

	// Rotate buttons are repositioned dynamically in phase_drag.go, skip here

	// Next Round button
	if nextRoundBtn, exists := g.state.buttons["next_round"]; exists {
		nextRoundBtn.X = g.screenWidth/2 - 50
		nextRoundBtn.Y = g.height/2 + 50
	}

	// Info button
	if infoBtn, exists := g.state.buttons["info"]; exists {
		infoBtn.X = 10
		infoBtn.Y = infoBarY + 45
	}

	// Close info button
	if closeInfoBtn, exists := g.state.buttons["close_info"]; exists {
		closeInfoBtn.X = g.screenWidth/2 - 50
		closeInfoBtn.Y = g.height/2 + 50
	}

	// Restock button
	if restockBtn, exists := g.state.buttons["restock"]; exists {
		gridRightEdge := g.gridStartX + displayCols*g.cellSize + (displayCols-1)*g.gridMargin
		restockX := gridRightEdge - 2*g.cellSize - g.gridMargin - 80 - g.gridMargin - 80 - g.gridMargin
		restockBtn.X = restockX
		restockBtn.Y = g.availableY + g.cellSize + g.gridMargin
	}

	// Sell button is repositioned dynamically in phase_drag.go, skip here

	// Popup restart button
	if popupRestartBtn, exists := g.state.buttons["popup_restart"]; exists {
		popupRestartBtn.X = g.screenWidth/2 - 50
		popupRestartBtn.Y = g.height/2 + 50
	}
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

// // updateButtonPositions updates dynamic button positions based on selected machines.
// func (g *Game) updateButtonPositions() {
// 	selected := getSelectedMachine(g.state)
// 	if selected == nil || !selected.IsPlaced {
// 		return
// 	}

// 	// Find the position of the selected machine
// 	for pos, ms := range g.state.machines {
// 		if ms == selected {
// 			col := pos % gridCols
// 			row := pos / gridCols
// 			if row >= 1 && row <= displayRows && col >= 1 && col <= displayCols {
// 				// Calculate screen position of the selected machine
// 				machineX := g.gridStartX + (col-1)*(g.cellSize+g.gridMargin)
// 				machineY := g.gridStartY + (row-1)*(g.cellSize+g.gridMargin)

// 				// Position buttons below the selected machine, offset from grid alignment
// 				buttonSize := g.cellSize / 2
// 				buttonY := machineY + g.cellSize + g.gridMargin + 10 // Extra offset from grid

// 				// Center the buttons below the machine
// 				machineCenterX := machineX + g.cellSize/2
// 				totalButtonWidth := 3*buttonSize + 2*5 // 3 buttons + 2 gaps of 5px
// 				startX := machineCenterX - totalButtonWidth/2

// 				// Update rotate left button
// 				if rotateLeft, exists := g.state.buttons["rotate_left"]; exists {
// 					rotateLeft.X = startX
// 					rotateLeft.Y = buttonY
// 					rotateLeft.Width = buttonSize
// 					rotateLeft.Height = buttonSize
// 				}

// 				// Update rotate right button
// 				if rotateRight, exists := g.state.buttons["rotate_right"]; exists {
// 					rotateRight.X = startX + buttonSize + 5
// 					rotateRight.Y = buttonY
// 					rotateRight.Width = buttonSize
// 					rotateRight.Height = buttonSize
// 				}

// 				// Update sell button
// 				if sellBtn, exists := g.state.buttons["sell"]; exists {
// 					sellBtn.X = startX + 2*buttonSize + 2*5
// 					sellBtn.Y = buttonY
// 					sellBtn.Width = buttonSize
// 					sellBtn.Height = buttonSize
// 				}
// 			}
// 			break
// 		}
// 	}
// }

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

	// // Update button positions based on current state
	// g.updateButtonPositions()

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

	// Draw tooltip if hovering over inventory
	g.drawTooltip(screen)

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
	} else {
		g.repositionButtons()
	}

	if g.vignetteImage == nil || g.vignetteImage.Bounds().Dx() != outsideWidth || g.vignetteImage.Bounds().Dy() != outsideHeight {
		// Create the vignette with 50% strength and a falloff of 1.5
		g.vignetteImage = createVignetteImage(outsideWidth, outsideHeight, 0.5, 1.5)
	}

	return outsideWidth, outsideHeight
}
