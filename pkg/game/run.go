package game

// SimulateRun simulates the entire run sequence.
func SimulateRun(machines []*MachineState) ([][]*Change, error) {
	history := [][]*Object{{}} // start with empty objects for tick 0
	allChanges := [][]*Change{}

	for tick := 0; tick < 1000; tick++ {
		changes := []*Change{}
		history = append(history, []*Object{})
		for pos, ms := range machines {
			if ms == nil {
				continue
			}
			chs := ms.Machine.Process(pos, history, tick, ms.Orientation)
			changes = append(changes, chs...)
		}
		if len(changes) == 0 {
			break
		}
		allChanges = append(allChanges, changes)
		for _, change := range changes {
			if change.EndObject == nil {
				continue
			}
			history[tick] = append(history[tick], change.EndObject)
		}

	}

	return allChanges, nil
}
