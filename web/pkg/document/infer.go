package document

import (
	"errors"
	"strings"

	"go.trulyao.dev/hubble/web/pkg/result"
)

var (
	ErrEmptyMimetype    = errors.New("mimetype is empty")
	ErrEmptyExtension   = errors.New("extension is empty")
	ErrUnknownExtension = errors.New("unknown extension")
	ErrUnknownMimetype  = errors.New("unknown mimetype")
	ErrNoSlash          = errors.New("mimetype does not contain a slash")
)

func InferExtension(filename string) string {
	// If there is no dot, there is no extension
	if strings.LastIndex(filename, ".") == -1 {
		return filename
	}

	return filename[strings.LastIndex(filename, ".")+1:]
}

func InferTypeFromExtension(extension string) result.Result[EntryType] {
	if extension == "" {
		return result.Err[EntryType](ErrEmptyExtension)
	}

	t := EntryTypeOther

	switch extension {
	case "jpeg", "png", "apng", "avif", "gif", "bmp", "svg", "tiff", "ico":
		t = EntryTypeImage

	case "mp4", "mpeg", "mp2t", "avi", "mov", "webm", "3gp", "3g2":
		t = EntryTypeVideo

	case "aac", "ogg", "mp3", "midi", "wav", "flac":
		t = EntryTypeAudio

	case "md", "markdown":
		t = EntryTypeMarkdown

	case "txt":
		t = EntryTypePlainText

	case "css", "php", "sh", "js", "ts", "jsx", "tsx":
		t = EntryTypeCode

	case "csv", "ods", "xls", "xlsx":
		t = EntryTypeSpreadsheet

	case "html", "htm":
		t = EntryTypeHtml

	case "epub", "mobi":
		t = EntryTypeEpub

	case "pdf":
		t = EntryTypePdf

	case "doc", "docx", "odt", "rtf":
		t = EntryTypeWordDocument

	case "ppt", "pptx", "odp", "pps", "ppsx":
		t = EntryTypePresentation

	case "json", "jsonc", "ld+json", "xml", "yaml", "yml", "toml", "ini", "cfg", "conf":
		t = EntryTypeInterchange

	case "zip", "gz", "bz", "bz2", "jar", "rar", "tar":
		t = EntryTypeArchive
	}

	if t == EntryTypeOther {
		return result.Err[EntryType](ErrUnknownExtension)
	}

	return result.Ok(t)
}

func InferTypeFromMimetype(mimeType string) result.Result[EntryType] {
	if mimeType == "" {
		return result.Err[EntryType](ErrEmptyMimetype)
	}

	// If there is no slash, it's not a valid MIME type
	if !strings.Contains(mimeType, "/") {
		return result.Err[EntryType](ErrNoSlash)
	}

	// Get the first part to the left of the slash
	category := strings.Split(mimeType, "/")[0]

	t := EntryTypeOther

	switch category {
	case "image":
		t = EntryTypeImage

	case "audio":
		t = EntryTypeAudio

	case "video":
		t = EntryTypeVideo

	default:
		switch mimeType {
		case "text/plain":
			t = EntryTypePlainText

		case "text/css",
			"application/x-httpd-php",
			"application/x-sh",
			"text/javascript":
			t = EntryTypeCode

		case "text/csv",
			"application/vnd.ms-excel",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"application/vnd.oasis.opendocument.spreadsheet":
			t = EntryTypeSpreadsheet

		case "text/html", "application/xhtml+xml":
			t = EntryTypeHtml

		case "application/epub+zip":
			t = EntryTypeEpub

		case "application/msword",
			"application/vnd.oasis.opendocument.text",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"application/rtf":
			t = EntryTypeWordDocument

		case "application/pdf":
			t = EntryTypePdf

		case "application/vnd.openxmlformats-officedocument.presentationml.presentation",
			"application/vnd.ms-powerpoint",
			"application/vnd.oasis.opendocument.presentation":
			t = EntryTypePresentation

		case "application/json", "application/ld+json", "application/xml":
			t = EntryTypeInterchange

		case "application/zip",
			"application/gzip",
			"application/x-bzip",
			"application/x-bzip2",
			"application/java-archive",
			"application/vnd.rar",
			"application/x-7z-compressed",
			"application/x-tar":
			t = EntryTypeArchive

		default:
			t = EntryTypeOther
		}
	}

	if t == EntryTypeOther {
		return result.Err[EntryType](ErrUnknownMimetype)
	}

	return result.Ok(t)
}

func InferType(extension, mimeType string) EntryType {
	normalize := func(s string) string {
		return strings.ToLower(strings.TrimSpace(s))
	}

	extension = normalize(extension)
	mimeType = normalize(mimeType)

	// If the extension contains a dot, remove it and infer the extension
	if strings.Contains(extension, ".") {
		extension = InferExtension(extension)
	}

	// If both the extension and MIME type are empty, return "other"
	if extension == "" && mimeType == "" {
		return EntryTypeOther
	}

	// Attempt to infer the type from the extension
	if extension != "" {
		if t := InferTypeFromExtension(extension); t.IsOk() {
			return t.Unwrap(EntryTypeOther)
		}
	}

	// Attempt to infer the type from the MIME type
	if mimeType != "" {
		if t := InferTypeFromMimetype(mimeType); t.IsOk() {
			return t.Unwrap(EntryTypeOther)
		}
	}

	return EntryTypeOther
}
