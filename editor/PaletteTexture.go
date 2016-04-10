package editor

import (
	"github.com/inkyblackness/shocked-client/opengl"
)

// ColorsPerPalette defines how many colors are per palette. This value is 256 to cover byte-based bitmaps.
const ColorsPerPalette = 256

// BytesPerPaletteColor defines the bytes necessary for one palette entry.
const BytesPerPaletteColor = 4

// ColorProvider is a function to return the RGBA values for a certain palette index.
type ColorProvider func(index int) (byte, byte, byte, byte)

// PaletteTexture contains a palette stored as OpenGL texture.
type PaletteTexture struct {
	gl opengl.OpenGl

	colorProvider ColorProvider
	handle        uint32
}

// NewPaletteTexture creates a new PaletteTexture instance.
func NewPaletteTexture(gl opengl.OpenGl, colorProvider ColorProvider) *PaletteTexture {
	tex := &PaletteTexture{
		colorProvider: colorProvider,
		handle:        gl.GenTextures(1)[0]}

	var palette [ColorsPerPalette * BytesPerPaletteColor]byte

	tex.loadColors(&palette)
	gl.BindTexture(opengl.TEXTURE_2D, tex.handle)
	gl.TexImage2D(opengl.TEXTURE_2D, 0, opengl.RGBA, ColorsPerPalette, 1, 0, opengl.RGBA, opengl.UNSIGNED_BYTE, palette)
	gl.TexParameteri(opengl.TEXTURE_2D, opengl.TEXTURE_MAG_FILTER, opengl.NEAREST)
	gl.TexParameteri(opengl.TEXTURE_2D, opengl.TEXTURE_MIN_FILTER, opengl.NEAREST)
	gl.GenerateMipmap(opengl.TEXTURE_2D)
	gl.BindTexture(opengl.TEXTURE_2D, 0)

	return tex
}

// Handle returns the texture handle.
func (tex *PaletteTexture) Handle() uint32 {
	return tex.handle
}

func (tex *PaletteTexture) loadColors(palette *[ColorsPerPalette * BytesPerPaletteColor]byte) {
	for i := 0; i < ColorsPerPalette; i++ {
		r, g, b, a := tex.colorProvider(i)

		palette[i*BytesPerPaletteColor+0] = r
		palette[i*BytesPerPaletteColor+1] = g
		palette[i*BytesPerPaletteColor+2] = b
		palette[i*BytesPerPaletteColor+3] = a
	}
}