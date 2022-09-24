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
	"io"
	"os"

	"github.com/asiekierka/fbastool/internal"
	"github.com/spf13/cobra"
)

// basicCmd represents the basic command
var basicCmd = &cobra.Command{
	Use:   "basic",
	Short: "Convert BASIC files to/from text representation",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		encMode, err := cmd.PersistentFlags().GetBool("encode")
		if err != nil {
			panic(err)
		}
		outFile, err := cmd.PersistentFlags().GetString("output")
		if err != nil {
			panic(err)
		}
		if encMode {
			data, err := os.ReadFile(args[0])
			if err != nil {
				panic(err)
			}
			var fp io.Writer
			if outFile == "-" {
				fp = os.Stdout
			} else {
				file, err := os.Create(outFile)
				if err != nil {
					panic(err)
				}
				defer file.Close()
				fp = file
			}
			err = internal.FBBasicStringToBin(string(data), fp)
			if err != nil {
				panic(err)
			}
		} else {
			infp, err := os.Open(args[0])
			if err != nil {
				panic(err)
			}
			var fp io.Writer
			if outFile == "-" {
				fp = os.Stdout
			} else {
				file, err := os.Create(outFile)
				if err != nil {
					panic(err)
				}
				defer file.Close()
				fp = file
			}
			outstr, err := internal.FBBasicBinToString(infp)
			if err != nil {
				panic(err)
			}
			fp.Write([]byte(outstr))
		}
	},
}

func init() {
	rootCmd.AddCommand(basicCmd)
	basicCmd.PersistentFlags().StringP("output", "o", "-", "Output file")
	basicCmd.PersistentFlags().BoolP("encode", "e", false, "Encode to binary")
}
