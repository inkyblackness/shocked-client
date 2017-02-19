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

func (suite *AreaSuite) SetUpTest(c *check.C) {
	suite.builder = NewAreaBuilder()
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
