package mock

import (
	"net/http"
)

type sandboxFunc func(*http.Request) (*http.Response, error)

func (s sandboxFunc) Do(req *http.Request) (*http.Response, error) {
	return s(req)
}
