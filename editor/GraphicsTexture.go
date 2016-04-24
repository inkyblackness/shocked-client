package editor

// GraphicsTexture describes a texture in graphics memory.
type GraphicsTexture interface {
	// Dispose releases any internal resources.
	Dispose()
	// Handle returns the texture handle.
	Handle() uint32
}
