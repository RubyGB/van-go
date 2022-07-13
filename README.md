# van-go
A library of image manipulation tools usable from a CLI interface, written in go.

## Installation
Ensure you have the latest version of Go installed. Clone this repo, navigate to it and run `go build -o van.exe` to compile an executable.

In a future version I'm likely to set up installation with a package manager, negating the requirement to have Go installed locally.

## palettefy
Takes as arguments a source image and a palette of colors, and creates a new image which closely matches the original while only using colors from the palette provided. Colors are compared in CIELAB space (under CIE76 metric) which is perceptually close to uniform. Includes dithering support.

Usage:
```
van palettefy -s path/to/src/image.* -p  {"<palette>" | @path/to/palette.*} [-o <name>] [-d <code>]
```
The options -s, -p are mandatory.

-s gives the path to the source image.

-p defines the palette to be used by the program. If followed by an argument in quotes "...", the argument is interpreted as a sequence of hex codes, prefixed by # and separated by commas and spaces. If a file path is provided, the file is parsed as above where newlines are also valid separators in addition to commas and spaces.
Examples:
```
-p "#000000 #ffffff" will use a monotone palette.
-p "#0f380f, #306230, #8bac0f, #9bbc0f" will use the original Gameboy palette.
-p @somedir/mypalette.txt (where mypalette.txt contains a hex value #xxxxxx on each line) will use the colors defined in mypalette.txt.
```
-o specifies the output file name. Default is to affix '_p' to the name of the input file.

-d specifies dithering behaviour. Default is `-d 0` which implements no dithering and uses nearest neighbors in CIELAB space. Other options:
- `-d 1` employs Floyd-Steinberg dithering using nearest neighbors in CIELAB space. This can yield extremely good reconstructions of even photographic images using as few as 8 colors.

Example (van Gogh's Starry Night, using only 8 colors with dithering):
![image](https://user-images.githubusercontent.com/108897249/178592147-f11c800b-ce2c-4c3e-a599-d7e8977620d9.png)
