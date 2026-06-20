package morse

type Code interface {
	Encode(rune) (string, bool)
	Decode(string) (rune, bool)
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

func newCodeFromMap(encodeMap map[rune]string) Code {
	return &mapCode{encodeMap: encodeMap}
}
