package internal

import "fmt"

var nameHighToChar = []rune("「￥」^_アイウエオカキクケコサシスセソタチツテトナニヌネノハヒフヘホマミムメモヤユヨラリルレロワンヲァィゥェォャュョッガギグゲゴザジズゼゾダヂヅデドバビブベボパピプペポ□▫[]Ⓒ×÷")

func ByteToString(c byte) string {
	if c >= 0x20 && c <= 0x5A {
		return string(rune(c))
	} else if c >= 0x5B && int(c) < (0x5B+len(nameHighToChar)) {
		return string(nameHighToChar[c-0x5B])
	} else {
		return string(fmt.Sprintf("\\x%02X", c))
	}
}
