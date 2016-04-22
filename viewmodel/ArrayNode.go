package viewmodel

// ArrayListener is the callback type for changes in an ArrayNode.
type ArrayListener func(newEntries []Node)

// ArrayNode is a node holding an array of nodes.
type ArrayNode struct {
	listeners []ArrayListener
	entries   []Node
}

// NewArrayNode returns a new instance of an ArrayNode.
func NewArrayNode(entries []Node) *ArrayNode {
	node := &ArrayNode{entries: entries}

	return node
}

// Specialize is the Node interface implementation.
func (node *ArrayNode) Specialize(visitor NodeVisitor) {
	visitor.Array(node)
}

// Subscribe registers the provided listener for array changes.
func (node *ArrayNode) Subscribe(listener ArrayListener) {
	node.listeners = append(node.listeners, listener)
}

// Get returns the current entries.
func (node *ArrayNode) Get() []Node {
	return node.entries[:]
}

// Set changes the current entries
func (node *ArrayNode) Set(entries []Node) {
	newEntries := entries[:]

	node.entries = newEntries
	for _, listener := range node.listeners {
		listener(newEntries)
	}
}
