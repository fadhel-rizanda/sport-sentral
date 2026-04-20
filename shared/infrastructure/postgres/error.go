package postgres

import (
	"errors"
	"gorm.io/gorm"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// PostgreSQL error codes
// https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	ErrCodeUniqueViolation     = "23505"
	ErrCodeForeignKeyViolation = "23503"
	ErrCodeNotNullViolation    = "23502"
	ErrCodeCheckViolation      = "23514"

	ErrCodeInvalidTextRepresentation = "22P02"

	ErrCodeTooManyConnections = "53300"

	ErrCodeQueryCanceled = "57014"

	ErrCodeDeadlockDetected     = "40P01"
	ErrCodeSerializationFailure = "40001"
)

// ---- Low-level helpers ----

// sqlState extracts the PostgreSQL error code from an error, if available.
// Handles wrapped errors via errors.As.
func sqlState(err error) (string, bool) {
	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		return pe.Code, true
	}
	return "", false
}

func hasCode(err error, code string) bool {
	if err == nil {
		return false
	}
	state, ok := sqlState(err)
	return ok && state == code
}

// ---- Integrity Constraint Violations ----

func IsUniqueViolation(err error) bool {
	return hasCode(err, ErrCodeUniqueViolation)
}

func IsForeignKeyViolation(err error) bool {
	return hasCode(err, ErrCodeForeignKeyViolation)
}

func IsNotNullViolation(err error) bool {
	return hasCode(err, ErrCodeNotNullViolation)
}

func IsCheckViolation(err error) bool {
	return hasCode(err, ErrCodeCheckViolation)
}

// ---- Data Exceptions ----

func IsInvalidTextRepresentation(err error) bool {
	return hasCode(err, ErrCodeInvalidTextRepresentation)
}

// ---- Concurrency ----

func IsDeadlock(err error) bool {
	return hasCode(err, ErrCodeDeadlockDetected)
}

func IsSerializationFailure(err error) bool {
	return hasCode(err, ErrCodeSerializationFailure)
}

// IsRetryable reports whether err is safe to retry.
// Use this in retry middleware for transactional operations.
//
//	for retries := 0; retries < maxRetries; retries++ {
//	    err = doTx(ctx)
//	    if err == nil || !postgres.IsRetryable(err) {
//	        break
//	    }
//	}
func IsRetryable(err error) bool {
	return IsDeadlock(err) || IsSerializationFailure(err)
}

// ---- Operator Intervention ----

// IsQueryCanceled reports whether err is a query cancellation (57014),
// e.g. caused by a context timeout or statement_timeout.
func IsQueryCanceled(err error) bool {
	return hasCode(err, ErrCodeQueryCanceled)
}

// ---- Constraint name helpers ----

func ConstraintName(err error) string {
	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		return pe.ConstraintName
	}

	msg := err.Error()
	const marker = `constraint "`
	if idx := strings.Index(msg, marker); idx != -1 {
		rest := msg[idx+len(marker):]
		if end := strings.Index(rest, `"`); end != -1 {
			return rest[:end]
		}
	}
	return ""
}

// IsUniqueConstraint reports whether err is a unique violation on a specific
// constraint name. Combines IsUniqueViolation + ConstraintName for cleaner callsites.
//
//	if postgres.IsUniqueConstraint(err, "users_email_key") {
//	    return apperr.Conflict("email already taken")
//	}
func IsUniqueConstraint(err error, constraintName string) bool {
	return IsUniqueViolation(err) && ConstraintName(err) == constraintName
}

// IsForeignKeyConstraint reports whether err is a foreign key violation on a
// specific constraint name.
//
//	if postgres.IsForeignKeyConstraint(err, "user_roles_role_id_fkey") {
//	    return apperr.NotFound("role not found")
//	}
func IsForeignKeyConstraint(err error, constraintName string) bool {
	return IsForeignKeyViolation(err) && ConstraintName(err) == constraintName
}

// ---- ORM Level Helpers ----

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, gorm.ErrRecordNotFound)
}
