package morse

import "fmt"

type Code interface {
	Encode(rune) (string, bool)
	Decode(string) (rune, bool)
	EncodeStates(msg string) ([]State, error)
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

func (c *mapCode) EncodeStates(msg string) ([]State, error) {
	var states []State
	for _, r := range msg {
		switch r {
		case ' ':
			states = append(states, StateEOW)
			continue
		default:
			code, ok := c.Encode(r)
			if !ok {
				return nil, fmt.Errorf("failed to encode character: %c", r)
			}
			for _, ch := range code {
				switch ch {
				case '.':
					states = append(states, StateDit)
				case '-':
					states = append(states, StateDah)
				default:
					return nil, fmt.Errorf("unexpected character in code: %c", ch)
				}
			}
		}
		states = append(states, StateEOC)
	}
	if len(states) > 0 {
		states[len(states)-1] = StateEOW
	}
	return states, nil
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
