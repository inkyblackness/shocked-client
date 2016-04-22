package viewmodel

// Node is the base interface for an element in the view model.
type Node interface {
	// Specialize implements the visitor pattern. The provided visitor will be called
	// with the appropriate method to properly downcast the current node.
	Specialize(visitor NodeVisitor)
}
