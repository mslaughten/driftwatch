// Package snapshot provides a persistent store for tracking the last known
// cryptographic hash of each watched configuration file.
//
// The store is backed by a JSON file on disk so that driftwatch can detect
// drift across process restarts without treating every file as changed on
// startup. Concurrent access is safe via an internal read/write mutex.
package snapshot
