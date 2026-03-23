package client

import (
	"fmt"
	"net/http"
)

func (httpClient *HttpClient) Get(url string) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request error: %v", err)
	}
	request.Header = httpClient.header
	return httpClient.Call(func() (*http.Response, error) {
		return httpClient.client.Do(request)
	})
}
