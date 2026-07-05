package handler

import (
	"encoding/json"
	"strings"
	"time"
)

// parseUpdatableDate reads a JSON date field for PATCH semantics so a set date
// can actually be cleared:
//
//	key absent         -> set=false (leave the value unchanged)
//	null or ""         -> set=true, t=nil  (clear the date)
//	"YYYY-MM-DD"        -> set=true, t=&parsed
//	RFC3339 timestamp   -> set=true, t=&parsed
//
// ok=false signals an unparseable value; the caller should return 400.
func parseUpdatableDate(raw json.RawMessage) (set bool, t *time.Time, ok bool) {
	if len(raw) == 0 {
		return false, nil, true
	}
	if string(raw) == "null" {
		return true, nil, true
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return false, nil, false
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return true, nil, true
	}
	if parsed, err := time.Parse(time.RFC3339, s); err == nil {
		return true, &parsed, true
	}
	if parsed, err := time.Parse("2006-01-02", s); err == nil {
		return true, &parsed, true
	}
	return false, nil, false
}
