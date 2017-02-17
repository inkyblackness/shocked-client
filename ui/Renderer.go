package ui

import (
	"image/color"
)

// Renderer provides rendering primitives for rendering an area.
type Renderer interface {
	// FillRectangle fills a rectangular area with a color.
	FillRectangle(left, top, right, bottom float32, fillColor color.Color)
}
