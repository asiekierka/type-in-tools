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
	"fmt"
	"io"
	"strconv"
)

const (
	FBNameTableOffsetX = 2
	FBNameTableOffsetY = 3
	FBNameTableWidth   = 28
	FBNameTableHeight  = 21
)

type FBNameTable struct {
	Tiles      [960]byte
	Attributes [64]byte
}

func NewFBNameTable() FBNameTable {
	nt := FBNameTable{}
	for i := 0; i < 960; i++ {
		nt.Tiles[i] = 0x20
	}
	return nt
}

func (f *FBNameTable) Read(reader io.Reader) error {
	_, err := reader.Read(f.Tiles[:])
	if err != nil {
		return err
	}

	_, err = reader.Read(f.Attributes[:])
	if err != nil {
		return err
	}

	return nil
}

func (f *FBNameTable) Write(writer io.Writer) error {
	_, err := writer.Write(f.Tiles[:])
	if err != nil {
		return err
	}

	_, err = writer.Write(f.Attributes[:])
	if err != nil {
		return err
	}

	return nil
}

func (f *FBNameTable) GetString(x, y int) string {
	nx := x + FBNameTableOffsetX
	ny := y + FBNameTableOffsetY
	tile := f.Tiles[ny*32+nx]

	tileString := ""
	isGfx := false
	if tile < 0x20 {
		tileString = fmt.Sprintf("%c%d", (tile>>3)+'A', (tile & 7))
		isGfx = true
	} else if tile < 0xB8 {
		tileString = FBByteToString(tile)
	} else {
		tileString = fmt.Sprintf("%c%d", ((tile-0xB8)>>3)+'E', (tile & 7))
		isGfx = true
	}

	attrShift := ((ny>>1)&1)*4 + ((nx>>1)&1)*2
	attr := (f.Attributes[(ny>>2)*8+(nx>>2)] >> byte(attrShift)) & 3

	if isGfx {
		tileString += strconv.Itoa(int(attr))
	}

	return tileString
}

func (f *FBNameTable) PutString(x, y int, tileString string) {
	if tileString == "" {
		tileString = " "
	}

	nx := x + FBNameTableOffsetX
	ny := y + FBNameTableOffsetY

	tile := byte(0)

	if len(tileString) >= 2 && tileString[0] >= 'A' && tileString[0] <= 'M' && tileString[1] >= '0' && tileString[1] <= '7' {
		// gfx mode
		if tileString[0] <= 'D' {
			tile = byte((tileString[0]-'A')*8 + (tileString[1] - '0'))
		} else {
			tile = byte(0xB8 + (tileString[0]-'E')*8 + (tileString[1] - '0'))
		}

		if len(tileString) >= 3 && tileString[2] >= '0' && tileString[2] <= '3' {
			attrVal := tileString[2] - '0'
			attrShift := ((ny>>1)&1)*4 + ((nx>>1)&1)*2
			attr := f.Attributes[(ny>>2)*8+(nx>>2)]
			attr = (attr & ((3 << attrShift) ^ 0xFF)) | (attrVal << byte(attrShift))
			f.Attributes[(ny>>2)*8+(nx>>2)] = attr
		}
	} else {
		// text mode
		tile = FBStringToBytes(tileString)[0]
	}

	f.Tiles[ny*32+nx] = tile
}
