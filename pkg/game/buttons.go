package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

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
	States              map[GamePhase]*ButtonState
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
	b.States = make(map[GamePhase]*ButtonState)
	b.OnClick = onClick
}

// Render draws the button on the screen.
func (b *Button) Render(screen *ebiten.Image, gameState *GameState) {
	state := b.States[gameState.phase]

	// Check if this button should be visible
	visible := false
	if state != nil {
		visible = state.Visible
	}

	if !visible {
		return
	}

	// Use button's default color if no state is defined, otherwise use state color
	btnColor := b.Color
	if state != nil {
		btnColor = state.Color
		if state.Disabled {
			btnColor = color.RGBA{R: 100, G: 100, B: 100, A: 255}
		}
	}
	vector.DrawFilledRect(screen, float32(b.X), float32(b.Y), float32(b.Width), float32(b.Height), btnColor, false)

	if b.CustomRender != nil {
		b.CustomRender(screen, b, gameState.phase)
	} else {
		// Use button text if no state is defined, otherwise use state text
		buttonText := b.Text
		if state != nil {
			buttonText = state.Text
		}
		textWidth := len(buttonText) * 6
		textX := b.X + (b.Width-textWidth)/2
		textY := b.Y + b.Height/2 + 5
		text.Draw(screen, buttonText, b.Font, textX, textY, color.Black)
	}
}

// IsClicked checks if the button was clicked using the input state.
func (b *Button) IsClicked(input InputState, gameState *GameState) bool {
	state := b.States[gameState.phase]

	// Only check disabled state if we have a state defined
	if state != nil && (state.Disabled || !state.Visible) {
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
	g.state.catalogue = []MachineInterface{
		&Conveyor{},
		&Processor{},
		&Miner{},
		&Splitter{},
		&GeneralConsumer{},
		&Amplifier{},
		&Combiner{},
		&Booster{},
		&Catalyst{},
	}
	g.state.inventorySize = 5
	g.state.restocksLeft = 3
	g.state.inventory = dealMachines(g.state.catalogue, 5, 6)
	g.state.inventorySelected = make([]bool, len(g.state.inventory))
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
	if selected != nil && selected.IsPlaced {
		g.state.money += selected.Machine.GetCost()
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
		// Reset machines
		g.state.machines = make([]*MachineState, gridCols*gridRows)
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
	fmt.Println("Rotate Left Clicked")
	selected := g.getSelectedMachine()
	if selected != nil {
		selected.Orientation = (selected.Orientation + 3) % 4
	}
}

func handleRotateRightClick(g *Game, input InputState) {
	selected := g.getSelectedMachine()
	if selected != nil {
		selected.Orientation = (selected.Orientation + 1) % 4
	}
}

func handleRestockClick(g *Game, input InputState) {
	if g.state.restocksLeft > 0 {
		selectedIndices := []int{}
		for i, sel := range g.state.inventorySelected {
			if sel {
				selectedIndices = append(selectedIndices, i)
			}
		}
		num := len(selectedIndices)
		if num > 0 {
			// Discard selected
			newInventory := []*MachineState{}
			newSelected := []bool{}
			for i, ms := range g.state.inventory {
				if !g.state.inventorySelected[i] {
					newInventory = append(newInventory, ms)
					newSelected = append(newSelected, false)
				}
			}
			g.state.inventory = newInventory
			g.state.inventorySelected = newSelected
			// Deal num new
			newMachines := dealMachines(g.state.catalogue, num, g.state.runsLeft)
			g.state.inventory = append(g.state.inventory, newMachines...)
			g.state.inventorySelected = append(g.state.inventorySelected, make([]bool, num)...)
			g.state.restocksLeft--
		}
	}
}
