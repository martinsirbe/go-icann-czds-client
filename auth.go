package czds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

// TokenStore defines an interface for JWT storage solutions. It allows clients to implement their
// own mechanism for storing JWT tokens, facilitating a flexible approach to token management.
// The Save method is designed to be called by the client when the provided JWT token has expired
// or does not exist in the storage. This design abstracts the complexities of token persistence away
// from the client, allowing the client to focus solely on how and where to store the JWT.
// Through this interface, clients can tailor their storage strategy to fit their application's needs,
// whether that involves in-memory caching, database storage, or any other persistent storage solution.
// The goal is to minimise token refetching by efficiently managing token expiration and renewal,
// streamlining the authentication process.
type TokenStore interface {
	Save(token string) error
	Get() string
}

type authTransport struct {
	httpClient         *http.Client
	email              string
	password           string
	tokenStore         TokenStore
	accountsAPIBaseURL string
}

func (a *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token := a.tokenStore.Get()
	if !isTokenValid(token) {
		var err error
		token, err = a.fetchJWT()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch JWT: %w", err)
		}

		if !isTokenValid(token) {
			return nil, fmt.Errorf("fetched JWT is not valid: %w", err)
		}

		if err := a.tokenStore.Save(token); err != nil {
			return nil, fmt.Errorf("failed to store JWT: %w", err)
		}
	}

	req.Header.Add("Authorization", "Bearer "+token)

	return http.DefaultTransport.RoundTrip(req)
}

func (a *authTransport) fetchJWT() (string, error) {
	type credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	creds := &credentials{
		Username: a.email,
		Password: a.password,
	}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(&creds); err != nil {
		return "", fmt.Errorf("failed to encode credentials for auth request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.accountsAPIBaseURL+"/authenticate", body)
	if err != nil {
		return "", fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("authentication request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("expected HTTP 200 response, got %d", resp.StatusCode)
	}

	var auth authResponse
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return "", fmt.Errorf("failed to decode auth response body: %w", err)
	}

	return auth.AccessToken, nil
}

func isTokenValid(token string) bool {
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return false
	}

	var expiresAt time.Time
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			expiresAt = time.Unix(int64(exp), 0)
		}
	}

	return time.Now().UTC().Before(expiresAt.UTC())
}
