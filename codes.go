package morse

import "fmt"

type Code interface {
	EncodeRune(rune) ([]State, bool)
	DecodeRune([]State) (rune, bool)
	EncodeString(msg string) ([]State, error)
	DecodeString([]State) (string, error)
}

type mapCode struct {
	encodeMap map[rune][]State
}

func (c *mapCode) EncodeRune(r rune) ([]State, bool) {
	code, ok := c.encodeMap[r]
	if !ok {
		return nil, false
	}
	return code, true
}

func (c *mapCode) DecodeRune(code []State) (rune, bool) {
	for r, c := range c.encodeMap {
		if len(c) != len(code) {
			continue
		}
		match := true
		for i := range c {
			if c[i] != code[i] {
				match = false
				break
			}
		}
		if match {
			return r, true
		}
	}
	return 0, false
}

func (c *mapCode) EncodeString(msg string) ([]State, error) {
	var states []State
	for _, r := range msg {
		switch r {
		case ' ':
			states = append(states, StateEOW)
			continue
		default:
			code, ok := c.EncodeRune(r)
			if !ok {
				return nil, fmt.Errorf("failed to encode character: %c", r)
			}
			states = append(states, code...)
		}
		states = append(states, StateEOC)
	}
	if len(states) > 0 {
		states[len(states)-1] = StateEOW
	}
	return states, nil
}

func (c *mapCode) DecodeString(states []State) (string, error) {
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
				var code []State
				for _, ch := range currentCharacter {
					switch ch {
					case '.':
						code = append(code, StateDit)
					case '-':
						code = append(code, StateDah)
					default:
						return "", fmt.Errorf("unexpected character in code: %c", ch)
					}
				}
				r, ok := c.DecodeRune(code)
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
		var code []State
		for _, ch := range currentCharacter {
			switch ch {
			case '.':
				code = append(code, StateDit)
			case '-':
				code = append(code, StateDah)
			default:
				return "", fmt.Errorf("unexpected character in code: %c", ch)
			}
		}
		r, ok := c.DecodeRune(code)
		if !ok {
			return "", fmt.Errorf("failed to decode character: %s", currentCharacter)
		}
		result += string(r)
		currentCharacter = ""
	}
	return result, nil
}

func newCodeFromMap(encodeMap map[rune][]State) Code {
	return &mapCode{encodeMap: encodeMap}
}
