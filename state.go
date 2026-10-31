package morse

type State uint8

const (
	StateDit State = iota
	StateDah
	StateEOC
	StateEOW
)

func (s State) String() string {
	switch s {
	case StateDit:
		return "."
	case StateDah:
		return "-"
	case StateEOC:
		return ""
	case StateEOW:
		return " "
	default:
		return ""
	}
}
