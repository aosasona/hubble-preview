package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrInvalidCookieEncoding  = errors.New("invalid cookie encoding")
	ErrInvalidSignatureLength = errors.New("invalid signature length")
	ErrInvalidSignature       = errors.New("cookie signature is invalid")
)

const (
	CookieAuthSession = "auth_session"
)

const (
	// StateKeyUserID is the key used to store the `User` in the context (type: `int32`).
	StateKeyUserID = "user_id"
	// StateKeySession is the key used to store the `Session` in the contex (type: `repository.Session`).
	StateKeySession = "session"
)

type (
	SignedCookie struct {
		name      string
		value     string
		signature string
	}

	SignCookieArgs struct {
		Name   string
		Value  string
		Secret string
	}
)

// Name returns the name of the signed cookie.
func (s SignedCookie) Name() string { return s.name }

// Value returns the value of the signed cookie.
func (s SignedCookie) Value() string { return s.value }

// Signature returns the signature of the signed cookie.
func (s SignedCookie) Signature() string { return s.signature }

/*
String returns the debug-friendly string representation of the signed cookie.

WARNING: this should not be used as the cookie value
*/
func (s SignedCookie) String() string {
	return fmt.Sprintf(
		"Cookie{name=%s, value=%s, signature=%x}",
		s.name,
		s.value,
		s.signature,
	)
}

// Base64EncodedValue returns the signed cookie as a base64 encoded string.
func (s SignedCookie) Base64EncodedValue() string {
	strCookie := string(s.signature) + s.value
	return base64.StdEncoding.EncodeToString([]byte(strCookie))
}

func hash(secret []byte, data ...string) []byte {
	h := hmac.New(sha256.New, secret)
	for _, d := range data {
		h.Write([]byte(d))
	}
	return h.Sum(nil)
}

/*
SignCookie signs a cookie value with a secret.
*/
func SignCookie(args SignCookieArgs) SignedCookie {
	signature := hash([]byte(args.Secret), args.Name, args.Value)

	return SignedCookie{
		name:      args.Name,
		value:     args.Value,
		signature: string(signature),
	}
}

/*
DecodeSignedCookie reads a signed cookie, making sure the signature is valid and returning the cookie value.
*/
func DecodeSignedCookie(
	cookie *http.Cookie,
	secret string,
) (SignedCookie, error) {
	rawValue, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return SignedCookie{}, ErrInvalidCookieEncoding
	}

	// The signature should be in this format: <signature><value>
	// Ensure our signature is at least the size of a valid SHA-256 hash.
	if len(rawValue) < sha256.Size {
		return SignedCookie{}, ErrInvalidSignatureLength
	}

	signature := rawValue[:sha256.Size]
	value := rawValue[sha256.Size:]
	expectedSignature := hash([]byte(secret), cookie.Name, string(value))

	if !hmac.Equal(signature, expectedSignature) {
		return SignedCookie{}, ErrInvalidSignature
	}

	return SignedCookie{
		name:      cookie.Name,
		value:     string(value),
		signature: string(signature),
	}, nil
}
