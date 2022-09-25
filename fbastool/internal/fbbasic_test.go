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
	"bytes"
	"strings"
	"testing"
)

var enriExampleText = `
10 FOR I=0 TO 10
20 PRINT "TEST ";
30 NEXT
`
var enriExampleBin = []byte{
	0x11,
	0x0A, 0x00,
	0x8C, 0x20, 0x49, 0xF6, 0x12, 0x00, 0x00, 0x20, 0x88, 0x20, 0x12, 0x0A, 0x00,
	0x00,

	0x0E,
	0x14, 0x00,
	0x8B, 0x20, 0x22, 0x54, 0x45, 0x53, 0x54, 0x20, 0x22, 0x3B,
	0x00,

	0x05,
	0x1E, 0x00,
	0x8D,
	0x00,
}

func TestEnriString(t *testing.T) {
	enriGeneratedText, err := FBBasicBinToString(bytes.NewReader(enriExampleBin))
	if err != nil {
		t.Error(err)
	}
	if strings.TrimSpace(enriGeneratedText) != strings.TrimSpace(enriExampleText) {
		t.Errorf("mismatch\nexpected:\n%s\n\nactual:\n%s", enriExampleText, enriGeneratedText)
	}
}
