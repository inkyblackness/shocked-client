package ui

import (
	check "gopkg.in/check.v1"
)

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
