package graphics

// TextureRenderer renders textures.
type TextureRenderer interface {
	// Render takes the portion defined by textureRect out of texture to
	// render it within the given display rectangle.
	// textureRect coordinates are given in fractions of the texture.
	Render(displayRect Rectangle, texture Texture, textureRect Rectangle)
}
