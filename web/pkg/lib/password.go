package lib

import (
	"github.com/matthewhartstonge/argon2"
	"go.trulyao.dev/seer"
)

type VerifyPasswordParams struct {
	// Password is the password to verify (the user-provided password)
	Password string
	// Hash is the argon2 hash to compare the password against
	Hash string
}

/*
VerifyPassword verifies a user-provided password against an argon2 hash.
*/
func VerifyPassword(params VerifyPasswordParams) (bool, error) {
	ok, err := argon2.VerifyEncoded([]byte(params.Password), []byte(params.Hash))
	if err != nil {
		return false, seer.Wrap("argon2_verify_encoded", err)
	}

	return ok, nil
}
