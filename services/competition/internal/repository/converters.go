package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ============================================================================
// Shared Type Converters (pgtype <-> standard Go types)
// Used across all repositories to avoid duplication
// ============================================================================

// UUID converters

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgtypeToUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.UUID{}
	}
	return id.Bytes
}

func uuidToPgtypeUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func pgtypeUUIDToUUIDPtr(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	uid := uuid.UUID(id.Bytes)
	return &uid
}

// Text/String converters

func stringToPgtypeText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func pgtypeTextToString(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func pgtypeText(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  true,
	}
}

// Timestamp converters

func timeToPgtypeTimestamp(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func pgtypeTimestampToTime(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func pgtypeTimestampToString(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format("2006-01-02T15:04:05Z07:00")
}

// Int32 converters

func int32ToPgtypeInt4(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

func int32ValueToPgtypeInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

func pgtypeInt4ToInt32(i pgtype.Int4) int32 {
	if !i.Valid {
		return 0
	}
	return i.Int32
}

func pgtypeInt4ToInt32Ptr(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	return &i.Int32
}

// Bool converters

func boolToPgtypeBool(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Bool: true, Valid: true} // default value
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

func pgtypeBoolToBool(b pgtype.Bool) bool {
	if !b.Valid {
		return false
	}
	return b.Bool
}

