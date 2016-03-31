package browser

import (
	"bytes"
	"encoding/json"

	"github.com/Archs/js/JSON"
	"github.com/Archs/js/gopherjs-ko"
	"github.com/gopherjs/jquery"

	"github.com/inkyblackness/shocked-client/env"
)

func getResource(url string, responseData interface{}, onSuccess func(), onFailure func()) {
	ajaxopt := map[string]interface{}{
		"method":   "GET",
		"url":      url,
		"dataType": "json",
		"data":     nil,
		"jsonp":    false,
		"success": func(data interface{}) {
			dataString := JSON.Stringify(data)

			json.Unmarshal(bytes.NewBufferString(dataString).Bytes(), responseData)
			onSuccess()
		},
		"error": func(status interface{}) {
			onFailure()
		},
	}

	jquery.Ajax(ajaxopt)
}

type ViewModel struct {
	*ko.BaseViewModel

	MainSections        *ko.Observable `js:"mainSections"`
	SelectedMainSection *ko.Observable `js:"selectedMainSection"`
}

func NewViewModel() *ViewModel {
	viewModel := new(ViewModel)
	viewModel.BaseViewModel = ko.NewBaseViewModel()
	viewModel.MainSections = ko.NewObservableArray([]string{"main", "empty"})
	viewModel.SelectedMainSection = ko.NewObservable("main")

	return viewModel
}

func Run(app env.Application) {
	vm := NewViewModel()

	canvas := jquery.NewJQuery("canvas#output")
	window, _ := NewWebGlWindow(canvas.Get(0))

	app.Init(window)

	ko.ApplyBindings(vm)
}
