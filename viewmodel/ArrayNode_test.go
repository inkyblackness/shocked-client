package viewmodel

import (
	check "gopkg.in/check.v1"
)

type ArrayNodeSuite struct {
}

var _ = check.Suite(&ArrayNodeSuite{})

func (suite *ArrayNodeSuite) TestSpecializeCallsArray(c *check.C) {
	node := NewArrayNode([]Node{})
	visitor := NewTestingNodeVisitor()

	node.Specialize(visitor)

	c.Check(visitor.arrayNodes, check.DeepEquals, []Node{node})
}

func (suite *ArrayNodeSuite) TestGetReturnsInitialValue(c *check.C) {
	initial := []Node{NewStringValueNode("abc")}
	c.Check(NewArrayNode(initial).Get(), check.DeepEquals, initial)
}

func (suite *ArrayNodeSuite) TestSetChangesCurrentValue(c *check.C) {
	node := NewArrayNode(nil)
	newEntry := NewStringValueNode("efg")

	node.Set([]Node{newEntry, newEntry})

	c.Check(node.Get(), check.DeepEquals, []Node{newEntry, newEntry})
}

func (suite *ArrayNodeSuite) TestSetCallsRegisteredSubscriberWithNewEntries(c *check.C) {
	node := NewArrayNode(nil)
	var capturedEntries []Node

	node.Subscribe(func(newEntries []Node) {
		capturedEntries = newEntries
	})

	newEntry := NewStringValueNode("efg")
	node.Set([]Node{newEntry})

	c.Check(capturedEntries, check.DeepEquals, []Node{newEntry})
}
