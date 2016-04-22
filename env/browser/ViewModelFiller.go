package browser

import (
	"github.com/Archs/js/gopherjs-ko"
	"github.com/gopherjs/gopherjs/js"

	"github.com/inkyblackness/shocked-client/viewmodel"
)

type viewModelFiller struct {
	object *js.Object
}

func newViewModelFiller() *viewModelFiller {
	return &viewModelFiller{}
}

func (filler *viewModelFiller) StringValue(node *viewmodel.StringValueNode) {
	observable := ko.NewObservable(node.Get())

	filler.object = observable.ToJS()
	node.Subscribe(func(newValue string) {
		if observable.Get().String() != newValue {
			observable.Set(newValue)
		}
	})
	observable.Subscribe(func(jsValue *js.Object) {
		newValue := jsValue.String()

		if node.Get() != newValue {
			node.Set(newValue)
		}
	})
}

func (filler *viewModelFiller) Container(node *viewmodel.ContainerNode) {
	obj := js.Global.Get("Object").New()

	filler.object = obj
	for name, sub := range node.Get() {
		subFiller := newViewModelFiller()
		sub.Specialize(subFiller)
		obj.Set(name, subFiller.object)
	}
}

func (filler *viewModelFiller) Array(node *viewmodel.ArrayNode) {
	observable := ko.NewObservableArray()
	setEntries := func(nodeEntries []viewmodel.Node) {
		objEntries := make([]*js.Object, len(nodeEntries))

		for index, nodeEntry := range nodeEntries {
			subFiller := newViewModelFiller()
			nodeEntry.Specialize(subFiller)
			objEntries[index] = subFiller.object
		}
		observable.Set(objEntries)
	}

	filler.object = observable.ToJS()

	setEntries(node.Get())
	node.Subscribe(setEntries)
}
