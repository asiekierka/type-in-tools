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
	"os"
	"path/filepath"
	"strconv"

	"github.com/asiekierka/fbastool/internal"
	"github.com/spf13/cobra"
)

// wavCmd represents the wav command
var wavCmd = &cobra.Command{
	Use:   "wav",
	Short: "Convert WAV files to/from binary data",
	Long: `Convert WAV files to/from binary data.
	
To decode, run "wav [-d] file.wav [output dir]".
To encode, run "wav -e file.wav [binary files/directories...]".`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		encMode, err := cmd.PersistentFlags().GetBool("encode")
		if err != nil {
			panic(err)
		}

		if encMode {
			panic(fmt.Errorf("TODO"))
		} else {
			if len(args) > 2 {
				panic(fmt.Errorf("specified %d args, expected at most 2", len(args)))
			}

			outPath := ""
			if len(args) >= 2 {
				outPath = args[1]
			} else {
				outPath = "."
			}

			wavToBin(args[0], outPath)
		}
	},
}

func wavToBin(filename string, outPath string) {
	fp, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	tapeEncInfo := internal.NewTapeEncodingInfo()
	tapeReader, err := internal.NewTapeReader(fp, tapeEncInfo)
	if err != nil {
		panic(err)
	}

	var files []*internal.FBFile
	filesByFilename := make(map[string][]*internal.FBFile)

	for {
		file, err := tapeReader.NextFile()
		if err != nil {
			fmt.Println(err)
			break
		}
		files = append(files, file)
		filename := file.Info.NameStr()
		filesByFilename[filename] = append(filesByFilename[filename], file)
	}

	fmt.Printf("found %d files\n", len(files))
	for filename, files := range filesByFilename {
		for i, file := range files {
			fmt.Printf("- %s (%v, %d bytes)\n", file.Info.NameStr(), file.Info.Type, file.Info.Length)

			suffix := ".bin"
			if file.Info.Type == internal.FileTypeBasic {
				suffix = ".prg"
			} else if file.Info.Type == internal.FileTypeBgGraphics {
				suffix = ".gfx"
			}
			if len(files) >= 2 {
				suffix = "_" + strconv.Itoa(i) + suffix
			}

			f, err := os.Create(filepath.Join(outPath, filename+suffix))
			if err != nil {
				panic(err)
			}
			f.Write(file.Data)
			f.Close()

			f, err = os.Create(filepath.Join(outPath, filename+suffix+".info"))
			if err != nil {
				panic(err)
			}
			infoBytes, err := file.Info.MarshalBinary()
			if err != nil {
				panic(err)
			}
			f.Write(infoBytes)
			f.Close()
		}
	}
}

func init() {
	rootCmd.AddCommand(wavCmd)
	wavCmd.PersistentFlags().BoolP("encode", "e", false, "Encoding mode")
}
