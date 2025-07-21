package models

import "github.com/jackc/pgx/v5/pgtype"

type TotpSecret struct {
	ID        int32       `json:"id"`
	AccountID pgtype.UUID `json:"account_id"`
	Hash      []byte      `json:"-"`
	Version   int16       `json:"version"`
}
