package viewmodel

// StringValueListener is the callback type for changes in a StringValueNode.
type StringValueListener func(newValue string)

// StringValueNode is a node holding a simple string value.
type StringValueNode struct {
	label     string
	listeners []StringValueListener
	value     string
}

// NewStringValueNode returns a new instance of a StringValueNode.
func NewStringValueNode(label string, value string) *StringValueNode {
	node := &StringValueNode{
		label: label,
		value: value}

	return node
}

// Label is the Node interface implementation.
func (node *StringValueNode) Label() string {
	return node.label
}

// Specialize is the Node interface implementation.
func (node *StringValueNode) Specialize(visitor NodeVisitor) {
	visitor.StringValue(node)
}

// Subscribe registers the provided listener for value changes.
func (node *StringValueNode) Subscribe(listener StringValueListener) {
	node.listeners = append(node.listeners, listener)
}

// Get returns the current value.
func (node *StringValueNode) Get() string {
	return node.value
}

// Set requests to set a new value.
func (node *StringValueNode) Set(value string) {
	node.value = value
	for _, listener := range node.listeners {
		listener(value)
	}
}
