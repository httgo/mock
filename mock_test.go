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

func TestMockOnlyDefinedHost(t *testing.T) {
	ft := &tTesting{}
	mock := Mock{
		Testing: ft,
		Host:    "google.com",
		Ts:      httptest.NewUnstartedServer(mux),
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
		Ts:      httptest.NewUnstartedServer(mux),
	}
	mock.Start()
	defer mock.Done()

	req, err := http.NewRequest("GET", "http://api.example.com", nil)
	check(t, err)

	_, err = mock.Do(req)
	check(t, err)
	assert.Equal(t, req.URL.String(), "http://api.example.com")
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

func TestSchemeOnlyMocksForLikedScheme(t *testing.T) {
	ft := &tTesting{}
	mock := Mock{
		Testing: ft,
		Scheme:  "https",
		Ts:      httptest.NewUnstartedServer(mux),
	}
	mock.Start()
	defer mock.Done()

	req, err := http.NewRequest("GET", "http://api.example.com", nil)
	check(t, err)

	mock.Do(req)
	assert.Equal(t, ft.ErrorMsg,
		"mock error: called to unmocked URL: [GET] http://api.example.com")
}

func TestSchemeHTTPSIsRequiredForTLS(t *testing.T) {
	ts := httptest.NewUnstartedServer(mux)
	ts.TLS = &tls.Config{InsecureSkipVerify: true}

	mock := Mock{
		Testing: t,
		Scheme:  "https",
		Ts:      ts,
	}
	mock.Start()
	defer mock.Done()

	req, err := http.NewRequest("GET", "https://api.example.com", nil)
	check(t, err)

	resp, err := mock.Do(req)
	check(t, err)
	buf := readBody(t, resp.Body)
	assert.Equal(t, buf.String(), "Hello World!")
}
