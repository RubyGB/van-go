package colors

import "image/color"

type RGB struct {
	R, G, B uint8
}
// Implements the color.Color interface
// Note that r,g,b are pre-multiplied by the alpha a
// All values are in [0, 0xffff]
func (c RGB) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	g = uint32(c.G)
	g |= g << 8
	b = uint32(c.B)
	b |= b << 8
	a = 0xffff
	return
}

// RGBModel implements color.Model interface function Convert
var RGBModel color.Model = color.ModelFunc(func(c color.Color) color.Color {
	r, g, b, _ := c.RGBA()
	return RGB{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
})


