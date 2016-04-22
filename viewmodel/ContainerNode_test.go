package viewmodel

import (
	check "gopkg.in/check.v1"
)

type ContainerNodeSuite struct {
}

var _ = check.Suite(&ContainerNodeSuite{})

func (suite *ContainerNodeSuite) TestSpecializeCallsContainer(c *check.C) {
	node := NewContainerNode(map[string]Node{})
	visitor := NewTestingNodeVisitor()

	node.Specialize(visitor)

	c.Check(visitor.containerNodes, check.DeepEquals, []Node{node})
}

func (suite *ContainerNodeSuite) TestGetReturnsInitialValue(c *check.C) {
	initial := map[string]Node{"a": NewStringValueNode("abc"), "b": NewStringValueNode("def")}
	c.Check(NewContainerNode(initial).Get(), check.DeepEquals, initial)
}
