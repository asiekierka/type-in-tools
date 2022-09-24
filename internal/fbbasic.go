// Copyright (c) 2022 Adrian Siekierka
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package internal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var idToKeywordMap = map[byte]string{
	0x80: "GOTO",
	0x81: "GOSUB",
	0x82: "RUN",
	0x83: "RETURN",
	0x84: "RESTORE",
	0x85: "THEN",
	0x86: "LIST",
	0x87: "SYSTEM",
	0x88: "TO",
	0x89: "STEP",
	0x8A: "SPRITE",
	0x8B: "PRINT",
	0x8C: "FOR",
	0x8D: "NEXT",
	0x8E: "PAUSE",
	0x8F: "INPUT",
	0x90: "LINPUT",
	0x91: "DATA",
	0x92: "IF",
	0x93: "READ",
	0x94: "DIM",
	0x95: "REM",
	0x96: "STOP",
	0x97: "CONT",
	0x98: "CLS",
	0x99: "CLEAR",
	0x9A: "ON",
	0x9B: "OFF",
	0x9C: "CUT",
	0x9D: "NEW",
	0x9E: "POKE",
	0x9F: "CGSET",
	0xA0: "VIEW",
	0xA1: "MOVE",
	0xA2: "END",
	0xA3: "PLAY",
	0xA4: "BEEP",
	0xA5: "LOAD",
	0xA6: "SAVE",
	0xA7: "POSITION",
	0xA8: "KEY",
	0xA9: "COLOR",
	0xAA: "DEF",
	0xAB: "CGEN",
	0xAC: "SWAP",
	0xAD: "CALL",
	0xAE: "LOCATE",
	0xAF: "PALET",
	0xB0: "ERA",

	0xB1: "TR",     /* Family Basic V3 */
	0xB2: "FIND",   /* Family Basic V3 */
	0xB3: "GAME",   /* Family Basic V3 */
	0xB4: "BGTOOL", /* Family Basic V3 */
	0xB5: "AUTO",   /* Family Basic V3 */
	0xB6: "DELETE", /* Family Basic V3 */
	0xB7: "RENUM",  /* Family Basic V3 */
	0xB8: "FILTER", /* Family Basic V3 */
	0xB9: "CLICK",  /* Family Basic V3 */
	0xBA: "SCREEN", /* Family Basic V3 */
	0xBB: "BACKUP", /* Family Basic V3 */
	0xBC: "ERROR",  /* Family Basic V3 */
	0xBD: "RESUME", /* Family Basic V3 */
	0xBE: "BGPUT",  /* Family Basic V3 */
	0xBF: "BGGET",  /* Family Basic V3 */
	0xC0: "CAN",    /* Family Basic V3 */

	0xCA: "ABS",
	0xCB: "ASC",
	0xCC: "STR$",
	0xCD: "FRE",
	0xCE: "LEN",
	0xCF: "PEEK",
	0xD0: "RND",
	0xD1: "SGN",
	0xD2: "SPC",
	0xD3: "TAB",
	0xD4: "MID$",
	0xD5: "STICK",
	0xD6: "STRIG",
	0xD7: "XPOS",
	0xD8: "YPOS",
	0xD9: "VAL",
	0xDA: "POS",
	0xDB: "CSRLIN",
	0xDC: "CHR$",
	0xDD: "HEX$",
	0xDE: "INKEY$",
	0xDF: "RIGHT$",
	0xE0: "LEFT$",
	0xE1: "SCR$",

	0xE2: "INSTR", /* Family Basic V3 */
	0xE3: "CRASH", /* Family Basic V3 */
	0xE4: "ERR",   /* Family Basic V3 */
	0xE5: "ERL",   /* Family Basic V3 */
	0xE6: "VCT",   /* Family Basic V3 */

	0xEF: "XOR",
	0xF0: "OR",
	0xF1: "AND",
	0xF2: "NOT",
	0xF3: "<>",
	0xF4: ">=",
	0xF5: "<=",
	0xF6: "=",
	0xF7: ">",
	0xF8: "<",
	0xF9: "+",
	0xFA: "-",
	0xFB: "MOD",
	0xFC: "/",
	0xFD: "*",
}

func hexToStringSpaces(buf []byte) string {
	var s strings.Builder
	s.WriteString("[")
	for i := 0; i < len(buf); i++ {
		if i > 0 {
			s.WriteString(" ")
		}
		s.WriteString(fmt.Sprintf("%02X", buf[i]))
	}
	s.WriteString("]")
	return s.String()
}

func FBBasicBinToString(reader io.Reader) (string, error) {
	var s strings.Builder
	endOfProgram := false

	for !endOfProgram {
		var nextOffset byte
		binary.Read(reader, binary.LittleEndian, &nextOffset)
		if nextOffset == 0 {
			break
		}

		lineLength := int(nextOffset) - 3
		var lineNumber uint16
		binary.Read(reader, binary.LittleEndian, &lineNumber)
		s.WriteString(strconv.Itoa(int(lineNumber)))
		s.WriteString(" ")

		lineData := make([]byte, lineLength)
		reader.Read(lineData)
		parsingComment := false
		parsingString := false
		nextNumberNegative := 0

		for i := 0; i < len(lineData); i++ {
			nextNumberNegative -= 1
			id := lineData[i]
			if id == 0x00 {
				// end of line
				break
			} else if parsingComment {
				s.WriteString(FBByteToString(id))
			} else if parsingString {
				s.WriteString(FBByteToString(id))
				if id == '"' {
					parsingString = false
				}
			} else if id >= 0x80 {
				keyword, ok := idToKeywordMap[id]
				if !ok {
					fmt.Fprintf(os.Stderr, "skipping line %d, unknown token 0x%02X (%s)\n", lineNumber, id, hexToStringSpaces(lineData[i:]))
					break
				}
				s.WriteString(keyword)
				if id == 0x95 {
					// REM acts as comment
					parsingComment = true
				}
			} else if id == '\'' {
				parsingComment = true
				s.WriteString("'")
			} else if id == '"' {
				parsingString = true
				s.WriteString("\"")
			} else if id >= 0x20 && id <= 0x5B {
				s.WriteString(FBByteToString(id))
			} else if id == 0x12 {
				// constant number
				i++
				v := int(lineData[i])
				i++
				v = v | (int(lineData[i]) << 8)
				if nextNumberNegative == 1 {
					v = -v
				}
				s.WriteString(strconv.Itoa(v))
			} else if id == 0x11 {
				// hex number
				i++
				v := int(lineData[i])
				i++
				v = v | (int(lineData[i]) << 8)
				s.WriteString(fmt.Sprintf("&H%X", v))
			} else if id == 0x0B {
				// line number
				i++
				v := int(lineData[i])
				i++
				v = v | (int(lineData[i]) << 8)
				s.WriteString(strconv.Itoa(v))
			} else if id == 0xFA {
				nextNumberNegative = 2
			} else if id >= 0x01 && id <= 0x0A {
				// constant number (short)
				v := int(id) - 1
				if nextNumberNegative == 1 {
					v = -v
				}
				s.WriteString(strconv.Itoa(v))
			} else {
				fmt.Fprintf(os.Stderr, "skipping line %d, unknown token 0x%02X (%v)\n", lineNumber, id, hexToStringSpaces(lineData[i:]))
				break
			}
		}
		s.WriteString("\n")
	}

	return s.String(), nil
}

func FBBasicStringToBin(s string, writer io.Writer) error {
	for _, line := range strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n") {
		var lineBuf bytes.Buffer
		lineArray := strings.SplitN(line, " ", 2)
		if len(lineArray) != 2 {
			continue
		}
		lineNumber, err := strconv.Atoi(lineArray[0])
		if err != nil {
			continue
		}
		s := lineArray[1]
		readingLineNumbers := false
		currAlpha := false
		lastAlpha := false
		lastLen := 0
		currLen := 0
		for len(s) > 0 {
			prefixByte := byte(0)
			prefixLen := 0
			for k, v := range idToKeywordMap {
				if len(v) > prefixLen && strings.HasPrefix(s, v) {
					prefixByte = k
					prefixLen = len(v)
				}
			}
			lastLen = currLen
			currLen = len(s)
			if lastLen == currLen {
				return fmt.Errorf("cannot parse line %d beyond '%s' (stuck)", lineNumber, s)
			}
			lastAlpha = currAlpha
			currAlpha = false
			if s[0] == 0x20 {
				// skip spaces early (so "GOTO"[space]"line number") works
				lineBuf.WriteByte(0x20)
				s = s[1:]
				continue
			}
			if prefixLen > 0 {
				lineBuf.WriteByte(prefixByte)
				currKeyword := s[:prefixLen]
				s = s[prefixLen:]
				if currKeyword == "REM" || currKeyword == "DATA" {
					// comment
					lineBuf.Write(FBStringToBytes(s))
					break
				}
				readingLineNumbers = (currKeyword == "GOSUB" || currKeyword == "GOTO" || currKeyword == "RETURN" || currKeyword == "RESTORE" || currKeyword == "RUN" || currKeyword == "THEN")
			} else if s[0] == '\'' {
				// comment
				lineBuf.Write(FBStringToBytes(s))
				break
			} else if s[0] == '"' {
				// string
				lineBuf.WriteByte('"')
				s = s[1:]
				splitPos := strings.Index(s, "\"")
				if splitPos < 0 {
					lineBuf.Write(FBStringToBytes(s))
					break
				} else {
					lineBuf.Write(FBStringToBytes(s[0 : splitPos+1]))
				}
				s = s[splitPos+1:]
			} else if !lastAlpha && (s[0] >= '0' && s[0] <= '9' || (len(s) >= 2 && s[0] == '-' && s[1] >= '0' && s[1] <= '9')) {
				// digits
				x := 0
				if s[0] == '-' {
					// negative value marker
					lineBuf.WriteByte(0xFA)
					x++
				}
				xStart := x
				for x < len(s) && s[x] >= '0' && s[x] <= '9' {
					x++
				}
				v, err := strconv.Atoi(s[xStart:x])
				if err != nil {
					return err
				}
				if v >= 65536 {
					return fmt.Errorf("value out of range: %d", v)
				}
				if readingLineNumbers {
					// line number
					lineBuf.WriteByte(0x0B)
					lineBuf.WriteByte(byte(v & 0xFF))
					lineBuf.WriteByte(byte((v >> 8) & 0xFF))
				} else if v >= 10 {
					// long constant number
					lineBuf.WriteByte(0x12)
					lineBuf.WriteByte(byte(v & 0xFF))
					lineBuf.WriteByte(byte((v >> 8) & 0xFF))
				} else {
					// short constant number
					lineBuf.WriteByte(byte(v + 1))
				}
				s = s[x:]
			} else if len(s) >= 2 && s[0] == '&' && s[1] == 'H' {
				// hex number
				x := 2
				xStart := x
				for x < len(s) && ((s[x] >= '0' && s[x] <= '9') || (s[x] >= 'A' && s[x] <= 'F')) {
					x++
				}
				v, err := strconv.ParseInt(s[xStart:x], 16, 32)
				if err != nil {
					return err
				}
				if v >= 65536 {
					return fmt.Errorf("value out of range: %d", v)
				}
				lineBuf.WriteByte(0x11)
				lineBuf.WriteByte(byte(v & 0xFF))
				lineBuf.WriteByte(byte((v >> 8) & 0xFF))
				s = s[x:]
			} else if s[0] >= 0x21 && s[0] <= 0x5B {
				// other character
				lineBuf.WriteByte(s[0])
				if s[0] != ',' {
					readingLineNumbers = false
				}
				if s[0] >= 0x41 && s[0] <= 0x5B {
					currAlpha = true
				}
				s = s[1:]
			} else {
				return fmt.Errorf("cannot parse line %d beyond '%s'", lineNumber, s)
			}
		}
		lineBuf.WriteByte(0x00)

		lineData := lineBuf.Bytes()
		if len(lineData) >= 253 {
			return fmt.Errorf("line too long")
		}
		writer.Write([]byte{byte(len(lineData) + 3), byte(lineNumber & 0xFF), byte((lineNumber >> 8) & 0xFF)})
		writer.Write(lineData)
	}
	writer.Write([]byte{0x00})

	return nil
}
