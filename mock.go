package mock

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
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

	_client  *http.Client
	_history map[string]map[string][]*http.Request
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

// writeHistory logs the requests on mock by Method : URLString : []Request
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

// sandbox takes a request and modifies it to mock and restores to it's original
// state
func (m Mock) sandbox(req *http.Request, fn sandboxFunc) (*http.Response, error) {
	ucopy, uorig, err := m.tsURLize(req)
	if err != nil {
		m.Testing.Fatal(err)
	}

	req.URL = ucopy
	r, err := copyBody(req.Body)
	if err != nil {
		m.Testing.Fatal(err)
	}

	if r != nil {
		req.Body = ioutil.NopCloser(r)
	}

	resp, err := fn.Do(req)

	req.URL = uorig
	if r != nil {
		r.Seek(0, 0) // rewind
	}

	return resp, err
}

// copyBody reads a ReadCloser and returns a bytes.Reader (which can be seeked)
func copyBody(c io.ReadCloser) (*bytes.Reader, error) {
	if c == nil {
		return nil, nil
	}
	defer c.Close()

	var b []byte
	buf := bytes.NewBuffer(b)
	_, err := io.Copy(buf, c)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(buf.Bytes())
	return r, nil
}

// Do is the interface to http.Client.Do
func (m *Mock) Do(req *http.Request) (*http.Response, error) {
	err := m.check(req)
	if err != nil {
		m.Testing.Error(err)
	}

	resp, err := m.sandbox(req, func(req *http.Request) (*http.Response, error) {
		return m.client().Do(req)
	})
	m.writeHistory(req)

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

// Done closes the test server and resets mock state
func (m *Mock) Done() {
	m._history = nil
	m.Ts.Close()
}

// History returns matching requests made on mock
// Accepts both a full urlString or a regexp for partial string matches
func (m Mock) History(method string, q interface{}) []*http.Request {
	meth := m._history[method]
	if meth == nil {
		return nil
	}

	return matchRoute(q, meth)
}

func matchRoute(q interface{}, r map[string][]*http.Request) []*http.Request {
	var reqs []*http.Request

	switch u := q.(type) {
	case *regexp.Regexp:
		for k, v := range r {
			if u.MatchString(k) {
				reqs = append(reqs, v...)
			}
		}
	case string:
		return r[u]
	}

	return reqs
}
