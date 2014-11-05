package mock

import (
	"crypto/tls"
	"github.com/nowk/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithoutHostMocksAllHosts(t *testing.T) {
	mock := Mock{
		Testing: t,
		Ts:      httptest.NewUnstartedServer(mux),
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

func TestDefiningHostThrowsErrorForNonMatchingURLs(t *testing.T) {
	ft := &tTesting{}
	mock := Mock{
		Testing: ft,
		Host:    "google.com",
		Ts:      httptest.NewUnstartedServer(mux),
	}
	mock.Start()
	defer mock.Done()

	for _, v := range []struct {
		u, err string
	}{
		{"http://api.example.com", "mock error: called to unmocked URL: [GET] http://api.example.com"},
		{"http://google.com", ""},
	} {
		req, err := http.NewRequest("GET", v.u, nil)
		check(t, err)

		mock.Do(req)
		assert.Equal(t, ft.ErrorMsg(), v.err)
	}
}

func TestURLSwapDoesNotAlterTheOriginalRequest(t *testing.T) {
	mock := Mock{
		Testing: t,
		Ts:      httptest.NewUnstartedServer(mux),
	}
	mock.Start()
	defer mock.Done()

	req, err := http.NewRequest("GET", "http://api.example.com/foo", nil)
	check(t, err)

	_, err = mock.Do(req)
	check(t, err)
	assert.Equal(t, req.URL.String(), "http://api.example.com/foo")
}

func TestWithoutSchemeMocksAllSchemes(t *testing.T) {
	mock := Mock{
		Testing: t,
		Ts:      httptest.NewUnstartedServer(mux),
	}
	mock.Start()
	defer mock.Done()

	for _, v := range []string{
		"http://api.example.com",
		"https://example.com",
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

func TestDefningSchemeThrowsErrorForNonMatchinScheme(t *testing.T) {
	ft := &tTesting{}
	mock := Mock{
		Testing: ft,
		Scheme:  "https",
		Ts:      httptest.NewUnstartedServer(mux),
	}
	mock.Start()
	defer mock.Done()

	for _, v := range []struct {
		u, err string
	}{
		{"http://api.example.com", "mock error: called to unmocked URL: [GET] http://api.example.com"},
		{"https://api.example.com", ""},
	} {
		req, err := http.NewRequest("GET", v.u, nil)
		check(t, err)

		mock.Do(req)
		assert.Equal(t, ft.ErrorMsg(), v.err)
	}
}

func TestHTTPSDefinesTLSConfigOnBothServerAndClient(t *testing.T) {
	ts := httptest.NewUnstartedServer(mux)
	ts.TLS = &tls.Config{InsecureSkipVerify: true}

	c := &http.Client{}
	c.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	mock := Mock{
		Testing: t,
		Scheme:  "https",
		Ts:      ts,
	}
	mock.SetClient(c)
	mock.Start()
	defer mock.Done()

	req, err := http.NewRequest("GET", "https://api.example.com", nil)
	check(t, err)

	resp, err := mock.Do(req)
	check(t, err)

	buf := readBody(t, resp.Body)
	assert.Equal(t, buf.String(), "Hello World!")
}
