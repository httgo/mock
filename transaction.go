package mock

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"gopkg.in/httgo/interfaces.v2"
)

// transaction provides a struct to transform and restore requests for use in
// mock
type transaction struct {
	*http.Request
	*Mock

	body *bytes.Reader
	url  *url.URL
}

var _ interfaces.HTTPClient = &transaction{}

// Do calls Do on the mock client after the request has been prepared
func (t *transaction) Do(req *http.Request) (*http.Response, error) {
	t.Request = req
	t.prepare()

	resp, err := t.Mock.client().Do(req)
	return resp, err
}

// Rollback rolls the request back to an original statue (pre Do)
func (t *transaction) Rollback() {
	t.Request.URL = t.url

	if t.body != nil {
		t.body.Seek(0, 0) // rewind
	}
}

// prepare preps the request while saving copies to be restored
func (t *transaction) prepare() {
	m := t.Mock
	r := t.Request

	uc, uo, err := mockURL(m, r)
	if err != nil {
		m.Testing.Fatal(err)
	}
	t.url = uo
	r.URL = uc

	if r.Body != nil {
		b, err := copyBody(r.Body)
		if err != nil {
			m.Testing.Fatal(err)
		}

		t.body = b
		r.Body = ioutil.NopCloser(b)
	}
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

// mockURL copies the orignal url setting the host to the test server's. It
// returns both the mocked copy and the original url
func mockURL(m *Mock, req *http.Request) (*url.URL, *url.URL, error) {
	u, err := url.Parse(m.Ts.URL)
	if err != nil {
		return nil, req.URL, err
	}

	c := *req.URL
	c.Host = u.Host

	// default to http if no scheme is defined on the mock
	if m.Scheme == "" {
		c.Scheme = "http"
	}

	o := req.URL
	return &c, o, nil
}
