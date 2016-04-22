package viewmodel

type TestingNodeVisitor struct {
	containerNodes   []Node
	arrayNodes       []Node
	stringValueNodes []Node
}

func NewTestingNodeVisitor() *TestingNodeVisitor {
	return &TestingNodeVisitor{}
}

func (visitor *TestingNodeVisitor) StringValue(node *StringValueNode) {
	visitor.stringValueNodes = append(visitor.stringValueNodes, node)
}

func (visitor *TestingNodeVisitor) Container(node *ContainerNode) {
	visitor.containerNodes = append(visitor.containerNodes, node)
}

func (visitor *TestingNodeVisitor) Array(node *ArrayNode) {
	visitor.arrayNodes = append(visitor.arrayNodes, node)
}
