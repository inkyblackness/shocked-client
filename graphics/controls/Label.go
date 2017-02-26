package controls

import (
	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/ui"
)

// BitmapTexturizer creates a bitmap texture from bitmap information.
type BitmapTexturizer func(*graphics.Bitmap) *graphics.BitmapTexture

// Label is a control for displaying a text within an area.
type Label struct {
	area *ui.Area

	textPainter     graphics.TextPainter
	texturizer      BitmapTexturizer
	textureRenderer *graphics.BitmapTextureRenderer

	scale             float32
	horizontalAligner Aligner
	verticalAligner   Aligner

	bitmap  graphics.TextBitmap
	texture *graphics.BitmapTexture
}

// SetText updates the current label text.
func (label *Label) SetText(text string) {
	if label.texture != nil {
		label.texture.Dispose()
		label.texture = nil
	}

	label.bitmap = label.textPainter.Paint(text)
	label.texture = label.texturizer(&label.bitmap.Bitmap)
}

func (label *Label) onRender(area *ui.Area) {
	u, v := label.texture.UV()
	textWidth, textHeight := label.texture.Size()
	areaLeft := area.Left().Value()
	areaWidth := area.Right().Value() - areaLeft
	areaTop := area.Top().Value()
	areaHeight := area.Bottom().Value() - areaTop
	modelMatrix := mgl.Ident4().Mul4(mgl.Translate3D(
		areaLeft+label.horizontalAligner(areaWidth, textWidth),
		areaTop+label.verticalAligner(areaHeight, textHeight), 0.0)).
		Mul4(mgl.Scale3D(textWidth*label.scale, textHeight*label.scale, 1.0))

	label.textureRenderer.Render(&modelMatrix, label.texture, graphics.RectByCoord(0.0, 0.0, u, v))
}
