package lib

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

/*
Substring returns a substring of a string.

- If start is less than 0, it will be set to 0.

- If start is greater than the length of the string, an empty string will be returned.

- If end is greater than the length of the string, it will be set to the length of the string.

Example:

Substring("hello", 1, 3)  "el"
*/
func Substring(s string, start, end int) string {
	if start < 0 {
		start = 0
	}

	if start > len(s) {
		return ""
	}

	if end > len(s) {
		end = len(s)
	}

	return s[start:end]
}

/*
ToTitleCase capitalizes the first letter of a string.

Example:

ToTitleCase("hello")  "Hello"
*/
func ToTitleCase(s string) string {
	if len(s) == 0 {
		return s
	}

	return cases.Title(language.English).String(s)
}

/*
UppercaseFirst capitalizes the first letter of a string and makes the rest lowercase.

Example:
hello World  ->  Hello world
*/
func UppercaseFirst(s string) string {
	if len(s) == 0 {
		return s
	}

	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

/*
NormalizeTitle converts a string to proper title

Example:

Input: Redmon_you_only_look_cvpr_2016_paper
Output: Redmon You Only Look Cvpr 2016 Paper
*/
var titleRegexp = regexp.MustCompile(`[_-]`)

func NormalizeTitle(s string) string {
	s = titleRegexp.ReplaceAllString(s, " ")
	return ToTitleCase(s)
}

/*
Slugify converts a string into a slug.

Example:
The lazy dog  ->  the-lazy-dog
*/
var (
	slugRegexp         = regexp.MustCompile("[^a-z0-9]+")
	hangingHyphenRegex = regexp.MustCompile("(^-)|(-$)")
	hyphenRegex        = regexp.MustCompile("-{2,}")
)

func Slugify(s string) string {
	s = strings.ToLower(s)
	s = slugRegexp.ReplaceAllString(s, "-")
	s = hyphenRegex.ReplaceAllString(s, "-")
	return hangingHyphenRegex.ReplaceAllString(s, "")
}

/*
A generic function to check if a comparable value is empty

- Strings are trimmed and checked if they are empty

- Integers and floats are compared to 0

- Booleans are checked if they are false

- Nil values are considered empty
*/
func Empty[T comparable](v T) bool {
	value := reflect.ValueOf(v).Interface()

	switch value := value.(type) {
	case string:
		return strings.TrimSpace(value) == ""

	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return value == 0

	case float32, float64:
		return value == 0.0

	case bool:
		return !value

	case nil:
		return true

	default:
		return false
	}
}

/*
FormatOTP formats a one-time password (OTP) into a human-readable format.

Example:

FormatOTP("pL8uMdf4")  "pL8u-Mdf4"
*/
func FormatOTP(otp string) string {
	return fmt.Sprintf("%s-%s", otp[:4], otp[4:])
}

/*
RandomInt generates a random integer between a minimum and maximum value.
*/
func RandomInt(min, max int) int {
	b := make([]byte, 8)
	_, _ = rand.Read(b)

	return min + int(b[0])%(max-min)
}

// UUIDFromString converts a string to a pgtype.UUID object.
func UUIDFromString(s string) (pgtype.UUID, error) {
	uuid := pgtype.UUID{}

	if err := uuid.Scan(s); err != nil {
		log.Error().Err(err).Msg("failed to scan UUID")
		return pgtype.UUID{}, err
	}

	return uuid, nil
}

// RedactString redacts a string by replacing all characters except the first n characters with asterisks.
func RedactString(s string, visiblePartsLen int) string {
	if len(s) <= visiblePartsLen {
		return s
	}

	return fmt.Sprintf(
		"%s%s",
		Substring(s, 0, visiblePartsLen),
		strings.Repeat("*", len(s)-visiblePartsLen),
	)
}

/*
RedactEmail redacts an email address by replacing the middle characters with asterisks.
*/
func RedactEmail(email string, visiblePartsLen int) string {
	emailParts := strings.Split(email, "@")
	if len(emailParts) != 2 {
		return email
	}

	username := emailParts[0]
	domain := emailParts[1]

	if len(username) <= visiblePartsLen {
		return email
	}

	redactedUsername := fmt.Sprintf(
		"%s%s",
		Substring(username, 0, visiblePartsLen),
		strings.Repeat("*", len(username)-visiblePartsLen),
	)

	return fmt.Sprintf("%s@%s", redactedUsername, domain)
}

/*
Chunk splits a slice into chunks of a specified size.
*/
func Chunk[T any](secrets []T, size int) [][]T {
	chunks := make([][]T, 0, (len(secrets)+size-1)/size)
	for size < len(secrets) {
		secrets, chunks = secrets[size:], append(chunks, secrets[0:size:size])
	}
	return append(chunks, secrets)
}

/*
UniqueSlice removes duplicate elements from a slice.
*/
func UniqueSlice[T comparable](slice []T) []T {
	keys := make(map[T]struct{})
	list := make([]T, 0)

	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = struct{}{}
			list = append(list, entry)
		}
	}

	return list
}

/*
Diff compares two slices and returns the elements that are not present in both slices.
*/
func Diff[T comparable](a, b []T) []T {
	uniqueA := UniqueSlice(a)
	uniqueB := UniqueSlice(b)

	diffA := make([]T, 0)
	diffB := make([]T, 0)

	for _, entry := range uniqueA {
		if !slices.Contains(uniqueB, entry) {
			diffA = append(diffA, entry)
		}
	}

	for _, entry := range uniqueB {
		if !slices.Contains(uniqueA, entry) {
			diffB = append(diffB, entry)
		}
	}

	return UniqueSlice(append(diffA, diffB...))
}

// WithMaxSize returns a slice with a maximum length of maxSize. If the slice is longer than max, it will be truncated.
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

func WithMaxSize[T any, L Integer](slice []T, maxLen L) []T {
	if maxLen < 0 {
		maxLen = 0
	}

	if maxLen < L(len(slice)) {
		return slice[:maxLen]
	}

	return slice
}
