package palettefy

import (
	"image"
	"image/draw"
	"image/color"

	colors "rubygb.com/van-go/pkg/colors"
)

func Palettefy(srcImage image.Image, palette color.Palette, ditherCode int) image.Image {
	
	var paletteCIELAB colors.CIELABPalette
	for _, c := range palette {
		paletteCIELAB = append(paletteCIELAB, colors.CIELABModel.Convert(c).(colors.CIELAB))
	}

	bounds := srcImage.Bounds()
	var dstImage draw.Image = image.NewNRGBA(bounds)
	
	switch ditherCode {
	case 0: // no dithering
		var AtRGB = func (x, y int) colors.RGB { return colors.RGBModel.Convert(srcImage.At(x,y)).(colors.RGB) }
		var NearestRGB = func (x, y int) colors.RGB { return palette[ paletteCIELAB.IndexCIE(AtRGB(x,y)) ].(colors.RGB) }
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				dstImage.Set(x,y,NearestRGB(x,y))
			}
		}
	case 1: // CIELAB Floyd-Steinberg dithering
		dx, dy := bounds.Max.X - bounds.Min.X, bounds.Max.Y - bounds.Min.Y
		var RGBchan [][][3]int32 = make([][][3]int32, dx)
		for x := range RGBchan { RGBchan[x] = make([][3]int32, dy) }
		// populate channels with all their values first
		for y := 0; y < dy; y++ {
			for x := 0; x < dx; x++ {
				c := colors.RGBModel.Convert(srcImage.At(x, y)).(colors.RGB)
				RGBchan[x][y][0] = int32(c.R)
				RGBchan[x][y][1] = int32(c.G)
				RGBchan[x][y][2] = int32(c.B)
			}
		}
		
		var pushError = func (x,y int, delta [3]int32, fac float64) {
			for i := 0; i < 3; i++ { RGBchan[x][y][i] += int32(float64(delta[i]) * fac) }
		}
		var ClipRGB = func (x, y int) colors.RGB {
			var rgbu [3]uint8
			rgbs := RGBchan[x][y] // int32s, convert to uint8
			for i := 0; i < 3; i++ {
				if rgbs[i] < 0 {
					rgbu[i] = 0
				} else if rgbs[i] > 0xff {
					rgbu[i] = 0xff
				} else { rgbu[i] = uint8(rgbs[i]) }
			}
			return colors.RGB{R:rgbu[0], G:rgbu[1], B:rgbu[2]}
		}
		
		// Apply FS dithering channel-wise
		var cat, cnear colors.RGB
		for y := 0; y < dy; y++ {
			for x := 0; x < dx; x++ {
				cat = ClipRGB(x, y)
				cnear = palette[ paletteCIELAB.IndexCIE(cat) ].(colors.RGB)
				dstImage.Set(x + bounds.Min.X, y + bounds.Min.Y, cnear)

				// Spread quantization error over neighboring pixels
				qerr := [3]int32{int32(cat.R) - int32(cnear.R), int32(cat.G) - int32(cnear.G), int32(cat.B) - int32(cnear.B)}
				if x+1<dx {
					pushError(x+1, y, qerr, 7.0 / 16.0)
				}
				if y+1<dy {
					pushError(x, y+1, qerr, 5.0 / 16.0)
					if x>0 {
						pushError(x-1, y+1, qerr, 3.0 / 16.0)
					}
					if x+1<dx {
						pushError(x+1, y+1, qerr, 1.0 / 16.0)
					}
				}
			}
		}
	}
	
	return dstImage
}
