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
var handlerFn = func(bodyStr string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(bodyStr))
	})
}

var mux *http.ServeMux

// init some basic handles
func init() {
	mux = http.NewServeMux()
	mux.Handle("/", handlerFn("Hello World!"))
	mux.Handle("/foo", handlerFn("foo"))
	mux.Handle("/bar", handlerFn("bar"))
	mux.Handle("/baz", handlerFn("baz"))
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
	errorMsg string
	fatalMsg string
}

func (t *tTesting) Error(s ...interface{}) {
	t.errorMsg = fmt.Sprint(s...)
}

func (t *tTesting) Errorf(pat string, s ...interface{}) {
	t.errorMsg = fmt.Sprintf(pat, s...)
}

func (t *tTesting) Fatal(s ...interface{}) {
	t.fatalMsg = fmt.Sprint(s...)
}

func (t *tTesting) Fatalf(pat string, s ...interface{}) {
	t.fatalMsg = fmt.Sprintf(pat, s...)
}

func (t *tTesting) ErrorMsg() string {
	defer func() {
		t.errorMsg = ""
	}()
	return t.errorMsg
}

func (t *tTesting) FatalMsg() string {
	defer func() {
		t.fatalMsg = ""
	}()
	return t.fatalMsg
}
