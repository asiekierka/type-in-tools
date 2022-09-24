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
	"bufio"
	"bytes"
	"fmt"
	"os"

	"github.com/asiekierka/fbastool/internal"
	"github.com/spf13/cobra"
)

// testBasicCmd represents the testBasic command
var testBasicCmd = &cobra.Command{
	Use:   "testBasic",
	Short: "Round-trip convert binary -> text -> binary",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		origProgBin, err := os.ReadFile(args[0])
		if err != nil {
			panic(err)
		}
		fmt.Printf("binary -> text\n")
		progString, err := internal.FBBasicBinToString(bytes.NewReader(origProgBin))
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", progString)
		fmt.Printf("text -> binary\n")
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		err = internal.FBBasicStringToBin(progString, w)
		if err != nil {
			panic(err)
		}
		w.Flush()
		progBin := buf.Bytes()
		if !bytes.Equal(origProgBin, progBin) {
			fmt.Printf("Fail\n")
			os.WriteFile("test.orig", origProgBin, 0644)
			os.WriteFile("test.new", progBin, 0644)
		} else {
			fmt.Printf("Success\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(testBasicCmd)
}
