// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.

package httputil

type MockHttpClient struct {
	// overwrite these methods to get the desired output
	Getfunc    func(url string, headers map[string]string) (responseCode int, body []byte, err error)
	Postfunc   func(url string, headers map[string]string, payload []byte) (responseCode int, body []byte, err error)
	Putfunc    func(url string, headers map[string]string, payload []byte) (responseCode int, body []byte, err error)
	Deletefunc func(url string, headers map[string]string, payload []byte) (responseCode int, body []byte, err error)
}

func (client *MockHttpClient) Get(url string, headers map[string]string) (responseCode int, body []byte, err error) {
	return client.Getfunc(url, headers)
}

func (client *MockHttpClient) Post(url string, headers map[string]string, payload []byte) (responseCode int, body []byte, err error) {
	return client.Postfunc(url, headers, payload)
}
func (client *MockHttpClient) Put(url string, headers map[string]string, payload []byte) (responseCode int, body []byte, err error) {
	return client.Putfunc(url, headers, payload)
}
func (client *MockHttpClient) Delete(url string, headers map[string]string, payload []byte) (responseCode int, body []byte, err error) {
	return client.Deletefunc(url, headers, payload)
}
