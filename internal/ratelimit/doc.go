// Package ratelimit implements a token-bucket rate limiter used by
// driftwatch to throttle outbound webhook alerts and prevent
// notification storms when many watched files change simultaneously.
package ratelimit
