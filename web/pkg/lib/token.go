package lib

import (
	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/matthewhartstonge/argon2"
	"go.trulyao.dev/seer"
)

type Token struct {
	scope   string
	text    string
	encoded []byte
}

func DefaultArgon2Config() argon2.Config {
	return argon2.Config{
		HashLength:  32,
		SaltLength:  16,
		TimeCost:    2,
		MemoryCost:  32 * 1024, // 2^(16) (32MiB of RAM)
		Parallelism: 2,
		Mode:        argon2.ModeArgon2id,
		Version:     argon2.Version13,
	}
}

// String implements the fmt.Stringer interface, it returns the raw token text.
func (t Token) String() string { return t.text }

// Encoded returns the encoded token but as a string, see RawEncoded for the byte slice.
func (t Token) Encoded() string { return string(t.encoded) }

// EncodedByte returns the encoded token as a byte slice, see Encoded for the string.
func (t Token) EncodedByte() []byte { return t.encoded }

// Scope returns the scope of the token.
func (t Token) Scope() string { return t.scope }

// GenerateToken generates a token with a given length and encodes it using argon2.
//
// The default token length is 8 characters, but you can specify a custom length by passing it as an argument.
func GenerateToken(scope string, length ...int) (Token, error) {
	var token Token

	tokenLength := 8
	if len(length) > 0 && length[0] > 0 {
		tokenLength = length[0]
	}

	// Generate a token
	rawToken, err := nanoid.Generate(
		"abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRTVWXYZ123456789",
		tokenLength,
	)
	if err != nil {
		return token, seer.Wrap("generate_nanoid_"+scope, err)
	}

	argon := DefaultArgon2Config()
	encodedToken, err := argon.HashEncoded([]byte(rawToken))
	if err != nil {
		return token, seer.Wrap("encode_token_"+scope, err)
	}

	return Token{text: rawToken, encoded: encodedToken, scope: scope}, nil
}

func GenerateAuthToken() (Token, error) {
	return GenerateToken("auth", 24)
}
