package game

// SimulateRun simulates the entire run sequence.
func SimulateRun(machines []*MachineState, initialObjects []*Object, round int) ([][]*Change, error) {
	history := [][]*Object{initialObjects}
	allChanges := [][]*Change{}

	for tick := 0; tick < 1000; tick++ {
		changes := []*Change{}
		for pos, ms := range machines {
			if ms != nil && ms.IsPlaced {
				chs := ms.Machine.Process(pos, history, round, ms.Orientation)
				changes = append(changes, chs...)
			}
		}
		if len(changes) == 0 {
			break
		}
		allChanges = append(allChanges, changes)

		// Apply changes to get new objects
		newObjects := make([]*Object, len(history[len(history)-1]))
		copy(newObjects, history[len(history)-1])
		applyChanges(changes, &newObjects)
		history = append(history, newObjects)
	}

	return allChanges, nil
}
