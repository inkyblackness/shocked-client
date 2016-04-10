package editor

// GraphicsTexture describes a texture in graphics memory.
type GraphicsTexture interface {
	// Handle returns the texture handle.
	Handle() uint32
}
