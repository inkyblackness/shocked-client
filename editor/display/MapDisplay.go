package display

import (
	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/graphics"
)

// MapDisplay is a display for a level map
type MapDisplay struct {
	viewMatrix    mgl.Mat4
	renderContext *graphics.RenderContext

	grid *GridRenderable
}

// NewMapDisplay returns a new instance.
func NewMapDisplay(renderContextFactory func(*mgl.Mat4) *graphics.RenderContext) *MapDisplay {
	display := &MapDisplay{
		viewMatrix: mgl.Ident4()}

	display.renderContext = renderContextFactory(&display.viewMatrix)
	display.grid = NewGridRenderable(display.renderContext)

	return display
}

// Render renders the map display
func (display *MapDisplay) Render() {
	display.grid.Render()
}
