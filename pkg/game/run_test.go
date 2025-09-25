package game

import (
	"testing"
)

func TestSimulateRun(t *testing.T) {
	// Test with a simple setup: start machine emitting to conveyor to end
	machines := make([]*MachineState, 49) // 7x7 grid
	machines[0] = &MachineState{Machine: &Start{}, Orientation: OrientationEast, IsPlaced: true, RunAdded: 0}
	machines[1] = &MachineState{Machine: &Conveyor{}, Orientation: OrientationEast, IsPlaced: true, RunAdded: 0}
	machines[2] = &MachineState{Machine: &End{}, Orientation: OrientationEast, IsPlaced: true, RunAdded: 0}

	changes, err := SimulateRun(machines, 1)
	if err != nil {
		t.Fatalf("SimulateRun failed: %v", err)
	}

	if len(changes) == 0 {
		t.Error("Expected some changes, got none")
	}

	// Check that start emits, conveyor moves, end consumes
	// More detailed checks can be added
}

func TestSimulateRunNoMachines(t *testing.T) {
	machines := make([]*MachineState, 49)

	changes, err := SimulateRun(machines, 1)
	if err != nil {
		t.Fatalf("SimulateRun failed: %v", err)
	}

	if len(changes) != 0 {
		t.Errorf("Expected no changes, got %d ticks", len(changes))
	}
}
