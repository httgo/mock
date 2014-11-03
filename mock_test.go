package mock

import (
	"bytes"
	"github.com/nowk/assert"
	"io"
	"net/http"
	"testing"
)

func check(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

var mux *http.ServeMux
var handerFn = func(bodyStr string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(bodyStr))
	})
}

func init() {
	mux = http.NewServeMux()
	mux.Handle("/", handerFn("Hello World!"))
	mux.Handle("/foo", handerFn("foo"))
	mux.Handle("/bar", handerFn("bar"))
	mux.Handle("/baz", handerFn("baz"))
}

const host = "api.example.com"

func readBody(t *testing.T, r io.ReadCloser) *bytes.Buffer {
	defer r.Close()

	var b []byte
	buf := bytes.NewBuffer(b)
	_, err := buf.ReadFrom(r)
	check(t, err)

	return buf
}

func TestWithoutHostMocksAllHosts(t *testing.T) {
	mock := Mock{
		Testing: t,
		Mux:     mux,
	}
	mock.Start()
	defer mock.Done()

	for _, v := range []string{
		"http://api.example.com",
		"http://example.com",
		"http://blog.example.com",
	} {
		req, err := http.NewRequest("GET", v, nil)
		check(t, err)

		resp, err := mock.Do(req)
		check(t, err)

		buf := readBody(t, resp.Body)
		assert.Equal(t, buf.String(), "Hello World!")
	}
}

func TestMockOnlyDefinedHost(t *testing.T) {
	ft := &tTesting{}
	mock := Mock{
		Testing: ft,
		Host:    "google.com",
		Mux:     mux,
	}
	mock.Start()
	defer mock.Done()

	req, err := http.NewRequest("GET", "http://api.example.com", nil)
	check(t, err)

	mock.Do(req)
	assert.Equal(t, ft.ErrorMsg,
		"mock error: called to unmocked URL: [GET] http://api.example.com")

	req, err = http.NewRequest("GET", "http://google.com", nil)
	check(t, err)
	resp, err := mock.Do(req)
	check(t, err)
	buf := readBody(t, resp.Body)
	assert.Equal(t, buf.String(), "Hello World!")
}

func TestURLSwapDoesNotAlterTheOriginalRequest(t *testing.T) {
	mock := Mock{
		Testing: t,
		Mux:     mux,
	}
	mock.Start()
	defer mock.Done()

	req, err := http.NewRequest("GET", "http://api.example.com", nil)
	check(t, err)

	_, err = mock.Do(req)
	check(t, err)

	assert.Equal(t, req.URL.String(), "http://api.example.com")
}
