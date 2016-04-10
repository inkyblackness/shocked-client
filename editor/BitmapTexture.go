package editor

import (
	"github.com/inkyblackness/shocked-client/opengl"
)

// BitmapTexture contains a bitmap stored as OpenGL texture.
type BitmapTexture struct {
	gl opengl.OpenGl

	handle uint32
}

// NewBitmapTexture downloads the provided raw data to OpenGL and returns a BitmapTexture instance.
func NewBitmapTexture(gl opengl.OpenGl, width, height int, pixelData []byte) *BitmapTexture {
	tex := &BitmapTexture{
		handle: gl.GenTextures(1)[0]}

	gl.BindTexture(opengl.TEXTURE_2D, tex.handle)
	gl.TexImage2D(opengl.TEXTURE_2D, 0, opengl.ALPHA, int32(width), int32(height), 0, opengl.ALPHA, opengl.UNSIGNED_BYTE, pixelData)
	gl.TexParameteri(opengl.TEXTURE_2D, opengl.TEXTURE_MAG_FILTER, opengl.NEAREST)
	gl.TexParameteri(opengl.TEXTURE_2D, opengl.TEXTURE_MIN_FILTER, opengl.NEAREST)
	gl.GenerateMipmap(opengl.TEXTURE_2D)
	gl.BindTexture(opengl.TEXTURE_2D, 0)

	return tex
}

// Handle returns the texture handle.
func (tex *BitmapTexture) Handle() uint32 {
	return tex.handle
}
