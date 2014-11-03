package mock

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func check(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

// handlerFn helper to create handlers with a particular body string
var handerFn = func(bodyStr string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(bodyStr))
	})
}

var mux *http.ServeMux

// init some basic handles
func init() {
	mux = http.NewServeMux()
	mux.Handle("/", handerFn("Hello World!"))
	mux.Handle("/foo", handerFn("foo"))
	mux.Handle("/bar", handerFn("bar"))
	mux.Handle("/baz", handerFn("baz"))
}

// readBody helper reads from ReadCloser
func readBody(t *testing.T, r io.ReadCloser) *bytes.Buffer {
	defer r.Close()

	var b []byte
	buf := bytes.NewBuffer(b)
	_, err := buf.ReadFrom(r)
	check(t, err)

	return buf
}

// tTesting contains several of the testing package's actual methods
type tTesting struct {
	ErrorMsg string
	FatalMsg string
}

func (t *tTesting) Error(s ...interface{}) {
	t.ErrorMsg = fmt.Sprint(s...)
}

func (t *tTesting) Errorf(pat string, s ...interface{}) {
	t.ErrorMsg = fmt.Sprintf(pat, s...)
}

func (t *tTesting) Fatal(s ...interface{}) {
	t.FatalMsg = fmt.Sprint(s...)
}

func (t *tTesting) Fatalf(pat string, s ...interface{}) {
	t.FatalMsg = fmt.Sprintf(pat, s...)
}
