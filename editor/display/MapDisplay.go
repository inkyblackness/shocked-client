package display

import (
	mgl "github.com/go-gl/mathgl/mgl32"

	"github.com/inkyblackness/shocked-client/editor/camera"
	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/graphics"
	"github.com/inkyblackness/shocked-client/ui"
	"github.com/inkyblackness/shocked-client/ui/events"
)

// MapDisplay is a display for a level map
type MapDisplay struct {
	area *ui.Area

	camera        *camera.LimitedCamera
	viewMatrix    mgl.Mat4
	renderContext *graphics.RenderContext

	grid *GridRenderable

	moveCapture func(pixelX, pixelY float32)
}

// NewMapDisplay returns a new instance.
func NewMapDisplay(parent *ui.Area, renderContextFactory func(*mgl.Mat4) *graphics.RenderContext) *MapDisplay {
	tileBaseLength := float32(32)
	tilesPerMapSide := float32(64.0)
	tileBaseHalf := tileBaseLength / 2.0
	camLimit := tilesPerMapSide*tileBaseLength - tileBaseHalf
	zoomLevelMin := float32(-2)
	zoomLevelMax := float32(4)

	display := &MapDisplay{
		camera:      camera.NewLimited(zoomLevelMin, zoomLevelMax, -tileBaseHalf, camLimit),
		moveCapture: func(float32, float32) {}}

	{
		builder := ui.NewAreaBuilder()
		builder.SetParent(parent)
		builder.SetLeft(ui.NewOffsetAnchor(parent.Left(), 0))
		builder.SetTop(ui.NewOffsetAnchor(parent.Top(), 0))
		builder.SetRight(ui.NewOffsetAnchor(parent.Right(), 0))
		builder.SetBottom(ui.NewOffsetAnchor(parent.Bottom(), 0))
		builder.SetVisible(false)
		builder.OnRender(func(area *ui.Area) { display.render() })
		builder.OnEvent(events.MouseScrollEventType, display.onMouseScroll)
		builder.OnEvent(events.MouseMoveEventType, display.onMouseMove)
		builder.OnEvent(events.MouseButtonDownEventType, display.onMouseButtonDown)
		builder.OnEvent(events.MouseButtonUpEventType, display.onMouseButtonUp)
		display.area = builder.Build()
	}

	display.renderContext = renderContextFactory(display.camera.ViewMatrix())
	display.grid = NewGridRenderable(display.renderContext)

	return display
}

// SetVisible sets the display visibility state.
func (display *MapDisplay) SetVisible(visible bool) {
	display.area.SetVisible(visible)
}

func (display *MapDisplay) render() {
	root := display.area.Root()
	display.camera.SetViewportSize(root.Right().Value(), root.Bottom().Value())
	display.grid.Render()
}

func (display *MapDisplay) unprojectPixel(pixelX, pixelY float32) (x, y float32) {
	pixelVec := mgl.Vec4{pixelX, pixelY, 0.0, 1.0}
	invertedView := display.camera.ViewMatrix().Inv()
	result := invertedView.Mul4x1(pixelVec)

	return result[0], result[1]
}

func (display *MapDisplay) onMouseScroll(area *ui.Area, event events.Event) bool {
	mouseEvent := event.(*events.MouseScrollEvent)
	mouseX, mouseY := mouseEvent.Position()
	worldX, worldY := display.unprojectPixel(mouseX, mouseY)
	_, dy := mouseEvent.Deltas()

	if dy > 0 {
		display.camera.ZoomAt(-0.5, worldX, worldY)
	}
	if dy < 0 {
		display.camera.ZoomAt(0.5, worldX, worldY)
	}

	return true
}

func (display *MapDisplay) onMouseMove(area *ui.Area, event events.Event) bool {
	mouseEvent := event.(*events.MouseMoveEvent)
	display.moveCapture(mouseEvent.Position())
	return true
}

func (display *MapDisplay) onMouseButtonDown(area *ui.Area, event events.Event) bool {
	mouseEvent := event.(*events.MouseButtonEvent)

	if mouseEvent.Buttons() == env.MousePrimary {
		lastPixelX, lastPixelY := mouseEvent.Position()

		display.area.RequestFocus()
		display.moveCapture = func(pixelX, pixelY float32) {
			lastWorldX, lastWorldY := display.unprojectPixel(lastPixelX, lastPixelY)
			worldX, worldY := display.unprojectPixel(pixelX, pixelY)

			display.camera.MoveBy(worldX-lastWorldX, worldY-lastWorldY)
			lastPixelX, lastPixelY = pixelX, pixelY
		}
	}

	return true
}

func (display *MapDisplay) onMouseButtonUp(area *ui.Area, event events.Event) bool {
	mouseEvent := event.(*events.MouseButtonEvent)

	if mouseEvent.AffectedButtons() == env.MousePrimary {
		if display.area.HasFocus() {
			display.area.ReleaseFocus()
		}
		display.moveCapture = func(float32, float32) {}
	}

	return true
}
