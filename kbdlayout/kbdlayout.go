package kbdlayout

import "fmt"
import "log"
import "unicode"
import "unicode/utf8"

// layout[i] is a number 0..255 which corresponds an character id
// in a keyboard mapping, while i is 0..29 the location on keyboard:
//
//  0  1  2  3   4    5   6  7  8  9
// 10 11 12 13  14   15  16 17 18 19
// 20 21 22 23  24   25  26 27 28 29
//
// for example on qwerty:
// map.ID2Rune[layout[0]] == 'q'
// map.ID2Rune[layout[1]] == 'w'
// map.ID2Rune[layout[10]] == 'a'
type KeyboardLayout [30]uint8

type KeyboardMapping struct {
	ID2Rune []rune // will be limited to 256, as we're using uint8
	Rune2ID map[rune]uint8
}

func NewMapping(keys string) *KeyboardMapping {
	runeCount := utf8.RuneCountInString(keys)
	if runeCount > 256 {
		log.Fatalf("KeyboardMapping currently supports only up to 256 characters. (got %d)", runeCount)
	}
	mapping := &KeyboardMapping{
		ID2Rune: make([]rune, runeCount),
		Rune2ID: make(map[rune]uint8),
	}

	for i := 0; i < runeCount; i++ {
		character, size := utf8.DecodeRuneInString(keys)
		character = unicode.ToLower(character)
		keys = keys[size:]
		mapping.ID2Rune[i] = character
		mapping.Rune2ID[character] = uint8(i) // will stay below 256
	}

	return mapping
}

const (
	Qwerty  = "qwertyuiopasdfghjkl;zxcvbnm.,/"
	Abcde   = "abcdefghijklmnopqrstuvwxyz.,;/"
	Dvorak  = "/,.pyfgcrlaoeuidhtns;qjkxbmwvz"
	Colemak = "qwfpgjluy;arstdhneiozxcvbkm.,/"
	Asset   = "qwjfgypul;asetdhniorzxcvbkm,./"
	Workman = "qdrwbjfup;ashtgyneoizxmcvkl,./"
	Nail    = ",.cgkxbou;therfdnail/pysqjmwvz"
	Layman  = "/pu.xqfcgyaserlhntoikv;dzbmw,j"
)

func NewLayout(l string, m *KeyboardMapping) KeyboardLayout {
	layout := [30]uint8{}
	for i := 0; i < 30; i++ {
		character, size := utf8.DecodeRuneInString(l)
		character = unicode.ToLower(character)
		id, ok := m.Rune2ID[character]
		if !ok {
			log.Fatalf("could not map %c on the layout", character)
		}
		layout[i] = id
		l = l[size:]
	}
	return KeyboardLayout(layout)
}

func (m *KeyboardMapping) PrintLayout(l *KeyboardLayout) {
	k := [30]rune{}
	for i := 0; i < 30; i++ {
		k[i] = m.ID2Rune[l[i]]
	}
	fmt.Printf("%c%c%c%c %c  %c %c%c%c%c\n", k[0], k[1], k[2], k[3], k[4], k[5], k[6], k[7], k[8], k[9])
	fmt.Printf("%c%c%c%c %c  %c %c%c%c%c\n", k[10], k[11], k[12], k[13], k[14], k[15], k[16], k[17], k[18], k[19])
	fmt.Printf("%c%c%c%c %c  %c %c%c%c%c\n", k[20], k[21], k[22], k[23], k[24], k[25], k[26], k[27], k[28], k[29])
}
