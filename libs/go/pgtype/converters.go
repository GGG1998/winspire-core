// Package pgtype provides shared type converters between pgx/v5 pgtype and standard Go types.
// This ensures consistent type conversion logic across all services.
package pgtype

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ============================================================================
// UUID Converters
// ============================================================================

// UUIDToPgtype converts a Go uuid.UUID to pgtype.UUID.
func UUIDToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// PgtypeToUUID converts a pgtype.UUID to Go uuid.UUID.
// Returns zero UUID if the pgtype.UUID is invalid.
func PgtypeToUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.UUID{}
	}
	return id.Bytes
}

// UUIDPtrToPgtype converts a Go *uuid.UUID to pgtype.UUID.
// Returns invalid pgtype.UUID if the pointer is nil.
func UUIDPtrToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

// PgtypeToUUIDPtr converts a pgtype.UUID to Go *uuid.UUID.
// Returns nil if the pgtype.UUID is invalid.
func PgtypeToUUIDPtr(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	uid := uuid.UUID(id.Bytes)
	return &uid
}

// ============================================================================
// Text/String Converters
// ============================================================================

// StringToPgtype converts a Go string to pgtype.Text.
func StringToPgtype(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

// PgtypeToString converts a pgtype.Text to Go string.
// Returns empty string if the pgtype.Text is invalid.
func PgtypeToString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// StringPtrToPgtype converts a Go *string to pgtype.Text.
// Returns invalid pgtype.Text if the pointer is nil.
func StringPtrToPgtype(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// PgtypeToStringPtr converts a pgtype.Text to Go *string.
// Returns nil if the pgtype.Text is invalid.
func PgtypeToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// ============================================================================
// Int32 Converters
// ============================================================================

// Int32ToPgtype converts a Go int32 to pgtype.Int4.
func Int32ToPgtype(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

// PgtypeToInt32 converts a pgtype.Int4 to Go int32.
// Returns 0 if the pgtype.Int4 is invalid.
func PgtypeToInt32(i pgtype.Int4) int32 {
	if !i.Valid {
		return 0
	}
	return i.Int32
}

// Int32PtrToPgtype converts a Go *int32 to pgtype.Int4.
// Returns invalid pgtype.Int4 if the pointer is nil.
func Int32PtrToPgtype(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

// PgtypeToInt32Ptr converts a pgtype.Int4 to Go *int32.
// Returns nil if the pgtype.Int4 is invalid.
func PgtypeToInt32Ptr(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	return &i.Int32
}

// ============================================================================
// Int Converters (int ↔ int32)
// ============================================================================

// IntToPgtype converts a Go int to pgtype.Int4 via int32.
func IntToPgtype(i int) pgtype.Int4 {
	return pgtype.Int4{Int32: int32(i), Valid: true}
}

// PgtypeToInt converts a pgtype.Int4 to Go int via int32.
// Returns 0 if the pgtype.Int4 is invalid.
func PgtypeToInt(i pgtype.Int4) int {
	if !i.Valid {
		return 0
	}
	return int(i.Int32)
}

// PgtypeToIntPtr converts a pgtype.Int4 to Go *int.
// Returns nil if the pgtype.Int4 is invalid.
func PgtypeToIntPtr(i pgtype.Int4) *int {
	if !i.Valid {
		return nil
	}
	val := int(i.Int32)
	return &val
}

// ============================================================================
// Bool Converters
// ============================================================================

// BoolToPgtype converts a Go bool to pgtype.Bool.
func BoolToPgtype(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}

// PgtypeToBool converts a pgtype.Bool to Go bool.
// Returns false if the pgtype.Bool is invalid.
func PgtypeToBool(b pgtype.Bool) bool {
	if !b.Valid {
		return false
	}
	return b.Bool
}

// BoolPtrToPgtype converts a Go *bool to pgtype.Bool.
// Returns pgtype.Bool{Bool: true, Valid: true} if the pointer is nil (default value).
func BoolPtrToPgtype(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Bool: true, Valid: true} // default value
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

// ============================================================================
// Timestamp Converters
// ============================================================================

// TimeToPgtype converts a Go time.Time to pgtype.Timestamp.
func TimeToPgtype(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: true}
}

// PgtypeToTime converts a pgtype.Timestamp to Go time.Time.
// Returns zero time if the pgtype.Timestamp is invalid.
func PgtypeToTime(t pgtype.Timestamp) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

// TimePtrToPgtype converts a Go *time.Time to pgtype.Timestamp.
// Returns invalid pgtype.Timestamp if the pointer is nil.
func TimePtrToPgtype(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: *t, Valid: true}
}

// PgtypeToTimePtr converts a pgtype.Timestamp to Go *time.Time.
// Returns nil if the pgtype.Timestamp is invalid.
func PgtypeToTimePtr(t pgtype.Timestamp) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// ============================================================================
// Timestamptz Converters (for timezone-aware timestamps)
// ============================================================================

// TimeToPgtypeTimestamptz converts a Go time.Time to pgtype.Timestamptz.
func TimeToPgtypeTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// PgtypeTimestamptzToTime converts a pgtype.Timestamptz to Go time.Time.
// Returns zero time if the pgtype.Timestamptz is invalid.
func PgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

// TimePtrToPgtypeTimestamptz converts a Go *time.Time to pgtype.Timestamptz.
// Returns invalid pgtype.Timestamptz if the pointer is nil.
func TimePtrToPgtypeTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// PgtypeTimestamptzToTimePtr converts a pgtype.Timestamptz to Go *time.Time.
// Returns nil if the pgtype.Timestamptz is invalid.
func PgtypeTimestamptzToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// PgtypeTimestamptzToString formats a pgtype.Timestamptz as ISO 8601 string.
// Returns empty string if the pgtype.Timestamptz is invalid.
func PgtypeTimestamptzToString(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format("2006-01-02T15:04:05Z07:00")
}

