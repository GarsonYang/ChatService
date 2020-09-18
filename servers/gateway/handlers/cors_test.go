package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORS(t *testing.T) {
	cases := []struct {
		method string
		url    string
	}{
		{
			"POST",
			"v1/users",
		},
		{
			"GET",
			"v1/users/me",
		},
		{
			"GET",
			"v1/users/abc",
		},
		{
			"PATCH",
			"v1/users/abc",
		},
		{
			"PATCH",
			"v1/users/me",
		},
		{
			"POST",
			"v1/sessions",
		},
		{
			"DELETE",
			"v1/sessions/mine",
		},
	}

	for _, c := range cases {
		resp := httptest.NewRecorder()
		req, _ := http.NewRequest(c.method, c.url, nil)

		cors := &CORS{
			Handler: http.NewServeMux(),
		}
		cors.ServeHTTP(resp, req)

		expectedCORS := "*"
		CORS := resp.Header().Get("Access-Control-Allow-Origin")
		if len(CORS) == 0 {
			t.Errorf("No `Access-Control-Allow-Origin` header found in the response: must be `%s`", expectedCORS)
		} else if !strings.HasPrefix(CORS, expectedCORS) {
			t.Errorf("incorrect `Access-Control-Allow-Origin` header value: expected it to be `%s` but got `%s`", expectedCORS, CORS)
		}
	}
}
