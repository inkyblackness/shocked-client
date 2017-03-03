package graphics

// Context is a provider of graphic utilities.
type Context interface {
	RectangleRenderer() *RectangleRenderer
	TextPainter() TextPainter
	Texturize(bmp *Bitmap) *BitmapTexture
	UITextRenderer() *BitmapTextureRenderer
}
