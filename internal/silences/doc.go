// Package silences manages time-bounded muting of drift alerts for
// specific watched paths. A silenced path will not trigger webhook
// notifications until its silence expires or is manually removed.
package silences
