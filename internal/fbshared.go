package internal

import (
	"bytes"
	"fmt"
)

var nameHighToChar = []rune("「￥」^_アイウエオカキクケコサシスセソタチツテトナニヌネノハヒフヘホマミムメモヤユヨラリルレロワンヲァィゥェォャュョッガギグゲゴザジズゼゾダヂヅデドバビブベボパピプペポ□▫[]Ⓒ×÷")
var stringToChar = make(map[string]byte)
var stringToCharMax = 0

func FBByteToString(c byte) string {
	if c < 0x20 {
		return fmt.Sprintf("\\%c%d", (c>>3)+'A', (c & 7))
	} else if c >= 0x20 && c <= 0x5A {
		return string(rune(c))
	} else if c >= 0x5B && int(c) < (0x5B+len(nameHighToChar)) {
		return string(nameHighToChar[c-0x5B])
	} else if c >= 0xB8 {
		return fmt.Sprintf("\\%c%d", ((c-0xB8)>>3)+'D', (c & 7))
	} else if c == 0xB7 {
		return "\\xB7"
	} else {
		panic(fmt.Errorf("character not found: %d", c))
	}
}

func FBStringToBytes(s string) []byte {
	var buf bytes.Buffer

	for len(s) > 0 {
		maxLen := len(s)
		if maxLen > stringToCharMax {
			maxLen = stringToCharMax
		}
		for i := maxLen; i > 0; i-- {
			v, ok := stringToChar[s[0:i]]
			if ok {
				buf.WriteByte(v)
				s = s[i:]
				break
			}
		}
	}

	return buf.Bytes()
}

func init() {
	for i := 0; i < 0x100; i++ {
		s := FBByteToString(byte(i))
		stringToChar[s] = byte(i)
		if len(s) > stringToCharMax {
			stringToCharMax = len(s)
		}
	}
}
