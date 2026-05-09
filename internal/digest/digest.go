// Package digest provides a registry for tracking per-path webhook payload
// digests, allowing driftwatch to suppress duplicate drift alerts when the
// file content has not changed since the last notification was sent.
package digest

import (
	"crypto/sha256"
	"fmt"
	"sync"
)

// Registry stores the last-sent digest for each watched path.
type Registry struct {
	mu      sync.Mutex
	digests map[string]string
}

// New returns an initialised Registry.
func New() *Registry {
	return &Registry{
		digests: make(map[string]string),
	}
}

// Hash returns a deterministic SHA-256 hex digest for the given payload bytes.
func Hash(payload []byte) string {
	sum := sha256.Sum256(payload)
	return fmt.Sprintf("%x", sum)
}

// IsDuplicate reports whether the digest for path matches the last recorded
// digest. If it does not match (or no digest exists), the registry is updated
// and false is returned so the caller knows to proceed with the alert.
func (r *Registry) IsDuplicate(path string, payload []byte) bool {
	h := Hash(payload)

	r.mu.Lock()
	defer r.mu.Unlock()

	if prev, ok := r.digests[path]; ok && prev == h {
		return true
	}

	r.digests[path] = h
	return false
}

// Reset removes the stored digest for path, forcing the next call to
// IsDuplicate to treat the payload as new regardless of content.
func (r *Registry) Reset(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.digests, path)
}

// Len returns the number of paths currently tracked.
func (r *Registry) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.digests)
}
