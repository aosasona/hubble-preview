package lib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"

	"go.trulyao.dev/seer"
)

type (
	EncryptAesParams struct {
		// Key is the encryption key
		Key string

		// PlainText is the text to encrypt
		PlainText []byte
	}

	DecryptAesParams struct {
		// Key is the encryption key
		Key string

		// CipherText is the encrypted text
		CipherText []byte
	}
)

/*
EncryptAES encrypts the given plain text using the provided key.
*/
func EncryptAES(params EncryptAesParams) ([]byte, error) {
	if len(params.Key) != 32 {
		return nil, seer.New("encrypt_aes", "key must be 32 bytes")
	}

	if len(params.PlainText) == 0 {
		return nil, seer.New("encrypt_aes", "plain text must not be empty")
	}

	// Decode the key from hex
	key, err := hex.DecodeString(params.Key)
	if err != nil {
		return nil, seer.Wrap("decode_key_from_hex", err)
	}

	// Create cipher block
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, seer.Wrap("create_cipher", err)
	}

	// Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return nil, seer.Wrap("create_gcm", err)
	}

	// Generate a cryptographically secure nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, seer.Wrap("generate_nonce", err)
	}

	cipherText := gcm.Seal(nonce, nonce, params.PlainText, nil)
	return cipherText, nil
}

/*
DecryptAES decrypts the given cipher text using the provided key.
*/
func DecryptAES(params DecryptAesParams) ([]byte, error) {
	if len(params.Key) != 32 {
		return nil, seer.New("decrypt_aes", "key must be 32 bytes")
	}

	if len(params.CipherText) == 0 {
		return nil, seer.New("decrypt_aes", "cipher text must not be empty")
	}

	// Decode the key from hex
	key, err := hex.DecodeString(params.Key)
	if err != nil {
		return nil, seer.Wrap("decode_key_from_hex", err)
	}

	// Create cipher block
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, seer.Wrap("create_cipher", err)
	}

	// Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return nil, seer.Wrap("create_gcm", err)
	}

	nonceSize := gcm.NonceSize()

	// Extract the nonce from the beginning of the cipher text
	nonce, cipherText := params.CipherText[:nonceSize], params.CipherText[nonceSize:]

	// Decrypt the data
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, seer.Wrap("decrypt_data", err)
	}

	return plainText, nil
}
