package native

import (
	"io/ioutil"
	"net/http"
)

// RestTransport uses the http package from Go for its implementation
type RestTransport struct {
	serverBase string
	deferrer   chan<- func()
	client     *http.Client
}

// NewRestTransport returns a new instance of RestTransport.
func NewRestTransport(serverBase string, deferrer chan<- func()) *RestTransport {
	return &RestTransport{
		serverBase: serverBase,
		deferrer:   deferrer,
		client:     new(http.Client)}
}

// Get retrieves data from the given URL.
func (rest *RestTransport) Get(url string, onSuccess func(jsonString string), onFailure func()) {
	request, _ := http.NewRequest(http.MethodGet, rest.serverBase+url, nil)

	rest.handle(request, onSuccess, onFailure)
}

// Put stores data at the given URL.
func (rest *RestTransport) Put(url string, jsonString string, onSuccess func(jsonString string), onFailure func()) {

}

// Post requests to add new data at the given URL.
func (rest *RestTransport) Post(url string, jsonString string, onSucces func(jsonString string), onFailure func()) {
}

func (rest *RestTransport) handle(request *http.Request, onSuccess func(jsonString string), onFailure func()) {
	go func() {
		response, err := rest.client.Do(request)
		task := onFailure

		defer func() {
			rest.deferrer <- task
		}()
		if response != nil {
			defer response.Body.Close()
		}

		if err == nil && response.StatusCode == 200 {
			var bodyData []byte

			bodyData, err = ioutil.ReadAll(response.Body)
			if err == nil {
				task = func() { onSuccess(string(bodyData)) }
			}
		}
	}()
}
