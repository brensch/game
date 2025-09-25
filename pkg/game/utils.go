package game

// GetAdjacentPosition returns the grid position adjacent to the given position in the specified orientation.
// Assumes gridCols is the number of columns in the grid.
func GetAdjacentPosition(pos int, orientation Orientation) int {
	row := pos / gridCols
	col := pos % gridCols

	switch orientation {
	case OrientationNorth:
		row--
	case OrientationSouth:
		row++
	case OrientationEast:
		col++
	case OrientationWest:
		col--
	}

	// Return the new position, even if out of bounds
	return row*gridCols + col
}
