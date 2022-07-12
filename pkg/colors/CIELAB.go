package colors

import (
	"image/color"
	"math"
)

type CIELAB struct {
	L, a, b float64
}
// Implements the color.Color interface
// All values are in [0, 0xffff]
func (c CIELAB) RGBA() (r, g, b, a uint32) {
	// Procedure is essentially the inverse of CIELABModel.Convert().
	// CIELAB -> CIEXYZ -> RGB.
	// TODO
	return 0, 0, 0, 0
}
// CIELABModel implements color.Model interface
var CIELABModel color.Model = color.ModelFunc(func(c color.Color) color.Color {
	// See if we are already in CIELAB format
	clab, cok := c.(CIELAB)
	if cok { return clab }

	// Convert to RGBA first then carry out RGB -> CIEXYZ -> CIELAB.
	crgba, _ := color.RGBAModel.Convert(c).(color.RGBA)

	var linearizeSRGB = func (c_srgb float64) float64 {
		if c_srgb <= 0.04045 { return c_srgb / 12.92 }
		return math.Pow((c_srgb + 0.055) / 1.055, 2.4)
	}

	R, G, B := float64(crgba.R) / 255.0, float64(crgba.G) / 255.0, float64(crgba.B) / 255.0
	R, G, B = linearizeSRGB(R), linearizeSRGB(G), linearizeSRGB(B)

	// This matrix calculates CIE XYZ coordinates w.r.t. D65 Illuminant
	X := 0.4124*R + 0.3576*G + 0.1805*B
	Y := 0.2126*R + 0.7152*G + 0.0722*B
	Z := 0.0193*R + 0.1192*G + 0.9505*B
	
	// Adjust so that pure white w.r.t Illuminant D65 is (1,1,1)
	X /= 0.9505
	Z /= 1.0890
	
	var fLab = func (t float64) float64 {
		const delta float64 = 6.0 / 29.0
		const delta3 float64 = delta * delta * delta

		if t > delta3 { return math.Cbrt(t) }
		return t / (3 * delta * delta) + (4.0 / 29.0)
	}

	return CIELAB{116*fLab(Y) - 16, 500*(fLab(X) - fLab(Y)), 200*(fLab(Y) - fLab(Z))}
})

// In certain places we want to enforce that CIELAB palettes are provided
type CIELABPalette []CIELAB

// We use the CIE76 formula for speed
func deltaESquared(c1, c2 CIELAB) float64 {
	dL, da, db := c2.L - c1.L, c2.a - c1.a, c2.b - c1.b
	return dL * dL + da * da + db * db
}
func deltaE(c1, c2 CIELAB) float64 {
	// deltaE ~= 2.3 corresponds to a JND
	return math.Sqrt(deltaESquared(c1, c2))
}

func (p CIELABPalette) ConvertCIE(c color.Color) CIELAB {
	if len(p) == 0 { return CIELAB{} }
	return p[p.IndexCIE(c)]
}
func (p CIELABPalette) IndexCIE(c color.Color) int {
	var i int
	c_lab := CIELABModel.Convert(c).(CIELAB)
	// Iterate colors in p and assign the closest in CIELAB space
	var curr float64
	least := math.MaxFloat64
	for k, pc := range p {
		curr = deltaESquared(c_lab, pc)
		if curr < least {
			least = curr
			i = k
		}
	}
	return i
}
