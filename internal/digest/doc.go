// Package digest tracks per-path payload digests so that driftwatch can avoid
// sending duplicate webhook alerts when the same drift is detected repeatedly
// without any change in the underlying file content.
package digest
