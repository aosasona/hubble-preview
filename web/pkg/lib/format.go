package lib

import (
	"fmt"
	"path/filepath"
	"time"
)

/*
StripExtension removes the file extension from a filename.
*/
func StripExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return filename
	}

	return filename[:len(filename)-len(ext)]
}

/*
Pluralize returns the plural form of a word based on the count provided.

Example: Pluralize("apple", 2) returns "apples"
*/
func Pluralize(singular string, count int) string {
	if count == 1 {
		return singular
	}

	if singular[len(singular)-1] == 'y' {
		return singular[:len(singular)-1] + "ies"
	}

	return singular + "s"
}

/*
ToHumanReadableDate converts a time.Time object to a human-readable date string for use in, for example, an email.

Example: "Monday, 2 January 2006 at 3:04 PM"
*/
func ToHumanReadableDate(t time.Time) string {
	var formattedDate string

	if t.Hour() == 0 && t.Minute() == 0 {
		formattedDate = t.Format("Monday, 2 January 2006")
	} else {
		formattedDate = t.Format("Monday, 2 January 2006 at 3:04 PM")
	}

	return formattedDate
}

/*
ToHumanReadableDuration converts a time.Duration object to a human-readable duration string.

Example: "1 day, 2 hours, 3 minutes and 4 seconds"
*/
func ToHumanReadableDuration(duration time.Duration) string {
	duration = duration.Round(time.Second)
	formattedDuration := ""

	if duration == 0 {
		return "0 seconds"
	}

	if duration.Hours() >= 24 {
		days := int(duration.Hours() / 24)
		formattedDuration += fmt.Sprintf("%d day", days)
		if days > 1 {
			formattedDuration += "s"
		}

		duration -= time.Duration(days) * 24 * time.Hour
	}

	if hours := int(duration.Hours()); hours > 0 {
		if formattedDuration != "" {
			formattedDuration += ", "
		}

		formattedDuration += fmt.Sprintf("%d hour", hours)
		if hours > 1 {
			formattedDuration += "s"
		}

		duration -= time.Duration(hours) * time.Hour
	}

	if minutes := int(duration.Minutes()); minutes > 0 {
		if formattedDuration != "" {
			formattedDuration += ", "
		}

		formattedDuration += fmt.Sprintf("%d minute", minutes)
		if minutes > 1 {
			formattedDuration += "s"
		}

		duration -= time.Duration(minutes) * time.Minute
	}

	if duration.Seconds() > 0 {
		seconds := int(duration.Seconds())
		if formattedDuration != "" {
			formattedDuration += " and "
		}

		formattedDuration += fmt.Sprintf("%d second", seconds)
		if seconds > 1 {
			formattedDuration += "s"
		}
	}

	return formattedDuration
}
