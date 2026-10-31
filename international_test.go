package morse_test

import (
	"testing"

	"github.com/jwetzell/morse-go"
	"github.com/stretchr/testify/require"
)

func TestGoodIntlMorseCode(t *testing.T) {
	testCases := []struct {
		name     string
		input    rune
		expected []morse.State
	}{
		{"A", 'A', []morse.State{morse.StateDit, morse.StateDah}},
		{"B", 'B', []morse.State{morse.StateDah, morse.StateDit, morse.StateDit, morse.StateDit}},
		{"C", 'C', []morse.State{morse.StateDah, morse.StateDit, morse.StateDah, morse.StateDit}},
		{"D", 'D', []morse.State{morse.StateDah, morse.StateDit, morse.StateDit}},
		{"E", 'E', []morse.State{morse.StateDit}},
		{"F", 'F', []morse.State{morse.StateDit, morse.StateDit, morse.StateDah, morse.StateDit}},
		{"G", 'G', []morse.State{morse.StateDah, morse.StateDah, morse.StateDit}},
		{"H", 'H', []morse.State{morse.StateDit, morse.StateDit, morse.StateDit, morse.StateDit}},
		{"I", 'I', []morse.State{morse.StateDit, morse.StateDit}},
		{"J", 'J', []morse.State{morse.StateDit, morse.StateDah, morse.StateDah, morse.StateDah}},
		{"K", 'K', []morse.State{morse.StateDah, morse.StateDit, morse.StateDah}},
		{"L", 'L', []morse.State{morse.StateDit, morse.StateDah, morse.StateDit, morse.StateDit}},
		{"M", 'M', []morse.State{morse.StateDah, morse.StateDah}},
		{"N", 'N', []morse.State{morse.StateDah, morse.StateDit}},
		{"O", 'O', []morse.State{morse.StateDah, morse.StateDah, morse.StateDah}},
		{"P", 'P', []morse.State{morse.StateDit, morse.StateDah, morse.StateDah, morse.StateDit}},
		{"Q", 'Q', []morse.State{morse.StateDah, morse.StateDah, morse.StateDit, morse.StateDah}},
		{"R", 'R', []morse.State{morse.StateDit, morse.StateDah, morse.StateDit}},
		{"S", 'S', []morse.State{morse.StateDit, morse.StateDit, morse.StateDit}},
		{"T", 'T', []morse.State{morse.StateDah}},
		{"U", 'U', []morse.State{morse.StateDit, morse.StateDit, morse.StateDah}},
		{"V", 'V', []morse.State{morse.StateDit, morse.StateDit, morse.StateDit, morse.StateDah}},
		{"W", 'W', []morse.State{morse.StateDit, morse.StateDah, morse.StateDah}},
		{"X", 'X', []morse.State{morse.StateDah, morse.StateDit, morse.StateDit, morse.StateDah}},
		{"Y", 'Y', []morse.State{morse.StateDah, morse.StateDit, morse.StateDah, morse.StateDah}},
		{"Z", 'Z', []morse.State{morse.StateDah, morse.StateDah, morse.StateDit, morse.StateDit}},
		{"0", '0', []morse.State{morse.StateDah, morse.StateDah, morse.StateDah, morse.StateDah, morse.StateDah}},
		{"1", '1', []morse.State{morse.StateDit, morse.StateDah, morse.StateDah, morse.StateDah, morse.StateDah}},
		{"2", '2', []morse.State{morse.StateDit, morse.StateDit, morse.StateDah, morse.StateDah, morse.StateDah}},
		{"3", '3', []morse.State{morse.StateDit, morse.StateDit, morse.StateDit, morse.StateDah, morse.StateDah}},
		{"4", '4', []morse.State{morse.StateDit, morse.StateDit, morse.StateDit, morse.StateDit, morse.StateDah}},
		{"5", '5', []morse.State{morse.StateDit, morse.StateDit, morse.StateDit, morse.StateDit, morse.StateDit}},
		{"6", '6', []morse.State{morse.StateDah, morse.StateDit, morse.StateDit, morse.StateDit, morse.StateDit}},
		{"7", '7', []morse.State{morse.StateDah, morse.StateDah, morse.StateDit, morse.StateDit, morse.StateDit}},
		{"8", '8', []morse.State{morse.StateDah, morse.StateDah, morse.StateDah, morse.StateDit, morse.StateDit}},
		{"9", '9', []morse.State{morse.StateDah, morse.StateDah, morse.StateDah, morse.StateDah, morse.StateDit}},
		{"&", '&', []morse.State{morse.StateDit, morse.StateDah, morse.StateDit, morse.StateDit, morse.StateDit}},
		{"'", '\'', []morse.State{morse.StateDit, morse.StateDah, morse.StateDah, morse.StateDah, morse.StateDah, morse.StateDit}},
		{"@", '@', []morse.State{morse.StateDit, morse.StateDah, morse.StateDah, morse.StateDit, morse.StateDah, morse.StateDit}},
		{")", ')', []morse.State{morse.StateDah, morse.StateDit, morse.StateDah, morse.StateDah, morse.StateDit, morse.StateDah}},
		{"(", '(', []morse.State{morse.StateDah, morse.StateDit, morse.StateDah, morse.StateDah, morse.StateDit}},
		{":", ':', []morse.State{morse.StateDah, morse.StateDah, morse.StateDah, morse.StateDit, morse.StateDit, morse.StateDit}},
		{",", ',', []morse.State{morse.StateDah, morse.StateDah, morse.StateDit, morse.StateDit, morse.StateDah, morse.StateDah}},
		{"=", '=', []morse.State{morse.StateDah, morse.StateDit, morse.StateDit, morse.StateDit, morse.StateDah}},
		{"!", '!', []morse.State{morse.StateDah, morse.StateDit, morse.StateDah, morse.StateDit, morse.StateDah, morse.StateDah}},
		{".", '.', []morse.State{morse.StateDit, morse.StateDah, morse.StateDit, morse.StateDah, morse.StateDit, morse.StateDah}},
		{"-", '-', []morse.State{morse.StateDah, morse.StateDit, morse.StateDit, morse.StateDit, morse.StateDit, morse.StateDah}},
		{"+", '+', []morse.State{morse.StateDit, morse.StateDah, morse.StateDit, morse.StateDah, morse.StateDit}},
		{"\"", '"', []morse.State{morse.StateDit, morse.StateDah, morse.StateDit, morse.StateDit, morse.StateDah, morse.StateDit}},
		{"?", '?', []morse.State{morse.StateDit, morse.StateDit, morse.StateDah, morse.StateDah, morse.StateDit, morse.StateDit}},
		{"/", '/', []morse.State{morse.StateDah, morse.StateDit, morse.StateDit, morse.StateDah, morse.StateDit}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := morse.IntlMorseCode.EncodeRune(tc.input)
			require.True(t, ok)
			require.Equal(t, tc.expected, got)
		})
	}
}
