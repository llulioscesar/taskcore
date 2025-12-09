package store

import "strings"

// IsUniqueViolation checks if the error is a PostgreSQL unique constraint violation (code 23505)
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "unique")
}
