package internal

import (
	"errors"
	"fmt"
	"io"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

type pulseType uint8

const (
	FAMICOM_FREQUENCY = 1789773

	pulseUnknown pulseType = iota
	pulseShort
	pulseLong
)

type TapeEncodingInfo struct {
	CyclesPerByte   int
	ShortPulseWidth int
	LongPulseWidth  int
	SyncPulseCount  int
	PulseTolerance  float32
	SyncTolerance   float32
}

func NewTapeEncodingInfo() TapeEncodingInfo {
	return TapeEncodingInfo{
		CyclesPerByte:   45,
		ShortPulseWidth: 20,
		LongPulseWidth:  40,
		SyncPulseCount:  20000,
		PulseTolerance:  1.25,
		SyncTolerance:   1.2,
	}
}

func (info *TapeEncodingInfo) TapeFrequency() float64 {
	return FAMICOM_FREQUENCY / float64(info.CyclesPerByte)
}

func (info *TapeEncodingInfo) getPulseType(pulseSamples int64, sampleRate uint32) pulseType {
	if pulseSamples <= 0 {
		return pulseUnknown
	}

	tapeFrequency := info.TapeFrequency()
	tapePulseSamples := float32(float64(pulseSamples) * tapeFrequency / float64(sampleRate))
	// fmt.Printf("%f\n", tapePulseSamples)

	if tapePulseSamples >= (float32(info.ShortPulseWidth)/info.PulseTolerance) && tapePulseSamples <= (float32(info.ShortPulseWidth)*info.PulseTolerance) {
		return pulseShort
	} else if tapePulseSamples >= (float32(info.LongPulseWidth)/info.PulseTolerance) && tapePulseSamples <= (float32(info.LongPulseWidth)*info.PulseTolerance) {
		return pulseLong
	} else {
		return pulseUnknown
	}
}

type TapeReader struct {
	reader            io.ReadSeeker
	wav               *wav.Decoder
	encInfo           TapeEncodingInfo
	buffer            *audio.IntBuffer
	audioSampleOffset int

	// hacks...
	peekedBit byte
}

func NewTapeReader(reader io.ReadSeeker, encInfo TapeEncodingInfo) (*TapeReader, error) {
	tapeReader := TapeReader{
		reader:    reader,
		encInfo:   encInfo,
		peekedBit: 255,
	}

	wav := wav.NewDecoder(reader)
	wav.ReadInfo()
	tapeReader.wav = wav
	tapeReader.buffer = &audio.IntBuffer{Data: make([]int, wav.NumChans)}
	if tapeReader.wav.BitDepth == 16 {
		tapeReader.audioSampleOffset = 0
	} else if tapeReader.wav.BitDepth == 8 {
		tapeReader.audioSampleOffset = 128
	} else {
		return nil, errors.New("could not read wave file")
	}

	return &tapeReader, nil
}

func (reader *TapeReader) GetPosition() int64 {
	pos, _ := reader.wav.Seek(0, io.SeekCurrent)
	return pos
}

func (reader *TapeReader) SetPosition(pos int64) {
	reader.wav.Seek(pos, io.SeekStart)
}

// TODO: Handle non-pristine tapes.
func (reader *TapeReader) nextPulse() (int64, error) {
	pos := int64(0)
	stage := 0

	prevSample := 0

	for {
		count, err := reader.wav.PCMBuffer(reader.buffer)
		if err != nil {
			return 0, err
		} else if count <= 0 {
			return 0, errors.New("end of file")
		}
		sample := 0
		for _, s := range reader.buffer.Data {
			sample += s - reader.audioSampleOffset
		}

		pos += 1

		if prevSample < 0 && sample > 0 {
			stage += 1
			if stage == 1 {
				pos = 0
			}
		} else if prevSample > 0 && sample < 0 {
			stage += 1
			if stage == 1 {
				pos = 0
			}
		}

		if stage >= 2 {
			return pos * 2, nil
		}

		prevSample = sample
	}

}

func (reader *TapeReader) RewindBit(bit byte) {
	reader.peekedBit = bit
}

func (reader *TapeReader) NextBit() (byte, error) {
	if reader.peekedBit != 255 {
		bit := reader.peekedBit
		reader.peekedBit = 255
		return bit, nil
	}
	pulse, err := reader.nextPulse()
	if err != nil {
		return 255, err
	}
	ptype := reader.encInfo.getPulseType(pulse, reader.wav.SampleRate)
	switch ptype {
	case pulseShort:
		return 0, nil
	case pulseLong:
		return 1, nil
	default:
		return 255, nil
	}
}

func (reader *TapeReader) VerifyBit(bit byte) error {
	actual, err := reader.NextBit()
	if err != nil {
		return err
	} else if actual != bit {
		return fmt.Errorf("%d expected, %d actual", bit, actual)
	} else {
		return nil
	}
}

// FIXME: broken
/* func (reader *TapeReader) peekNextBit() (uint8, int64, error) {
	prevPosition := reader.GetPosition()
	bit, err := reader.NextBit()
	nextPosition := reader.GetPosition()
	reader.SetPosition(prevPosition)
	return bit, nextPosition, err
} */

func (reader *TapeReader) NextByte() (byte, error) {
	bit, err := reader.NextBit()
	if err != nil {
		return 0, err
	} else if bit != 1 {
		return 0, fmt.Errorf("starter bit not 1 (%d)", bit)
	}
	v := byte(0)
	for i := 7; i >= 0; i-- {
		bit, err := reader.NextBit()
		if err != nil {
			return 0, err
		} else if bit == 255 {
			return 0, fmt.Errorf("bit read error")
		} else if bit == 1 {
			v = v | byte(1<<i)
		}
	}
	return v, nil
}

func (reader *TapeReader) NextBytes(len int) ([]byte, error) {
	buffer := make([]byte, len)
	for i := 0; i < len; i++ {
		v, err := reader.NextByte()
		if err != nil {
			return nil, fmt.Errorf("could not read byte %d/%d: %v", i+1, len, err)
		}
		buffer[i] = v
	}
	return buffer, nil
}

func (reader *TapeReader) SyncToBlock() (BlockType, error) {
	// 0 x ~20000 -> 1 x N -> 0 x N
	state := 0
	currentBit := uint8(255)
	bitCount := 0
	firstBitCount := 0
	secondBitCount := 0

	for {
		bit, err := reader.NextBit()
		if err != nil {
			return BlockUnknown, fmt.Errorf("could not find synchronization signal: %v", err)
		} else if bit == 255 {
			continue
		}

		if bit == currentBit {
			bitCount++
		} else {
			reader.RewindBit(bit)
			switch state {
			case 0: /* syncing */
				if currentBit == 0 && bitCount >= int(float32(reader.encInfo.SyncPulseCount)/reader.encInfo.PulseTolerance) && bitCount <= int(float32(reader.encInfo.SyncPulseCount)*reader.encInfo.PulseTolerance) {
					state = 1
				}
			case 1: /* 1 */
				firstBitCount = bitCount
				state = 2
			case 2: /* 0 */
				secondBitCount = bitCount
				if firstBitCount != secondBitCount {
					return BlockUnknown, fmt.Errorf("bit count mismatch on block type (%d != %d)", firstBitCount, secondBitCount)
				} else if firstBitCount == 40 {
					return BlockInformation, nil
				} else if firstBitCount == 20 {
					return BlockData, nil
				} else {
					return BlockUnknown, fmt.Errorf("could not recognize block type (%d)", firstBitCount)
				}
			}
			bitCount = 0
			currentBit = bit
		}
	}
}
