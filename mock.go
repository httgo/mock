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

func (m Mock) Do(req *http.Request) (*http.Response, error) {
	tsReq := *req // do not alter the actual request

	host := tsReq.URL.Host
	scheme := tsReq.URL.Scheme
	if (m.Host != "" && host != m.Host) || (m.Scheme != "" && scheme != m.Scheme) {
		m.Testing.Errorf("mock error: called to unmocked URL: [%s] %s",
			tsReq.Method,
			tsReq.URL.String())
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
