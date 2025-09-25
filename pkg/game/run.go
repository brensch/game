package game

// SimulateRun simulates the entire run sequence.
func SimulateRun(machines []*MachineState) ([][]*Change, error) {
	history := [][]*Object{{}}
	allChanges := [][]*Change{}

	for tick := 0; tick < 1000; tick++ {
		var changes []*Change
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
		history = append(history, []*Object{})
		allChanges = append(allChanges, changes)
		for _, change := range changes {
			if change.EndObject == nil {
				continue
			}
			history[tick+1] = append(history[tick+1], change.EndObject)
		}

	}

	return allChanges, nil
}
