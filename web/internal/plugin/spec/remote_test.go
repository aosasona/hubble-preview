package spec_test

import (
	"testing"

	"go.trulyao.dev/hubble/web/internal/plugin/spec"
)

func Test_ParseSSH(t *testing.T) {
	tests := []struct {
		rawUrl  string
		source  string
		wantErr bool
	}{
		{"ssh://github.com/username/repo.git", "", true},
		{"ssh://git:johndoe@github.com/project.git", "git@github.com:johndoe/project.git", false},
		{"ssh://git:johndoe@github.com", "git@github.com:johndoe/project.git", true},
	}

	for _, test := range tests {
		source, err := spec.ParseRemoteSource(test.rawUrl)
		if (err != nil) != test.wantErr {
			t.Errorf("ParseRemoteSource(%q) error = %v, wantErr %v", test.rawUrl, err, test.wantErr)
		}

		if err == nil && source.FormattedGitURL() != test.source {
			t.Errorf(
				"ParseRemoteSource(%q) = %q, want %q",
				test.rawUrl,
				source.FormattedGitURL(),
				test.source,
			)
		}
	}
}

func Test_ParseHTTP(t *testing.T) {
	tests := []struct {
		rawUrl  string
		source  string
		wantErr bool
	}{
		{"http://github.com/johndoe", "", true},
		{"http://github.com/johndoe/foo.git", "http://github.com/johndoe/foo.git", false},
		{"https://github.com/johndoe/foo.git", "https://github.com/johndoe/foo.git", false},
		{
			"https://foo:bar@github.com/johndoe/foo.git",
			"https://foo:bar@github.com/johndoe/foo.git",
			false,
		},
		{
			"git://foo:bar@github.com/johndoe/foo.git",
			"https://foo:bar@github.com/johndoe/foo.git",
			false,
		},
		{
			"https://foo@bar:github.com/johndoe/foo.git",
			"https://foo:bar@github.com/johndoe/foo.git",
			true,
		},
		{
			"https://foo@bargithub.com/johndoe/foo.git",
			"",
			true,
		},
	}

	for _, test := range tests {
		source, err := spec.ParseRemoteSource(test.rawUrl)
		if (err != nil) != test.wantErr {
			t.Errorf("ParseRemoteSource(%q) error = %v, wantErr %v", test.rawUrl, err, test.wantErr)
		}

		if err == nil && source.FormattedGitURL() != test.source {
			t.Errorf(
				"ParseRemoteSource(%q) = %q, want %q",
				test.rawUrl,
				source.FormattedGitURL(),
				test.source,
			)
		}
	}
}
