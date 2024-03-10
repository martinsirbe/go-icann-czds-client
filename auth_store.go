package czds

import (
	"context"
	"sync"
)

// InMemoryTokenStore implements TokenStore to provide an in-memory storage mechanism for JWT tokens.
type InMemoryTokenStore struct {
	jwt string
	mu  sync.Mutex
}

// Save stores the given JWT token in the in-memory store.
func (ts *InMemoryTokenStore) Save(_ context.Context, token string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.jwt = token
	return nil
}

// Get retrieves the stored JWT token from the in-memory store.
func (ts *InMemoryTokenStore) Get(_ context.Context) string {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.jwt
}
