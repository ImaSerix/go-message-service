package domain

type Direction string

const (
	DirectionForward  Direction = "forward"  // newer than cursor
	DirectionBackward Direction = "backward" // older than cursor
)
