package mock

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

type Mock struct {
	Testing tester
	Scheme  string
	Host    string
	Ts      *httptest.Server

	_client  *http.Client
	_history map[string]map[string][]*http.Request
}

func (m *Mock) TestingT(t *testing.T) {
	m.Testing = t
}

// UseClient allows you to define an http.Client to use in the mock
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

// eligible checks the scheme and host eligibility
func (m Mock) eligible(req *http.Request) error {
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

// Do implements the http.Client.Do
func (m *Mock) Do(req *http.Request) (*http.Response, error) {
	err := m.eligible(req)
	if err != nil {
		m.Testing.Error(err)
	}

	tr := transaction{
		Mock: m,
	}
	resp, err := tr.Do(req)
	tr.Rollback()

	m.writeHistory(req)
	return resp, err
}

// Start starts the test server
func (m *Mock) Start() *httptest.Server {
	m.Ts.Start()
	return m.Ts
}

// StartTLS starts the test server in TLS
func (m *Mock) StartTLS() *httptest.Server {
	m.Ts.StartTLS()
	return m.Ts
}

// Done closes the test server and resets mock state
func (m *Mock) Done() {
	m._history = nil
	m.Ts.Close()
}

// writeHistory logs the requests made on mock
func (m *Mock) writeHistory(req *http.Request) {
	if m._history == nil {
		m._history = make(map[string]map[string][]*http.Request)
	}
	meth, u := req.Method, req.URL

	h := m._history[meth]
	if h == nil {
		h = make(map[string][]*http.Request)
	}

	s := u.String()
	h[s] = append(h[s], req)
	m._history[meth] = h
}

// History returns matching requests made on mock
// Accepts both a full urlString or a regexp for partial string matches
func (m Mock) History(method string, q interface{}) []*http.Request {
	reqs := m._history[method]
	if reqs == nil {
		return nil
	}

	return matchRoute(reqs, q)
}

func matchRoute(reqs map[string][]*http.Request, q interface{}) []*http.Request {
	var r []*http.Request

	switch u := q.(type) {
	case *regexp.Regexp:
		for k, v := range reqs {
			if u.MatchString(k) {
				r = append(r, v...)
			}
		}
	case string:
		return reqs[u]
	}

	return r
}
