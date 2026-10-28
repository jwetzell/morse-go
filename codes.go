package morse

import "fmt"

type Code interface {
	Encode(rune) (string, bool)
	Decode(string) (rune, bool)
	DecodeStates([]State) (string, error)
}

type mapCode struct {
	encodeMap map[rune]string
}

func (c *mapCode) Encode(r rune) (string, bool) {
	code, ok := c.encodeMap[r]
	return code, ok
}

func (c *mapCode) Decode(code string) (rune, bool) {
	for r, c := range c.encodeMap {
		if c == code {
			return r, true
		}
	}
	return 0, false
}

func (c *mapCode) DecodeStates(states []State) (string, error) {
	var result string

	currentCharacter := ""
	for i, s := range states {
		switch s {
		case StateDit:
			currentCharacter += "."
		case StateDah:
			currentCharacter += "-"
		case StateEOC, StateEOW:
			if currentCharacter != "" {
				r, ok := c.Decode(currentCharacter)
				if !ok {
					return "", fmt.Errorf("failed to decode character: %s", currentCharacter)
				}
				result += string(r)
				currentCharacter = ""
			}
		default:
			return "", fmt.Errorf("unexpected state: %v", s)
		}
		if s == StateEOW && i < len(states)-1 {
			result += " "
		}
	}
	if currentCharacter != "" {
		r, ok := c.Decode(currentCharacter)
		if !ok {
			return "", fmt.Errorf("failed to decode character: %s", currentCharacter)
		}
		result += string(r)
		currentCharacter = ""
	}
	return result, nil
}

func newCodeFromMap(encodeMap map[rune]string) Code {
	return &mapCode{encodeMap: encodeMap}
}
