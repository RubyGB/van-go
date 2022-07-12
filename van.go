package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"os"
	"strings"
	"encoding/hex"
	"io/ioutil"

	palettefy "rubygb.com/van-go/cmd/palettefy"
	colors "rubygb.com/van-go/pkg/colors"
)

func isHelp(s string) bool {
	return s == "help" || s == "-h" || s == "--h" || s == "-help" || s == "--help"
}
func isVersion(s string) bool {
	return s == "version" || s == "-v" || s == "--v" || s == "-version" || s == "--version"
}

// Helper function. It's common to print a message then quit processing.
func quitMsg(s string) {
	fmt.Println(s)
	os.Exit(1)
}

const VERSION = "1.0.0"
const VAN_HELP = `Standard usage of van is 'van <command> [args]'. Implemented commands:

  palettefy    Restricts a source image to the colors in a given palette. See 'van palettefy -help'.`

const PALETTEFY_HELP = `Palettefy restricts a source image to the colors in a given palette. Usage:

van palettefy -s path/to/src/image.* -p {"<palette>" | @path/to/palette.*} [-o <name>] [-d <code>]

Explanation:
The options -s, -p are mandatory.
-s gives the path to the source image.
-p defines the palette to be used by the program. If followed by an argument in quotes "...", the argument is interpreted as a sequence of hex codes, prefixed by # and separated by commas and spaces. If a file path is provided, the file is parsed as above where newlines are also valid separators in addition to commas and spaces.
Examples:
  -p "#000000 #ffffff" will use a monotone palette.
  -p "#0f380f, #306230, #8bac0f, #9bbc0f" will use the original Gameboy palette.
  -p @somedir/mypalette.txt (where mypalette.txt contains a hex value #xxxxxx on each line) will use the colors defined in mypalette.txt.

-o specifies the output file name. Default is to affix '_p' to the name of the input file.

-d specifies dithering behaviour. Default is -d 0 which implements no dithering and uses nearest neighbors in CIELAB space. Other options:
  -d 1 employs Floyd-Steinberg dithering using nearest neighbors in CIELAB space. This can yield extremely good reconstructions of even photographic images using as few as 8 colors.`


func main() {
	// Command line parsing
	if len(os.Args) == 1 || isHelp(os.Args[1]) {
		quitMsg(VAN_HELP)
	} else if isVersion(os.Args[1]) {
		quitMsg(VERSION)
	} else {
		// We have invoked a command (e.g. palettefy).
		switch os.Args[1] {
		case "help":
			if len(os.Args) == 2 {
				quitMsg(VAN_HELP)
			} else {
				switch os.Args[2] {
				case "palettefy":
					quitMsg(PALETTEFY_HELP)
				default:
					quitMsg(VAN_HELP)
				}
			}
		case "palettefy":
			if len(os.Args) == 2 || isHelp(os.Args[2]) {
				quitMsg(PALETTEFY_HELP)
			} else {
				// Parse palettefy options (see PALETTEFY_HELP).
				var srcFlag, paletteFlag, outputFlag string
				var ditherFlag int

				palettefyFlagSet := flag.NewFlagSet("", flag.ExitOnError)

				palettefyFlagSet.StringVar(&srcFlag, "s", "", "Mandatory. Sets the path to the source image to palettefy.")
				palettefyFlagSet.StringVar(&paletteFlag, "p", "", "Mandatory. Defines the palette to restrict the source image to. See 'van palettefy help' for formatting details.")
				palettefyFlagSet.StringVar(&outputFlag, "o", "", "Sets the name of the final output file. Default is the source file affixed with '_p'.")
				palettefyFlagSet.IntVar(&ditherFlag, "d", 0, "Set the dithering mode (int). Default is 0 for no dithering.")

				palettefyFlagSet.Parse(os.Args[2:])
				
				if srcFlag == "" { quitMsg("Error: Mandatory -s flag not set.") }
				if paletteFlag == "" { quitMsg("Error: Mandatory -p flag not set.") }
				pfy(srcFlag, paletteFlag, outputFlag, ditherFlag)
			}
		default:
			quitMsg(VAN_HELP)
		}
	}
}


// The arguments to this function are the unparsed flags submitted by the user. The only guarantee at function entry is that src and palette are nonempty.
func pfy(src string, palette string, oname string, dither int) {
	// Verify that the source file exists
	if _, err := os.Stat(src); errors.Is(err, os.ErrNotExist) {
		quitMsg("Error: -s flag file does not exist.")
	}

	reader, rerr := os.Open(src)
	if rerr != nil { quitMsg(rerr.Error()) }
	defer reader.Close()

	img, _, ierr := image.Decode(reader)
	if ierr != nil { quitMsg(ierr.Error()) }

	// Parse palette argument
	var colorPalette color.Palette
	var parsePalette = func (p string) color.Palette {
		p = strings.ReplaceAll(p, ",", "?")
		p = strings.ReplaceAll(p, " ", "?")
		p = strings.ReplaceAll(p, "\n", "?") // line break system dependent
		p = strings.ReplaceAll(p, "\r", "?")
		p = strings.ReplaceAll(p, "\r\n", "?")

		psplit := strings.Split(p, "?")
		var hexcodes []string
		for _, h := range psplit {
			if h != "" {
				if len(h) != 7 {
					fmt.Println(h)
					quitMsg("Error: -p flag one or more hex values is improperly formatted.")
				}
				hexcodes = append(hexcodes, h[1:])
			}
		}

		var cpal color.Palette
		for _, h := range hexcodes {
			bytes, berr := hex.DecodeString(h)
			if berr != nil { panic(berr) } // shouldn't happen
			cpal = append(cpal, colors.RGB{R:bytes[0], G:bytes[1], B:bytes[2]})
		}

		return cpal
	}
	
	
	if palette[0] == '@' {
		// Load file into string and parse
		file, ferr := ioutil.ReadFile(palette[1:])
		if ferr != nil { quitMsg(ferr.Error()) }
		colorPalette = parsePalette(string(file))
	} else {
		colorPalette = parsePalette(palette)
	}

	dstImage := palettefy.Palettefy(img, colorPalette, dither)
	
	if oname == "" {
		oname = strings.Clone(src)
		oname, _, _ = strings.Cut(oname, ".")
		oname = oname + "_p.png"
	}

	writer, werr := os.Create(oname)
	if werr != nil { panic(werr) }
	
	if err := png.Encode(writer, dstImage); err != nil {
		writer.Close()
		panic(err)
	}
	
	if err := writer.Close(); err != nil {
		panic(err)
	}
}
