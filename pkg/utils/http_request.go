package utils

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/cylonchau/pantheon/pkg/api/config"
)

// SendRequest sends an HTTP request and returns the response
func SendRequest(method, url string, body []byte, auth config.Auth) (*http.Response, error) {
	var authHeader string
	if auth.BaseAuth != "" {
		authHeader = fmt.Sprintf("Basic %s", auth.BaseAuth)
	} else if auth.BearerToken != "" {
		authHeader = fmt.Sprintf("Bearer %s", auth.BearerToken)
	} else if auth.SSOToken != "" {
		authHeader = fmt.Sprintf("sso=%s", auth.SSOToken)
	}

	// Create a new HTTP request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		if auth.SSOToken != "" {
			req.Header.Set("Cookie", authHeader)
		} else {
			req.Header.Set("Authorization", authHeader)
		}
	}

	// Perform the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	return resp, nil
}
