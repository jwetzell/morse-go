package morse

import "fmt"

var codeMap map[rune]string

var intlMorseCodeMap = map[rune]string{
	'A':  ".-",
	'B':  "-...",
	'C':  "-.-.",
	'D':  "-..",
	'E':  ".",
	'F':  "..-.",
	'G':  "--.",
	'H':  "....",
	'I':  "..",
	'J':  ".---",
	'K':  "-.-",
	'L':  ".-..",
	'M':  "--",
	'N':  "-.",
	'O':  "---",
	'P':  ".--.",
	'Q':  "--.-",
	'R':  ".-.",
	'S':  "...",
	'T':  "-",
	'U':  "..-",
	'V':  "...-",
	'W':  ".--",
	'X':  "-..-",
	'Y':  "-.--",
	'Z':  "--..",
	'0':  "-----",
	'1':  ".----",
	'2':  "..---",
	'3':  "...--",
	'4':  "....-",
	'5':  ".....",
	'6':  "-....",
	'7':  "--...",
	'8':  "---..",
	'9':  "----.",
	'&':  ".-...",
	'\'': ".----.",
	'@':  ".--.-.",
	')':  "-.--.-",
	'(':  "-.--.",
	':':  "---...",
	',':  "--..--",
	'=':  "-...-",
	'!':  "-.-.--",
	'.':  ".-.-.-",
	'-':  "-....-",
	'+':  ".-.-.",
	'"':  ".-..-.",
	'?':  "..--..",
	'/':  "-..-.",
}

func init() {
	codeMap = make(map[rune]string)
	for char, code := range intlMorseCodeMap {
		codeMap[char] = code
	}
}

func SetCodeMap(newMap map[rune]string) {
	for char, code := range newMap {
		codeMap[char] = code
	}
}

func ASCIIFromDitDahs(ditDahs string) (string, error) {
	var result string
	for _, char := range ditDahs {
		if char != '.' && char != '-' {
			return "", fmt.Errorf("invalid character: %c", char)
		}
	}

	for char, coded := range codeMap {
		if coded == ditDahs {
			result += string(char)
			break
		}
	}

	return result, nil
}

// TODO(jwetzell): super basic, only support a single character and international morse code timings
func CodeListToDitDahs(codeList []int32, ditMax int32) string {
	var result string
	for _, transition := range codeList {
		if transition > 0 {
			if transition > ditMax {
				result += "-"
			} else {
				result += "."
			}
		}
	}
	return result
}
