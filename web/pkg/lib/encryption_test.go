package lib_test

import (
	"testing"

	"go.trulyao.dev/hubble/web/pkg/lib"
)

func Test_AesEncryptionDecryption(t *testing.T) {
	type test struct {
		name    string
		key     string
		text    string
		wantErr bool
	}

	tests := []test{
		{
			name:    "valid key and text",
			key:     "25c5bbdf0c705e73ab71f75de01f3ca5",
			text:    "hello world",
			wantErr: false,
		},
		{
			name:    "invalid key",
			key:     "",
			text:    "hello world",
			wantErr: true,
		},
		{
			name:    "valid key and empty text",
			key:     "25c5bbdf0c705e73ab71f75de01f3ca5",
			text:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := lib.EncryptAES(lib.EncryptAesParams{
				Key:       tt.key,
				PlainText: []byte(tt.text),
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("[%s] EncryptAES() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			text, err := lib.DecryptAES(lib.DecryptAesParams{
				Key:        tt.key,
				CipherText: hash,
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("[%s] DecryptAES() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			if err == nil && string(text) != tt.text {
				t.Errorf("[%s] DecryptAES() = %v, want %v", tt.name, string(text), tt.text)
			}
		})
	}
}
