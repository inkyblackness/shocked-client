package browser

import (
	"github.com/Archs/js/JSON"
	"github.com/gopherjs/jquery"
)

// RestTransport uses jQuery Ajax requests for its implementation
type RestTransport struct {
}

// NewRestTransport returns a new instance of RestTransport.
func NewRestTransport() *RestTransport {
	return &RestTransport{}
}

// Get retrieves data from the given URL.
func (rest *RestTransport) Get(url string, onSuccess func(jsonString string), onFailure func()) {
	ajaxopt := map[string]interface{}{
		"method":   "GET",
		"url":      url,
		"dataType": "json",
		"data":     nil,
		"jsonp":    false,
		"success": func(data interface{}) {
			jsonString := JSON.Stringify(data)
			onSuccess(jsonString)
		},
		"error": func(status interface{}) {
			onFailure()
		},
	}

	jquery.Ajax(ajaxopt)
}

// Put stores data at the given URL.
func (rest *RestTransport) Put(url string, jsonString string, onSuccess func(jsonString string), onFailure func()) {

}

// Post requests to add new data at the given URL.
func (rest *RestTransport) Post(url string, jsonString string, onSucces func(jsonString string), onFailure func()) {
}