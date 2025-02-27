package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Reusable HTTP client
var httpClient = &http.Client{}

type OpenExchangeRatesResponse struct {
	Disclaimer string             `json:"disclaimer"`
	License    string             `json:"license"`
	Timestamp  int64              `json:"timestamp"`
	Base       string             `json:"base"`
	Rates      map[string]float64 `json:"rates"`
}

func PollAPI(ctx context.Context, apiURL string, apiKey string) (*OpenExchangeRatesResponse, error) {
	// Parse the URL and add the app_id query parameter
	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}
	query := parsedURL.Query()
	query.Set("app_id", apiKey)
	parsedURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
	}

	var data OpenExchangeRatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("JSON decoding failed: %v", err)
	}

	return &data, nil
}
