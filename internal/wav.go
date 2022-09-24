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
	CyclesPerByte     int
	ShortPulseWidth   int
	LongPulseWidth    int
	SyncMinPulseCount int
	PulseTolerance    float32
}

func NewTapeEncodingInfo() TapeEncodingInfo {
	return TapeEncodingInfo{
		CyclesPerByte:     45,
		ShortPulseWidth:   20,
		LongPulseWidth:    40,
		SyncMinPulseCount: 5000,
		PulseTolerance:    1.375,
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
	peekedBit         byte
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

	prevSample := -1000000

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

		if prevSample != -1000000 {
			if prevSample < 0 && sample >= 0 {
				stage += 1
			} else if prevSample >= 0 && sample < 0 {
				stage += 1
			}
		}

		if stage >= 2 {
			return pos, nil
		}

		prevSample = sample
	}

}

func (reader *TapeReader) RewindBit(bit byte) {
	reader.peekedBit = bit
}

func (reader *TapeReader) NextBit() (byte, error) {
	if reader.peekedBit != 255 {
		v := reader.peekedBit
		reader.peekedBit = 255
		return v, nil
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

func (reader *TapeReader) NextWord() (uint16, error) {
	b1, err := reader.NextByte()
	if err != nil {
		return 0, fmt.Errorf("could not read low word: %v", err)
	}
	b2, err := reader.NextByte()
	if err != nil {
		return 0, fmt.Errorf("could not read high word: %v", err)
	}
	return uint16(b2) | (uint16(b1) << 8), nil
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

func (reader *TapeReader) NextBytesWithChecksum(len int) ([]byte, uint16, error) {
	data, err := reader.NextBytes(len)
	if err != nil {
		return nil, 0, err
	}

	checksum, err := reader.NextWord()
	if err != nil {
		return data, 0, err
	}

	return data, checksum, nil
}

func checkChecksumAndPrint(data []byte, name string, checksum uint16) {
	actualChecksum := CalcDataChecksum(data)
	if actualChecksum != checksum {
		fmt.Printf("warning: block %s has invalid checksum %d != %d\n", name, checksum, actualChecksum)
	}
}

func (reader *TapeReader) NextFile() (*FBFile, error) {
	blockType, err := reader.SyncToBlock()
	if err != nil {
		return nil, fmt.Errorf("block sync error: %v", err)
	}
	if blockType != RawBlockInfo {
		return nil, errors.New("invalid block type (expected information)")
	}

	err = reader.VerifyBit(1)
	if err != nil {
		return nil, fmt.Errorf("block prelude error: %v", err)
	}

	fbInfoData, fbInfoChecksum, err := reader.NextBytesWithChecksum(128)
	if err != nil {
		return nil, fmt.Errorf("block read error: %v", err)
	}
	checkChecksumAndPrint(fbInfoData, "information", fbInfoChecksum)

	err = reader.VerifyBit(1)
	if err != nil {
		return nil, fmt.Errorf("block postlude error: %v", err)
	}

	fbInfo := FBFileInfo{}
	fbInfo.UnmarshalBinary(fbInfoData)

	blockType, err = reader.SyncToBlock()
	if err != nil {
		return nil, fmt.Errorf("block sync error: %v", err)
	}
	if blockType != RawBlockData {
		return nil, errors.New("invalid block type (expected data)")
	}

	err = reader.VerifyBit(1)
	if err != nil {
		return nil, fmt.Errorf("block prelude error: %v", err)
	}

	fbDataData, fbDataChecksum, err := reader.NextBytesWithChecksum(int(fbInfo.Length))
	if err != nil {
		return nil, fmt.Errorf("block read error: %v", err)
	}
	checkChecksumAndPrint(fbDataData, "data", fbDataChecksum)

	// don't check the final postlude
	/* err = reader.VerifyBit(1)
	if err != nil {
		return nil, fmt.Errorf("block postlude error: %v", err)
	} */

	return &FBFile{
		Info: fbInfo,
		Data: fbDataData,
	}, nil
}

func (reader *TapeReader) SyncToBlock() (RawBlockType, error) {
	state := 0
	currentBit := uint8(255)
	bitCount := 0
	firstBitCount := 0
	secondBitCount := 0

	for {
		bit, err := reader.NextBit()
		if err != nil {
			return RawBlockUnknown, fmt.Errorf("could not find synchronization signal: %v", err)
		} else if bit == 255 {
			continue
		}

		if bit == currentBit {
			bitCount++
		} else {
			reader.RewindBit(bit)
			// fmt.Printf("sync: read %d x %d\n", currentBit, bitCount)
			switch state {
			case 0: /* syncing */
				if currentBit == 0 && bitCount >= reader.encInfo.SyncMinPulseCount {
					state = 1
				}
			case 1: /* 1 */
				firstBitCount = bitCount
				state = 2
			case 2: /* 0 */
				secondBitCount = bitCount
				if firstBitCount != secondBitCount {
					return RawBlockUnknown, fmt.Errorf("bit count mismatch on block type (%d != %d)", firstBitCount, secondBitCount)
				} else if firstBitCount == 40 {
					return RawBlockInfo, nil
				} else if firstBitCount == 20 {
					return RawBlockData, nil
				} else {
					return RawBlockUnknown, fmt.Errorf("could not recognize block type (%d)", firstBitCount)
				}
			}
			bitCount = 0
			currentBit = bit
		}
	}
}

type TapeWriter struct {
	writer      io.WriteSeeker
	wav         *wav.Encoder
	encInfo     TapeEncodingInfo
	freqResidue float64
}

func NewTapeWriter(writer io.WriteSeeker, encInfo TapeEncodingInfo, frequency int) (*TapeWriter, error) {
	tapeWriter := TapeWriter{
		writer:  writer,
		encInfo: encInfo,
	}

	wav := wav.NewEncoder(writer, frequency, 8, 1, 0x1)
	tapeWriter.wav = wav

	return &tapeWriter, nil
}

func (writer *TapeWriter) WriteSilence(length float64) error {
	samples := int(length * float64(writer.wav.SampleRate))
	for i := 0; i < samples; i++ {
		writer.wav.AddLE(128)
	}
	return nil
}

func (writer *TapeWriter) WritePulse(length int) error {
	samplesF := writer.freqResidue + (float64(length) * float64(writer.wav.SampleRate) / writer.encInfo.TapeFrequency())
	samples := int(samplesF)
	writer.freqResidue = samplesF - float64(samples)

	for i := 0; i < samples; i++ {
		writer.wav.AddLE(96)
	}
	for i := 0; i < samples; i++ {
		writer.wav.AddLE(160)
	}
	return nil
}

func (writer *TapeWriter) WriteBit(bit byte) error {
	if bit == 0 {
		return writer.WritePulse(writer.encInfo.ShortPulseWidth)
	} else if bit == 1 {
		return writer.WritePulse(writer.encInfo.LongPulseWidth)
	} else {
		return fmt.Errorf("unknown bit: %d", bit)
	}
}
