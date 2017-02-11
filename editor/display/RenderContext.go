package display

import (
	"github.com/go-gl/mathgl/mgl32"
)

// RenderContext describes the current render properties
type RenderContext struct {
	viewportWidth, viewportHeight float32

	viewMatrix       mgl32.Mat4
	projectionMatrix mgl32.Mat4
}

// NewBasicRenderContext returns a render context for the provided parameters.
func NewBasicRenderContext(width, height float32, projectionMatrix mgl32.Mat4, viewMatrix mgl32.Mat4) *RenderContext {
	return &RenderContext{
		viewportWidth:    width,
		viewportHeight:   height,
		viewMatrix:       viewMatrix,
		projectionMatrix: projectionMatrix}
}

// ViewportSize returns the size of the current viewport, in pixel.
func (context *RenderContext) ViewportSize() (width, height float32) {
	return context.viewportWidth, context.viewportHeight
}

// ViewMatrix returns the current view matrix.
func (context *RenderContext) ViewMatrix() *mgl32.Mat4 {
	return &context.viewMatrix
}

// ProjectionMatrix returns the current projection matrix.
func (context *RenderContext) ProjectionMatrix() *mgl32.Mat4 {
	return &context.projectionMatrix
}
