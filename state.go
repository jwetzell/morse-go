package morse

type State uint8

const (
	StateDit State = iota
	StateDah
	StateEOC
	StateEOW
)
