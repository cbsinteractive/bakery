package tests

import "net/http"

// FakeClient is the client to be mocked
type FakeClient struct {
	DoFuncReturns func(req *http.Request) (*http.Response, error)
}

// Do is the fake client's Do function
func (f FakeClient) Do(req *http.Request) (*http.Response, error) {
	return f.DoFuncReturns(req)
}

//MockClient will return an instance of a client with a mocked response
func MockClient(do func(req *http.Request) (*http.Response, error)) FakeClient {
	return FakeClient{
		DoFuncReturns: do,
	}
}
