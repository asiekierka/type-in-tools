# fbastool

Simple tool for working with Family BASIC V2/V3 binary/text/audio data. Developed without access to proprietary Nintendo/Hudson code, if that is important to you.
MIT-licensed.

## Usage

To build, just run `go build`. It should create an `fbastool` binary.

### Typing in Family BASIC magazine software

    $ nano NAME.txt # type-in code
    $ ./fbastool bgedit NAME.gfx # type-in BG graphics
    $ ./fbastool basic -e NAME.txt NAME.prg
    $ ./fbastool record NAME.prg # outputs NAME.prg.wav
    $ ./fbastool record NAME.gfx # outputs NAME.gfx.wav

## Useful Development Resources

* [Enri's Family Basic V2.1A Notes](http://www43.tok2.com/home/cmpslv/Famic/Fambas.htm) - doesn't include extended V3 tokens
