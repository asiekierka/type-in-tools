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

package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/asiekierka/type-in-tools/fbastool/internal"
	"github.com/spf13/cobra"
)

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Convert binary file to WAV file",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		outFilename := ""
		if len(args) >= 2 {
			outFilename = args[1]
		} else {
			outFilename = args[0] + ".wav"
		}

		freq, err := cmd.PersistentFlags().GetInt("rate")
		if err != nil {
			panic(err)
		}

		argName, _ := cmd.PersistentFlags().GetString("name")

		ext := filepath.Ext(args[0])
		info := internal.FBFileInfo{}
		info.Reserved1 = 0
		if strings.HasSuffix(ext, "prg") {
			info.Type = internal.FileTypeBasic
			info.Length = 0
			info.LoadAddress = 0x6006
			info.ExecutionAddress = 0x2020
		} else if strings.HasSuffix(ext, "gfx") {
			info.Type = internal.FileTypeBgGraphics
			info.Length = 0x100
			info.LoadAddress = 0x700
			info.ExecutionAddress = 0x2000
		}

		tapeFileName := strings.TrimSuffix(filepath.Base(args[0]), ext)
		if len(argName) > 0 {
			tapeFileName = argName
		} else {
			if len(tapeFileName) <= 13 && info.Type == internal.FileTypeBgGraphics && !strings.HasSuffix(tapeFileName, " BG") {
				tapeFileName += " BG"
			}
		}
		info.SetName(tapeFileName)

		inpFile, err := os.Open(args[0])
		if err != nil {
			panic(err)
		}
		defer inpFile.Close()

		outFile, err := os.Create(outFilename)
		if err != nil {
			panic(err)
		}
		defer outFile.Close()

		tapeEncInfo := internal.NewTapeEncodingInfo()
		tapeWriter, err := internal.NewTapeWriter(outFile, tapeEncInfo, freq)
		if err != nil {
			panic(err)
		}
		defer tapeWriter.Close()

		if info.Length == 0 {
			inpFileStat, err := inpFile.Stat()
			if err != nil {
				panic(err)
			}

			info.Length = uint16(inpFileStat.Size())
		}

		fbFile := internal.FBFile{Info: info}
		tapeWriter.WriteSilence(0.25)

		buf := make([]byte, info.Length)
		for {
			n, err := inpFile.Read(buf)
			if n == 0 && err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			} else if n < int(info.Length) {
				panic(fmt.Errorf("read %d, expected %d", n, info.Length))
			}
			fbFile.Data = buf

			err = tapeWriter.WriteFile(fbFile)
			if err != nil {
				panic(err)
			}
		}

		tapeWriter.WriteSilence(0.25)
	},
}

func init() {
	rootCmd.AddCommand(recordCmd)
	recordCmd.PersistentFlags().IntP("rate", "r", 32000, "Audio frequency")
	recordCmd.PersistentFlags().String("name", "", "Output file name")
}
