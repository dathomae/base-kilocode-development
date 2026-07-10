package handlers

import (
	"time"

	c "chronocone/backend/constants"
)

func parseTimestamp(s string) (time.Time, error) {
	if s == "now" {
		return time.Now().UTC(), nil
	}
	if s == "infinity" {
		return c.InfinityTime, nil
	}
	return time.Parse(time.RFC3339, s)
}
