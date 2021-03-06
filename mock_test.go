package mock

import (
	"bytes"
	"crypto/tls"
	"github.com/nowk/assert"
	"net/http"
	"net/http/httptest"
	"regexp"
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
		{"http://api.example.com",
			"mock error: called to unmocked URL: [GET] http://api.example.com"},
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

	for _, v := range []string{
		"http://api.example.com",
		"http://api.example.com/foo",
		"http://api.example.com/bar",
	} {
		req, err := http.NewRequest("GET", v, nil)
		check(t, err)

		_, err = mock.Do(req)
		check(t, err)
		assert.Equal(t, req.URL.String(), v)
	}
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
		{"http://api.example.com",
			"mock error: called to unmocked URL: [GET] http://api.example.com"},
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
	mock.UseClient(c)
	mock.StartTLS()
	defer mock.Done()

	req, err := http.NewRequest("GET", "https://api.example.com", nil)
	check(t, err)

	resp, err := mock.Do(req)
	check(t, err)

	buf := readBody(t, resp.Body)
	assert.Equal(t, buf.String(), "Hello World!")
}

func TestCanRetrieveRequestHistoryByMethodAndURLString(t *testing.T) {
	mock := Mock{
		Testing: t,
		Ts:      httptest.NewUnstartedServer(mux),
	}
	mock.Start()
	defer mock.Done()

	req1, _ := http.NewRequest("POST", "https://api.example.com/foo", nil)
	req2, _ := http.NewRequest("POST", "https://api.example.com/foo", nil)
	req3, _ := http.NewRequest("POST", "https://api.example.com/foo", nil)
	req4, _ := http.NewRequest("GET", "https://api.example.com/foo", nil)

	mock.Do(req1)
	mock.Do(req1)
	mock.Do(req2)
	mock.Do(req3)
	mock.Do(req4)

	reqs := mock.History("POST", "https://api.example.com/foo")
	assert.Equal(t, 4, len(reqs))

	reqs = mock.History("GET", "https://api.example.com/foo")
	assert.Equal(t, 1, len(reqs))
}

func TestCanRetrieveRequestHistoryWithRegex(t *testing.T) {
	mock := Mock{
		Testing: t,
		Ts:      httptest.NewUnstartedServer(mux),
	}
	mock.Start()
	defer mock.Done()

	req1, _ := http.NewRequest("GET", "https://api.example.com", nil)
	req2, _ := http.NewRequest("GET", "https://api.example.com/foo", nil)
	req3, _ := http.NewRequest("GET", "https://api.example.com/bar", nil)
	req4, _ := http.NewRequest("GET", "https://api.example.com/baz", nil)

	mock.Do(req1)
	mock.Do(req2)
	mock.Do(req3)
	mock.Do(req4)

	for _, v := range []struct {
		r string
		n int
	}{
		{`api\.example\.com\/b\w+`, 2},
		{`api\.google\.com\/b\w+`, 0},
		{`api\.example\.com`, 4},
		{`api\.example\.com$`, 1},
	} {
		regp := regexp.MustCompile(v.r)
		reqs := mock.History("GET", regp)
		assert.Equal(t, v.n, len(reqs))
	}
}

func TestDoneResetsMockState(t *testing.T) {
	mock := Mock{
		Testing: t,
		Ts:      httptest.NewUnstartedServer(mux),
	}
	mock.Start()

	req, err := http.NewRequest("POST", "https://api.example.com/foo", nil)
	check(t, err)

	mock.Do(req)
	mock.Do(req)

	reqs := mock.History("POST", "https://api.example.com/foo")
	assert.Equal(t, 2, len(reqs))

	mock.Done()
	reqs = mock.History("POST", "https://api.example.com/foo")
	assert.Equal(t, 0, len(reqs))
}

func TestRequestBodyCanBeReadAtBothEnds(t *testing.T) {
	var b []byte
	buf := bytes.NewBuffer(b)

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		buf.ReadFrom(req.Body)
	})

	mux := http.NewServeMux()
	mux.Handle("/request/body", h)

	mock := Mock{
		Testing: t,
		Ts:      httptest.NewUnstartedServer(mux),
	}
	mock.Start()

	r := bytes.NewBufferString("Foo Bar")
	req, err := http.NewRequest("POST", "https://api.example.com/request/body", r)
	check(t, err)

	mock.Do(req)
	assert.Equal(t, "Foo Bar", buf.String())

	reqs := mock.History("POST", "https://api.example.com/request/body")
	body := readBody(t, reqs[0].Body)
	assert.Equal(t, "Foo Bar", body.String())
}
