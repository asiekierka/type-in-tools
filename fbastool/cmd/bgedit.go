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
	"strconv"
	"strings"

	g "github.com/AllenDang/giu"
	"github.com/asiekierka/type-in-tools/fbastool/internal"
	"github.com/spf13/cobra"
)

var bgFilename = ""
var nametable = internal.NewFBNameTable()

type BgCell struct {
	widget *g.InputTextWidget
	ix, iy int
	value  string
}

func bgeditLoop() {
	columns := make([]*g.TableColumnWidget, internal.FBNameTableWidth+1)
	rows := make([]*g.TableRowWidget, internal.FBNameTableHeight)

	columns[0] = g.TableColumn("")
	for ix := 0; ix < internal.FBNameTableWidth; ix++ {
		columns[ix+1] = g.TableColumn(strconv.Itoa(ix))
	}

	for iy := 0; iy < internal.FBNameTableHeight; iy++ {
		fields := make([]g.Widget, internal.FBNameTableWidth+1)
		fields[0] = g.Label(strconv.Itoa(iy))

		for ix := 0; ix < internal.FBNameTableWidth; ix++ {
			bgCell := new(BgCell)
			bgCell.ix = ix
			bgCell.iy = iy
			bgCell.value = nametable.GetString(bgCell.ix, bgCell.iy)
			bgCell.widget = g.InputText(&bgCell.value).Size(100).Flags(0).OnChange(func() {
				// fmt.Printf("setting %d %d to %s\n", bgCell.ix, bgCell.iy, bgCell.value)
				nametable.PutString(bgCell.ix, bgCell.iy, strings.TrimSpace(bgCell.value))
				bgCell.value = nametable.GetString(bgCell.ix, bgCell.iy)
			})
			fields[ix+1] = bgCell.widget
		}

		rows[iy] = g.TableRow(fields...)
	}

	g.SingleWindow().Layout(
		g.Table().Columns(columns...).Rows(rows...).Flags(g.TableFlagsSizingStretchSame),
		g.Button("Save").OnClick(func() {
			fp, err := os.Create(bgFilename)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			} else {
				nametable.Write(fp)
				fp.Close()
			}
		}),
	)
}

var bgeditCmd = &cobra.Command{
	Use:   "bgedit",
	Short: "GUI-based BG-GRAPHICS editor",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bgFilename = args[0]
		if _, err := os.Stat(bgFilename); err == nil {
			fp, err := os.Open(bgFilename)
			if err != nil {
				panic(err)
			}
			nametable.Read(fp)
			fp.Close()
		}

		wnd := g.NewMasterWindow("BG-GRAPHICS editor", 1120, 600, 0)
		wnd.Run(bgeditLoop)
	},
}

func init() {
	rootCmd.AddCommand(bgeditCmd)
}
