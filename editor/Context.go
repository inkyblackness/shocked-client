package editor

import (
	"github.com/inkyblackness/shocked-client/editor/model"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/graphics/controls"
	"github.com/inkyblackness/shocked-client/opengl"
)

// Context provides some global resources.
type Context interface {
	ModelAdapter() *model.Adapter
	OpenGl() opengl.OpenGl // alternate: NewRenderContext(*viewMatrix)
	ForGraphics() graphics.Context
	ControlFactory() controls.Factory
}
