package internal

import "math/bits"

type BlockType uint8

const (
	BlockUnknown BlockType = iota
	BlockInformation
	BlockData
)

func CalcFBChecksum(data []byte) uint16 {
	ck := 0
	for _, v := range data {
		ck += bits.OnesCount8(v)
	}
	return uint16(ck)
}
