package models

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pquerna/otp"
	"go.trulyao.dev/hubble/web/pkg/lib"
	"go.trulyao.dev/seer"
)

//go:generate go tool github.com/abice/go-enum --marshal

// ENUM(email,totp,totp_enrollment)
type MfaSessionType string

type MfaSessionMeta interface {
	AccountName() string
	Type() MfaSessionType
}

type EmailMeta struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (e *EmailMeta) AccountName() string {
	if e.Name == "" {
		return e.Email
	}

	return e.Name
}

func (e *EmailMeta) Type() MfaSessionType {
	return MfaSessionTypeEmail
}

type TotpMeta struct {
	Name string `json:"name"`
}

func (t *TotpMeta) AccountName() string {
	return t.Name
}

func (t *TotpMeta) Type() MfaSessionType {
	return MfaSessionTypeTotp
}

type TotpEnrollmentMeta struct {
	Name string `json:"name"`
	// This is used to derive other properties of the TOTP key like the URL
	RawKeyUrl string `json:"original_string"`
}

func (t *TotpEnrollmentMeta) AccountName() string {
	return t.Name
}

func (t *TotpEnrollmentMeta) Type() MfaSessionType {
	return MfaSessionTypeTotpEnrollment
}

func (t *TotpEnrollmentMeta) Key() (*otp.Key, error) {
	return otp.NewKeyFromURL(t.RawKeyUrl)
}

// -- MARK: MfaSession
func (session *MfaSession) UnmarshalJSON(data []byte) error {
	// We will first unmarshal into a lose map to determine the type of the session
	raw := make(map[string]any)

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if _, ok := raw["session_type"]; !ok {
		return ErrInvalidMfaSessionType
	}

	// We will switch on the type of the session
	switch raw["session_type"].(string) {
	case MfaSessionTypeEmail.String():
		session.SessionType = MfaSessionTypeEmail
		session.Meta = &EmailMeta{}
	case MfaSessionTypeTotp.String():
		session.SessionType = MfaSessionTypeTotp
		session.Meta = &TotpMeta{}
	case MfaSessionTypeTotpEnrollment.String():
		session.SessionType = MfaSessionTypeTotpEnrollment
		session.Meta = &TotpEnrollmentMeta{}
	default:
		return ErrInvalidMfaSessionType
	}

	// We will now unmarshal the data into the meta itself with mapstructure
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: false,
		ErrorUnset:  false,
		Result:      session,
		TagName:     "json",
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			TimeHookFunc(),
			UUIDHookFunc(),
		),
	})
	if err != nil {
		return seer.Wrap("create_mapstructure_decoder", err)
	}

	if err := decoder.Decode(raw); err != nil {
		return err
	}

	return nil
}

func UUIDHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if t != reflect.TypeOf(pgtype.UUID{}) {
			return data, nil
		}

		switch f.Kind() {
		case reflect.String:
			return lib.UUIDFromString(data.(string))
		default:
			return data, nil
		}
	}
}

func TimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch f.Kind() {
		case reflect.String:
			return time.Parse(time.RFC3339, data.(string))
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
	}
}

var (
	_ MfaSessionMeta = (*EmailMeta)(nil)
	_ MfaSessionMeta = (*TotpMeta)(nil)
	_ MfaSessionMeta = (*TotpEnrollmentMeta)(nil)

	_ json.Unmarshaler = (*MfaSession)(nil)
)
