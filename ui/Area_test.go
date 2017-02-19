package ui

import (
	"github.com/inkyblackness/shocked-client/ui/events"

	check "gopkg.in/check.v1"
)

type testingEvent struct {
	eventType events.EventType
}

func (event *testingEvent) EventType() events.EventType {
	return event.eventType
}

type AreaSuite struct {
	builder *AreaBuilder
}

var _ = check.Suite(&AreaSuite{})

func (suite *AreaSuite) aPositionalEvent(x, y float32) *events.MouseEvent {
	event := events.InitMouseEvent(events.EventType("test.positional"), x, y, 0, 0)

	return &event
}

func (suite *AreaSuite) SetUpTest(c *check.C) {
	suite.builder = NewAreaBuilder()
	suite.builder.SetRight(NewAbsoluteAnchor(100.0))
	suite.builder.SetBottom(NewAbsoluteAnchor(100.0))
}

func (suite *AreaSuite) TestRenderCallsOtherAreas(c *check.C) {
	renderCounter := 0
	renderCalls := make(map[int]int)
	renderFunc := func(index int) func(*Area, Renderer) {
		return func(*Area, Renderer) {
			renderCalls[index] = renderCounter
			renderCounter++
		}
	}
	parent := suite.builder.Build()
	NewAreaBuilder().SetParent(parent).OnRender(renderFunc(0)).Build()
	NewAreaBuilder().SetParent(parent).OnRender(renderFunc(1)).Build()

	parent.Render(nil)

	c.Check(renderCalls, check.DeepEquals, map[int]int{0: 0, 1: 1})
}

func (suite *AreaSuite) TestHandleEventReturnsFalseForUnhandledEvent(c *check.C) {
	area := suite.builder.Build()
	event1 := &testingEvent{events.EventType("testEvent")}

	c.Check(area.HandleEvent(event1), check.Equals, false)
}

func (suite *AreaSuite) TestHandleEventCallsRegisteredHandler(c *check.C) {
	eventType := events.EventType("registeredEvent")
	called := false
	handler := func(area *Area, event events.Event) bool {
		called = true
		return false
	}
	suite.builder.OnEvent(eventType, handler)
	area := suite.builder.Build()
	event1 := &testingEvent{eventType}

	area.HandleEvent(event1)

	c.Check(called, check.Equals, true)
}

func (suite *AreaSuite) TestHandleEventReturnsResultFromRegisteredHandler_A(c *check.C) {
	eventType := events.EventType("registeredEvent")
	handler := func(area *Area, event events.Event) bool {
		return false
	}
	suite.builder.OnEvent(eventType, handler)
	area := suite.builder.Build()
	event1 := &testingEvent{eventType}

	c.Check(area.HandleEvent(event1), check.Equals, false)
}

func (suite *AreaSuite) TestHandleEventReturnsResultFromRegisteredHandler_B(c *check.C) {
	eventType := events.EventType("registeredEvent")
	handler := func(area *Area, event events.Event) bool {
		return true
	}
	suite.builder.OnEvent(eventType, handler)
	area := suite.builder.Build()
	event1 := &testingEvent{eventType}

	c.Check(area.HandleEvent(event1), check.Equals, true)
}

func (suite *AreaSuite) TestDispatchPositionalEventCallsHandleEventIfNoChildMatches(c *check.C) {
	testEvent := suite.aPositionalEvent(10.0, 10.0)
	called := false
	handler := func(area *Area, event events.Event) bool {
		called = true
		return false
	}
	suite.builder.OnEvent(testEvent.EventType(), handler)
	area := suite.builder.Build()

	area.DispatchPositionalEvent(testEvent)

	c.Check(called, check.Equals, true)
}

func (suite *AreaSuite) TestDispatchPositionalEventCallsChildrenAtPositionHighestFirst(c *check.C) {
	testEvent := suite.aPositionalEvent(50.0, 50.0)
	handleSequence := []int{}
	aHandler := func(index int) func(*Area, events.Event) bool {
		return func(*Area, events.Event) bool {
			handleSequence = append(handleSequence, index)
			return false
		}
	}

	suite.builder.OnEvent(testEvent.EventType(), aHandler(0))
	area := suite.builder.Build()

	{
		subAreaBuilder := NewAreaBuilder()
		subAreaBuilder.OnEvent(testEvent.EventType(), aHandler(1))
		subAreaBuilder.SetRight(area.Right())
		subAreaBuilder.SetBottom(area.Bottom())
		subAreaBuilder.SetParent(area)
		subAreaBuilder.Build()
	}
	{
		subAreaBuilder := NewAreaBuilder()
		subAreaBuilder.OnEvent(testEvent.EventType(), aHandler(2))
		subAreaBuilder.SetRight(area.Right())
		subAreaBuilder.SetBottom(area.Bottom())
		subAreaBuilder.SetParent(area)
		subAreaBuilder.Build()
	}
	{
		subAreaBuilder := NewAreaBuilder()
		subAreaBuilder.OnEvent(testEvent.EventType(), aHandler(3))
		subAreaBuilder.SetRight(NewAbsoluteAnchor(10.0))
		subAreaBuilder.SetBottom(NewAbsoluteAnchor(10.0))
		subAreaBuilder.SetParent(area)
		subAreaBuilder.Build()
	}

	area.DispatchPositionalEvent(testEvent)

	c.Check(handleSequence, check.DeepEquals, []int{2, 1, 0})
}