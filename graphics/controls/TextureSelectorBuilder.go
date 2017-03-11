package controls

import (
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/ui"
	"github.com/inkyblackness/shocked-client/ui/events"
)

// TextureSelectorBuilder creates new instances of a TextureSelector.
type TextureSelectorBuilder struct {
	areaBuilder *ui.AreaBuilder

	rectangleRenderer *graphics.RectangleRenderer
	textureRenderer   graphics.TextureRenderer

	provider TextureProvider

	selectionChangeHandler TextureSelectionChangeHandler
}

// NewTextureSelectorBuilder returns a new instance of a TextureSelectorBuilder.
func NewTextureSelectorBuilder(rectangleRenderer *graphics.RectangleRenderer, textureRenderer graphics.TextureRenderer) *TextureSelectorBuilder {
	builder := &TextureSelectorBuilder{
		areaBuilder:            ui.NewAreaBuilder(),
		rectangleRenderer:      rectangleRenderer,
		textureRenderer:        textureRenderer,
		provider:               func() []*graphics.BitmapTexture { return nil },
		selectionChangeHandler: func(int) {}}

	return builder
}

// Build creates a new TextureSelector instance from the current parameters.
func (builder *TextureSelectorBuilder) Build() *TextureSelector {
	selector := &TextureSelector{
		rectangleRenderer:      builder.rectangleRenderer,
		textureRenderer:        builder.textureRenderer,
		provider:               builder.provider,
		selectedIndex:          -1,
		selectionChangeHandler: builder.selectionChangeHandler}

	builder.areaBuilder.OnRender(selector.onRender)
	builder.areaBuilder.OnEvent(events.MouseScrollEventType, selector.onMouseScroll)
	builder.areaBuilder.OnEvent(events.MouseButtonClickedEventType, selector.onMouseButtonClicked)
	selector.area = builder.areaBuilder.Build()

	return selector
}

// SetParent sets the parent area.
func (builder *TextureSelectorBuilder) SetParent(parent *ui.Area) *TextureSelectorBuilder {
	builder.areaBuilder.SetParent(parent)
	return builder
}

// SetLeft sets the left anchor. Default: ZeroAnchor
func (builder *TextureSelectorBuilder) SetLeft(value ui.Anchor) *TextureSelectorBuilder {
	builder.areaBuilder.SetLeft(value)
	return builder
}

// SetTop sets the top anchor. Default: ZeroAnchor
func (builder *TextureSelectorBuilder) SetTop(value ui.Anchor) *TextureSelectorBuilder {
	builder.areaBuilder.SetTop(value)
	return builder
}

// SetRight sets the right anchor. Default: ZeroAnchor
func (builder *TextureSelectorBuilder) SetRight(value ui.Anchor) *TextureSelectorBuilder {
	builder.areaBuilder.SetRight(value)
	return builder
}

// SetBottom sets the bottom anchor. Default: ZeroAnchor
func (builder *TextureSelectorBuilder) SetBottom(value ui.Anchor) *TextureSelectorBuilder {
	builder.areaBuilder.SetBottom(value)
	return builder
}

// WithProvider sets the provider of textures
func (builder *TextureSelectorBuilder) WithProvider(provider TextureProvider) *TextureSelectorBuilder {
	builder.provider = provider
	return builder
}

// WithSelectionChangeHandler registers the callback for a changed selection.
func (builder *TextureSelectorBuilder) WithSelectionChangeHandler(handler TextureSelectionChangeHandler) *TextureSelectorBuilder {
	builder.selectionChangeHandler = handler
	return builder
}
