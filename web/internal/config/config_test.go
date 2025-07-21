package config

import "testing"

func Test_parseTotpKey(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		want    TotpKey
		wantErr bool
	}{
		{
			name:   "valid secret",
			secret: "v1_20a6810db4bcde2700785c86",
			want:   TotpKey{1, "20a6810db4bcde2700785c86"},
		},
		{
			name:    "missing version",
			secret:  "20a6810db4bcde2700785c86",
			wantErr: true,
		},
		{
			name:    "invalid version",
			secret:  "vX_20a6810db4bcde2700785c86",
			wantErr: true,
		},
		{
			name:    "missing key",
			secret:  "v1",
			wantErr: true,
		},
		{
			name:    "short key",
			secret:  "v1_20a6810db",
			wantErr: true,
		},
		{
			name:    "empty key",
			secret:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTotpKey(tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("[%s] parseTotpKey() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if err == nil && (got.version != tt.want.version || got.secret != tt.want.secret) {
				t.Errorf("[%s] parseTotpKey() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
