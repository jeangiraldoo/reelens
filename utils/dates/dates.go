package dateUtils

import "time"

var dateFormat = "Jan 2, 2006"

// Converts an RFC3339 timestamp into a short human-friendly date.
// Values that fail to parse are returned unchanged, so legacy or
// unexpected data degrades visibly instead of vanishing.
func Humanize(stored string) string {
	t, err := time.Parse(time.RFC3339, stored)
	if err != nil {
		return stored
	}
	return t.Format(dateFormat)
}
