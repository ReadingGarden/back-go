package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

type parityCase struct {
	name    string
	method  string
	path    string
	body    string
	headers map[string]string
}

func TestAuthParityAgainstDjango(t *testing.T) {
	djangoBaseURL := os.Getenv("AUTH_DJANGO_BASE_URL")
	goBaseURL := os.Getenv("AUTH_GO_BASE_URL")
	if djangoBaseURL == "" || goBaseURL == "" {
		t.Skip("set AUTH_DJANGO_BASE_URL and AUTH_GO_BASE_URL to run parity contract test")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	cases := []parityCase{
		{
			name:    "login missing required field",
			method:  http.MethodPost,
			path:    "/api/v1/auth/login",
			body:    `{"user_email":"user@example.com"}`,
			headers: map[string]string{"Content-Type": "application/json"},
		},
		{
			name:    "refresh missing required field",
			method:  http.MethodPost,
			path:    "/api/v1/auth/refresh",
			body:    `{}`,
			headers: map[string]string{"Content-Type": "application/json"},
		},
		{
			name:    "protected route missing authorization",
			method:  http.MethodGet,
			path:    "/api/v1/auth",
			headers: map[string]string{},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			djangoStatus, djangoBody := doRequest(t, client, djangoBaseURL+tc.path, tc.method, tc.body, tc.headers)
			goStatus, goBody := doRequest(t, client, goBaseURL+tc.path, tc.method, tc.body, tc.headers)

			if djangoStatus != goStatus {
				t.Fatalf("status mismatch: django=%d go=%d", djangoStatus, goStatus)
			}
			if !jsonEqual(djangoBody, goBody) {
				t.Fatalf("body mismatch:\ndjango=%s\ngo=%s", djangoBody, goBody)
			}
		})
	}
}

func doRequest(t *testing.T, client *http.Client, url, method, body string, headers map[string]string) (int, string) {
	t.Helper()

	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do() error = %v", err)
	}
	defer resp.Body.Close()

	var parsed interface{}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("json decode error = %v", err)
	}

	encoded, err := json.Marshal(parsed)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	return resp.StatusCode, string(encoded)
}

func jsonEqual(left, right string) bool {
	var leftValue interface{}
	var rightValue interface{}
	if err := json.Unmarshal([]byte(left), &leftValue); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(right), &rightValue); err != nil {
		return false
	}

	leftEncoded, _ := json.Marshal(leftValue)
	rightEncoded, _ := json.Marshal(rightValue)
	return string(leftEncoded) == string(rightEncoded)
}
