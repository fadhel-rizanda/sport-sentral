package utils

import (
	"github.com/google/uuid"
)

// UUIDPtrToStringPtr converts a uuid.UUID pointer to a string pointer.
func UUIDPtrToStringPtr(u *uuid.UUID) *string {
	if u == nil {
		return nil
	}
	str := u.String()
	return &str
}

// UUIDToStringPtr converts a uuid.UUID to a string pointer, returning nil if the UUID is Nil.
func UUIDToStringPtr(u uuid.UUID) *string {
	if u == uuid.Nil {
		return nil
	}
	str := u.String()
	return &str
}

// UUIDPtrToString converts a uuid.UUID pointer to a string, returning an empty string if nil.
func UUIDPtrToString(u *uuid.UUID) string {
	if u == nil {
		return ""
	}
	return u.String()
}
