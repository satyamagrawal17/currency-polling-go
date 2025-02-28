package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPollAPI_Success(t *testing.T) {
	// Mock API response
	mockResponse := OpenExchangeRatesResponse{
		Disclaimer: "Test Disclaimer",
		License:    "Test License",
		Timestamp:  time.Now().Unix(),
		Base:       "USD",
		Rates:      map[string]float64{"EUR": 0.85, "GBP": 0.75},
	}
	mockResponseBytes, _ := json.Marshal(mockResponse)

	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/latest.json?app_id=testAPIKey", r.URL.String())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(mockResponseBytes)
	}))
	defer server.Close()

	// Call PollAPI with the mock server URL
	result, err := PollAPI(context.Background(), server.URL+"/api/latest.json", "testAPIKey")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, mockResponse.Disclaimer, result.Disclaimer)
	assert.Equal(t, mockResponse.License, result.License)
	assert.Equal(t, mockResponse.Base, result.Base)
	assert.Equal(t, mockResponse.Rates, result.Rates)
}

func TestPollAPI_NonStatusOK(t *testing.T) {
	// Mock HTTP server returning non-OK status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Call PollAPI
	_, err := PollAPI(context.Background(), server.URL+"/api/latest.json", "testAPIKey")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API returned non-OK status")
}

func TestPollAPI_InvalidJSON(t *testing.T) {
	// Mock HTTP server returning invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	// Call PollAPI
	_, err := PollAPI(context.Background(), server.URL+"/api/latest.json", "testAPIKey")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JSON decoding failed")
}

func TestPollAPI_RequestError(t *testing.T) {
	// Mock HTTP server that closes connection immediately
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatalf("ResponseWriter does not implement http.Hijacker")
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Fatalf("Hijack failed: %v", err)
		}
		conn.Close()
	}))
	defer server.Close()

	// Call PollAPI
	_, err := PollAPI(context.Background(), server.URL+"/api/latest.json", "testAPIKey")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed")
}
func TestPollAPI_InvalidURL(t *testing.T) {
	_, err := PollAPI(context.Background(), ":invalidURL", "testAPIKey")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse URL")
}

func TestPollAPI_ContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "{}")
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := PollAPI(ctx, server.URL+"/api/latest.json", "testAPIKey")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed")
}
