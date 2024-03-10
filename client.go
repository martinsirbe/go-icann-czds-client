// Package czds provides a Go client for interacting with the ICANN Centralized Zone Data Service (CZDS).
// It simplifies the process of querying CZDS to obtain zone files and lists the Top-Level Domains (TLDs).
// The client supports JWT authentication with an in-memory store by default and implements a refetching
// mechanism for the token if it doesn't exist, is invalid, or has expired.
// Users can provide a custom JWT token store by implementing the TokenStore interface. The client
// facilitates querying CZDS for zone file data, mapping the data by domain name and the associated
// DNS records. Additionally, it allows for listing TLDs
// available on the account and their status, providing a comprehensive tool for interacting with CZDS.
package czds

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client represents a client for interacting with the ICANN Centralized Zone Data Service (CZDS).
type Client struct {
	httpClient     *http.Client
	czdsAPIBaseURL string
}

// NewClient initialises and returns a new Client instance for interacting with the ICANN
// Centralized Zone Data Service (CZDS).
// It accepts email and password for authentication, enabling the client to obtain domain information
// from TLD zone files, list TLDs to allow to check their approval status for the account.
// If no TokenStore is provided, a new in-memory token store will be used by default, allowing for
// automatic management of JWT tokens.
func NewClient(email, password string, opts ...ClientOption) *Client {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	var tokenStore TokenStore
	if options.tokenStore == nil {
		tokenStore = &InMemoryTokenStore{}
	}

	accountsAPIBaseURL := "https://account-api.icann.org/api"
	if options.accountsAPIBaseURL != "" {
		accountsAPIBaseURL = options.accountsAPIBaseURL
	}

	czdsAPIBaseURL := "https://czds-api.icann.org/czds"
	if options.czdsAPIBaseURL != "" {
		czdsAPIBaseURL = options.czdsAPIBaseURL
	}

	httpClient := &http.Client{
		Transport: &authTransport{
			httpClient:         http.DefaultClient,
			email:              email,
			password:           password,
			tokenStore:         tokenStore,
			accountsAPIBaseURL: accountsAPIBaseURL,
		},
	}

	return &Client{
		httpClient:     httpClient,
		czdsAPIBaseURL: czdsAPIBaseURL,
	}
}

// GetZoneFile fetches and parses a zone file for a given TLD from the ICANN CZDS API.
// It requires a context for operation cancellation, a JWT for authorization, and the TLD name.
// The function returns a map of domain names to their records.
// An error is returned if the operation fails at any stage, including request creation, HTTP
// communication, decompression, or file parsing. It handles gzip-compressed zone files and expects
// authorized access to the requested zone file.
func (c *Client) GetZoneFile(ctx context.Context, tld string) (map[string][]string, error) {
	endpoint := fmt.Sprintf(c.czdsAPIBaseURL+"/downloads/%s.zone", tld)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create get zone file request for %s TLD", tld)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get zone file request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected HTTP 200 response, got %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Type") == "application/x-gzip" {
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	domainMap := make(map[string][]string)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) > 1 {
			domain := parts[0]
			record := strings.Join(parts[1:], ",")
			domainMap[domain] = append(domainMap[domain], record)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan zone file: %w", err)
	}

	return domainMap, nil
}

func (c *Client) ListTLDs(ctx context.Context) ([]TLD, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.czdsAPIBaseURL+"/tlds", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create list TLDs request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list TLDs request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected HTTP 200 response, got %d", resp.StatusCode)
	}

	var tlds []TLD
	if err := json.NewDecoder(resp.Body).Decode(&tlds); err != nil {
		return nil, fmt.Errorf("failed to decode list TLDs response body: %w", err)
	}

	return tlds, nil
}
