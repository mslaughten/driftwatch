// Package throttle implements per-path cooldown logic for driftwatch
// alerts, preventing repeated webhook calls for the same file within
// a configurable time window.
package throttle
