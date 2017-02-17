package ui

// AreaBuilder is used to create a new user-interface area.
type AreaBuilder struct {
	parent *Area

	left   Anchor
	top    Anchor
	right  Anchor
	bottom Anchor

	onRender RenderFunction
}

// NewAreaBuilder returns a new instance of a builder for creating areas.
func NewAreaBuilder() *AreaBuilder {
	builder := &AreaBuilder{
		left:     ZeroAnchor(),
		top:      ZeroAnchor(),
		right:    ZeroAnchor(),
		bottom:   ZeroAnchor(),
		onRender: func(*Area, Renderer) {}}

	return builder
}

// Build creates a new area instance based on the currently set properties.
func (builder *AreaBuilder) Build() *Area {
	area := &Area{
		parent: builder.parent,

		left:   builder.left,
		top:    builder.top,
		right:  builder.right,
		bottom: builder.bottom,

		onRender: builder.onRender}

	if area.parent != nil {
		area.parent.children = append(area.parent.children, area)
	}

	return area
}

// SetParent sets the parent area. By default, the builder has no parent set
// and the created area will be a root area.
func (builder *AreaBuilder) SetParent(parent *Area) *AreaBuilder {
	builder.parent = parent
	return builder
}

// SetLeft sets the left anchor. Default: ZeroAnchor
func (builder *AreaBuilder) SetLeft(value Anchor) *AreaBuilder {
	builder.left = value
	return builder
}

// SetTop sets the top anchor. Default: ZeroAnchor
func (builder *AreaBuilder) SetTop(value Anchor) *AreaBuilder {
	builder.top = value
	return builder
}

// SetRight sets the right anchor. Default: ZeroAnchor
func (builder *AreaBuilder) SetRight(value Anchor) *AreaBuilder {
	builder.right = value
	return builder
}

// SetBottom sets the bottom anchor. Default: ZeroAnchor
func (builder *AreaBuilder) SetBottom(value Anchor) *AreaBuilder {
	builder.bottom = value
	return builder
}

// OnRender sets the function for rendering the area.
// By default, an area has no own presentation.
func (builder *AreaBuilder) OnRender(render RenderFunction) *AreaBuilder {
	builder.onRender = render
	return builder
}
