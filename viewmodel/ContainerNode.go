package viewmodel

// ContainerNode is a node holding other nodes by a string identifier.
type ContainerNode struct {
	nodes map[string]Node
}

// NewContainerNode returns a new instance of a ContainerNode.
func NewContainerNode(nodes map[string]Node) *ContainerNode {
	node := &ContainerNode{nodes: nodes}

	return node
}

// Specialize is the Node interface implementation.
func (node *ContainerNode) Specialize(visitor NodeVisitor) {
	visitor.Container(node)
}

// Get returns the contained nodes.
func (node *ContainerNode) Get() map[string]Node {
	return node.nodes
}
