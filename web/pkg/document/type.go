package document

import "go.trulyao.dev/hubble/web/schema"

//go:generate go tool github.com/abice/go-enum --marshal

// ENUM(link,audio,video,image,pdf,interchange,epub,word_document,presentation,spreadsheet,html,markdown,plain_text,archive,code,comment,other)
type EntryType string

func (e EntryType) CapnpType() schema.Type {
	switch e {
	case EntryTypeLink:
		return schema.Type_link

	case EntryTypeAudio:
		return schema.Type_audio

	case EntryTypeVideo:
		return schema.Type_video

	case EntryTypeImage:
		return schema.Type_image

	case EntryTypePdf:
		return schema.Type_pdf

	case EntryTypeInterchange:
		return schema.Type_interchange

	case EntryTypeEpub:
		return schema.Type_epub

	case EntryTypeWordDocument:
		return schema.Type_wordDocument

	case EntryTypePresentation:
		return schema.Type_presentation

	case EntryTypeSpreadsheet:
		return schema.Type_spreadsheet

	case EntryTypeHtml:
		return schema.Type_html

	case EntryTypeMarkdown:
		return schema.Type_markdown

	case EntryTypePlainText:
		return schema.Type_plainText

	case EntryTypeArchive:
		return schema.Type_archive

	case EntryTypeCode:
		return schema.Type_code

	case EntryTypeComment:
		return schema.Type_comment

	case EntryTypeOther:
		return schema.Type_other
	}

	return schema.Type_other
}

type QueuePayload struct {
	Type EntryType `json:"type,omitempty"`
}
