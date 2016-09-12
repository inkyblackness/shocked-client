package viewmodel

import (
	check "gopkg.in/check.v1"
)

type SectionNodeSuite struct {
}

var _ = check.Suite(&SectionNodeSuite{})

func (suite *SectionNodeSuite) TestSpecializeCallsContainer(c *check.C) {
	node := NewSectionNode("", someNodeList(), NewBoolValueNode("", false))
	visitor := NewTestingNodeVisitor()

	node.Specialize(visitor)

	c.Check(visitor.sectionNodes, check.DeepEquals, []Node{node})
}

func (suite *SectionNodeSuite) TestLabel(c *check.C) {
	c.Check(NewSectionNode("l1", someNodeList(), NewBoolValueNode("", false)).Label(), check.Equals, "l1")
	c.Check(NewSectionNode("l2", someNodeList(), NewBoolValueNode("", false)).Label(), check.Equals, "l2")
}

func (suite *SectionNodeSuite) TestAvailable(c *check.C) {
	available := NewBoolValueNode("", false)

	c.Check(NewSectionNode("l1", someNodeList(), available).Available(), check.Equals, available)
}

func (suite *SectionNodeSuite) TestGetReturnsInitialValue(c *check.C) {
	initial := someNodeList()
	c.Check(NewSectionNode("", initial, NewBoolValueNode("", false)).Get(), check.DeepEquals, initial)
}

func (suite *SectionNodeSuite) TestSetCallsRegisteredSubscriber(c *check.C) {
	initial := someNodeList()
	node := NewSectionNode("", initial, NewBoolValueNode("", false))
	called := false

	node.Subscribe(func() {
		called = true
	})

	node.Set([]Node{})

	c.Check(called, check.Equals, true)
}
