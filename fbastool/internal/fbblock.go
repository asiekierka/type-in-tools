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
	"encoding/binary"
	"fmt"
	"math/bits"
	"strings"
)

type RawBlockType uint8
type FBFileType uint8

const (
	RawBlockUnknown RawBlockType = iota
	RawBlockInfo
	RawBlockData

	FileTypeBasic      FBFileType = 2
	FileTypeBgGraphics FBFileType = 3
)

type FBFileInfo struct {
	Type             FBFileType
	Name             [16]byte
	Reserved1        byte
	Length           uint16
	LoadAddress      uint16
	ExecutionAddress uint16
	Pad              [104]byte
}

type FBFile struct {
	Info FBFileInfo
	Data []byte
}

func (tp FBFileType) String() string {
	if tp == FileTypeBasic {
		return "BASIC"
	} else if tp == FileTypeBgGraphics {
		return "BG-GRAPHICS"
	} else {
		return "Unknown"
	}
}

func CalcDataChecksum(data []byte) uint16 {
	ck := 0
	for _, v := range data {
		ck += bits.OnesCount8(v)
	}
	return uint16(ck)
}

func (i FBFileInfo) NameStr() string {
	s := ""
	for _, c := range i.Name {
		if c == 0x00 {
			return s
		} else {
			s += FBByteToString(c)
		}
	}
	return s
}

func (i *FBFileInfo) SetName(s string) {
	v, _ := FBStringToBytes(strings.ToUpper(s))
	copy(i.Name[:], v)
	if len(v) < 16 {
		i.Name[len(v)] = 0
	}
}

func (i FBFileInfo) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 128)
	buf[0] = byte(i.Type)
	copy(buf[1:17], i.Name[:])
	buf[17] = byte(i.Reserved1)
	binary.LittleEndian.PutUint16(buf[18:], i.Length)
	binary.LittleEndian.PutUint16(buf[20:], i.LoadAddress)
	binary.LittleEndian.PutUint16(buf[22:], i.ExecutionAddress)
	copy(buf[24:128], i.Pad[:])
	return buf, nil
}

func (i *FBFileInfo) UnmarshalBinary(buf []byte) error {
	if len(buf) < 128 {
		return fmt.Errorf("buffer too small: %d < 128", len(buf))
	}

	if buf[0] != byte(FileTypeBasic) && buf[0] != byte(FileTypeBgGraphics) {
		return fmt.Errorf("unknown file type: %d", buf[0])
	}
	i.Type = FBFileType(buf[0])

	copy(i.Name[:], buf[1:17])
	i.Length = binary.LittleEndian.Uint16(buf[18:])
	i.LoadAddress = binary.LittleEndian.Uint16(buf[20:])
	i.ExecutionAddress = binary.LittleEndian.Uint16(buf[22:])
	copy(i.Pad[:], buf[24:128])

	return nil
}
