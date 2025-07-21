package lib

import (
	"slices"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

// List of supported PostgreSQL full-text search configurations (languages)
var supportedLanguages = []string{
	"simple", "english", "arabic", "danish", "dutch", "finnish", "french",
	"german", "greek", "hungarian", "indonesian", "irish", "italian",
	"lithuanian", "nepali", "norwegian", "portuguese", "romanian",
	"russian", "spanish", "swedish", "tamil", "turkish",
}

// Checks if a language is supported. If not, returns "simple"
func NormalizePgLanguage(lang string) string {
	lang = strings.ToLower(lang)
	if slices.Contains(supportedLanguages, lang) {
		return lang
	}
	return "simple"
}

func PgText(text string) pgtype.Text {
	if strings.TrimSpace(text) == "" {
		return pgtype.Text{String: "", Valid: false}
	}

	return pgtype.Text{String: text, Valid: true}
}

func PgInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{Valid: i != 0, Int32: i}
}

func PgUUIDString(uuid string) pgtype.UUID {
	if uuid == "" {
		return pgtype.UUID{Bytes: [16]byte{}, Valid: false}
	}

	u, err := UUIDFromString(uuid)
	if err != nil {
		return pgtype.UUID{Bytes: [16]byte{}, Valid: false}
	}

	return u
}
