package mock

import (
	"net/http"
	"net/http/httptest"
	"net/url"
)

// tester is an interface for the testing package
type tester interface {
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
}

type Mock struct {
	Testing tester
	Scheme  string
	Host    string
	Ts      *httptest.Server

	_client *http.Client
}

func (m *Mock) T(t tester) {
	m.Testing = t
}

// check checks the scheme and host eligibility
func (m Mock) check(req *http.Request) error {
	err := UnmockedError{
		Method: req.Method,
		URL:    req.URL.String(),
	}

	u := req.URL
	if m.Host != "" && m.Host != u.Host {
		return err
	}

	if m.Scheme != "" && m.Scheme != u.Scheme {
		return err
	}

	return nil
}

// tsURLize returns a new url set to the test server for this mock and a copy of
// the original url
func (m Mock) tsURLize(req *http.Request) (*url.URL, *url.URL, error) {
	tsurl, err := url.Parse(m.Ts.URL)
	if err != nil {
		return nil, req.URL, err
	}

	ucopy := *req.URL
	ucopy.Host = tsurl.Host

	// default to http is no scheme is defined on the mock
	if m.Scheme == "" {
		ucopy.Scheme = "http"
	}

	return &ucopy, req.URL, nil
}

// UseClient allows you to define an http.Client to use for the mock
// Primary use to set a client with a specific TLS configuration
func (m *Mock) UseClient(c *http.Client) {
	m._client = c
}

// client returns the defined client from UseClient() or defaults to
// http.DefaultClient
func (m Mock) client() *http.Client {
	if m._client == nil {
		m._client = http.DefaultClient
	}

	return m._client
}

// Do is the interface to http.DefaultClient.Do
func (m Mock) Do(req *http.Request) (*http.Response, error) {
	err := m.check(req)
	if err != nil {
		m.Testing.Error(err)
	}

	ucopy, uorig, err := m.tsURLize(req)
	if err != nil {
		m.Testing.Fatal(err)
	}
	req.URL = ucopy

	resp, err := m.client().Do(req)
	req.URL = uorig // restore
	return resp, err
}

// Start starts the httptest server
func (m *Mock) Start() *httptest.Server {
	m.Ts.Start()
	return m.Ts
}

// StartTLS starts the server TLS
func (m *Mock) StartTLS() *httptest.Server {
	m.Ts.StartTLS()
	return m.Ts
}

// Done closes the test server
func (m Mock) Done() {
	m.Ts.Close()
}
