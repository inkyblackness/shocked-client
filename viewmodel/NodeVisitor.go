package viewmodel

// NodeVisitor will be called with appropriate node instances while
// walking through a tree in the view model.
type NodeVisitor interface {
	// StringValue will be called for any StringValueNode.
	StringValue(node *StringValueNode)
	// Container will be called for any ContainerNode.
	Container(node *ContainerNode)
	// Array will be called for any ArrayNode.
	Array(node *ArrayNode)
}
