package rest

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

// IsAlreadyExists checks if the error is from postgres and then checks the error code
// keep adding these function until it's worth combining
func IsAlreadyExists(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return true
	}

	return false
}

func IsNotFound(err error) bool {
	pqErr := pq.Error{}
	if errors.Is(err, &pqErr) {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "20000" {
			return true
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return true
	}

	return false
}
