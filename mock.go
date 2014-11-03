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
	Mux     *http.ServeMux
	Ts      *httptest.Server
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

	uorig := *req.URL
	ucopy := *req.URL
	ucopy.Host = tsurl.Host

	// default to http is no scheme is defined on the mock
	if m.Scheme == "" {
		ucopy.Scheme = "http"
	}

	return &ucopy, &uorig, nil
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

	resp, err := http.DefaultClient.Do(req)
	req.URL = uorig // restore
	return resp, err
}

// Start starts the httptest server
// This needs to be called to begin the mock
func (m *Mock) Start() *httptest.Server {
	m.Ts = httptest.NewServer(m.Mux)
	return m.Ts
}

// Done closes the test server
func (m Mock) Done() {
	m.Ts.Close()
}
