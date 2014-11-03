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
	ts      *httptest.Server
}

func (m *Mock) T(t tester) {
	m.Testing = t
}

// swapHost returns a new url with the hsot swapped in with that of the test
// server
func swapHost(u *url.URL, ts *httptest.Server) (*url.URL, error) {
	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		return nil, err
	}

	newURL := *u
	newURL.Host = tsURL.Host
	return &newURL, nil
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

func (m Mock) Do(req *http.Request) (*http.Response, error) {
	tsReq := *req // do not alter the actual request

	err := m.check(req)
	if err != nil {
		m.Testing.Error(err)
	}

	u, err := swapHost(tsReq.URL, m.ts)
	if err != nil {
		m.Testing.Fatal(err)
	}
	tsReq.URL = u

	if m.Host == "" {
		tsReq.URL.Scheme = "http"
	}

	return http.DefaultClient.Do(&tsReq)
}

// Start starts the httptest server
// This needs to be called to begin the mock
func (m *Mock) Start() *httptest.Server {
	m.ts = httptest.NewServer(m.Mux)
	return m.ts
}

// Done closes the test server
func (m Mock) Done() {
	m.ts.Close()
}
