package queries

import "go.trulyao.dev/hubble/web/schema"

type PluginPrivilege struct {
	// Identifier is the identifier of the privilege
	Identifier string `json:"identifier"`
	// Description is the reason why this privilege is needed
	Description string `json:"description"`
}

func (es EntryStatus) ToCapnpType() schema.Status {
	switch es {
	case EntryStatusQueued:
		return schema.Status_queued

	case EntryStatusProcessing:
		return schema.Status_processing

	case EntryStatusCompleted:
		return schema.Status_completed

	case EntryStatusFailed:
		return schema.Status_failed

	case EntryStatusCanceled:
		return schema.Status_canceled

	case EntryStatusPaused:
		return schema.Status_paused

	default:
		return schema.Status_queued

	}
}
