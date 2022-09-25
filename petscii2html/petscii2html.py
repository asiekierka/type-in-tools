#!/usr/bin/env python3

# Copyright (c) 2022 Adrian Siekierka
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

from PIL import Image
import base64, html, io, struct, sys

glyph_images = []

petcat_glyph_list = [
	"@", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l",
	"m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	"[", "\\", "]", "^", "_", "{space}", "!", "\"", "#", "$",
	"%", "&", "'", "(", ")", "*", "+", ",", "-", ".", "/", "0", "1", "2",
	"3", "4", "5", "6", "7", "8", "9", ":", ";", "<", "=", ">", "?",
	"{SHIFT-*}", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L",
	"M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
	"{SHIFT-+}", "{CBM--}", "{SHIFT--}", "~", "{CBM-*}", None, "{CBM-K}",
	"{CBM-I}", "{CBM-T}", "{CBM-@}", "{CBM-G}", "{CBM-+}", "{CBM-M}",
	"{CBM-POUND}", "{SHIFT-POUND}", "{CBM-N}", "{CBM-Q}", "{CBM-D}", "{CBM-Z}", "{CBM-S}", "{CBM-P}", "{CBM-A}", "{CBM-E}", "{CBM-R}", "{CBM-W}", "{CBM-H}", "{CBM-J}", "{CBM-L}", "{CBM-Y}", "{CBM-U}", "{CBM-Q}", "{SHIFT-@}", "{CBM-F}", "{CBM-C}", "{CBM-X}", "{CBM-V}", "{CBM-B}",
	"{null}", "{CTRL-A}", "{CTRL-B}", "{stop}", "{CTRL-D}", "{wht}", "{CTRL-F}", "{CTRL-G}", "{dish}", "{ensh}", "{$0a}", "{CTRL-K}", "{\\f}", "{\\n}", "{swlc}", "{CTRL-O}", "{CTRL-P}", "{down}", "{rvon}", "{home}", "{del}", "{CTRL-U}", "{CTRL-V}", "{CTRL-W}", "{CTRL-X}", "{CTRL-Y}", "{CTRL-Z}", "{esc}", "{red}", "{rght}", "{grn}", "{blu}"
]
for i in range(0xA0, 0xC1):
	petcat_glyph_list += [None]
petcat_glyph_list += [
	"{orng}", "{$82}", "{$83}", "{$84}", "{f1}", "{f3}", "{f5}", "{f7}", "{f2}", "{f4}", "{f6}", "{f8}", "{stret}", "{swuc}", "{$8f}", "{blk}", "{up}", "{rvof}", "{clr}", "{inst}", "{brn}", "{lred}", "{gry1}", "{gry2}", "{lgrn}", "{lblu}", "{gry3}", "{pur}", "{left}", "{yel}", "{cn}"
]
assert len(petcat_glyph_list) == 0xE0

with open(sys.argv[1], "rb") as fp:
    for ig in range(0, 512):
        glyph = Image.new("1", (8, 8))
        for iy in range(0, 8):
            irow = struct.unpack("<B", fp.read(1))[0]
            for ix in range(0, 8):
                v = (irow >> (7 - ix)) & 1
                glyph.putpixel((ix, iy), v ^ 1)
        glyph = glyph.resize((16, 16), resample=Image.Resampling.NEAREST)
        glyph_images.append(glyph)
        glyph.save(f"test/{ig}.png")

print("""<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>petscii/petcat table</title>
<style type="text/css">
table, tr, td {
  border: 2px solid #d4d4d4;
  border-collapse: collapse;
}
td {
  width: 60px;
  padding: 4px 1px;
}
table {
  margin-top: 2em;
}
td {
  text-align: center;
}
img {
  border: 2px solid #2bb;
}
p {
  margin: 0; padding: 0;
  font-size: 12px;
  font-family: monospace;
}
</style>
</head>
<body>
<h1>petscii/petcat table</h1>
""")

glyphs_per_row = 16

for ig in range(0, 512, 256):
	print("<table>")
	for i in range(0, 240, glyphs_per_row):
		petcat_glyphs_row = petcat_glyph_list[i:(i+glyphs_per_row)]
		if not any(map(lambda x: x is not None, petcat_glyphs_row)):
			continue
		print("<tr>")
		for ix in range(0, glyphs_per_row):
			print("<td>")
			if petcat_glyphs_row[ix] is not None:
				print("<img src=\"data:image/png;base64,", end="")
				with io.BytesIO() as outf:
					img = glyph_images[ig + i + ix]
					img.save(outf, format="PNG")
					print(base64.b64encode(outf.getvalue()).decode('ascii'), end="")
				print("\"/>")
				print(f"<p>{html.escape(petcat_glyphs_row[ix])}</p>")
			print("</td>")
		print("</tr>")
	print("</table>")

print("""
</body>
</html>""")
