// ograph is a package that provides functionality to parse Open Graph metadata from a URL.
package ograph

import (
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		url     string
		options []Options
	}
	tests := []struct {
		name    string
		args    args
		want    *Metadata
		wantErr bool
	}{
		{
			name: "trulyao.dev",
			args: args{url: "https://trulyao.dev/posts/update-001-robin"},
			want: &Metadata{
				Title:       "What's new in Robin 0.5.0",
				Description: "Robin got a major(-ish) update, let's talk about what's changed since my last article about it.",
				Favicon:     "https://trulyao.dev/favicon.ico",
				Author:      "Ayodeji O.",
				Thumbnail:   "https://og.trulyao.dev/api/v1/images/trulyao/preview?variant=blog&style=blog&size=medium&vars=title%3AWhat%27s+new+in+Robin+0.5.0%2Cdate%3ANov+24+2024",
				SiteType:    "website",
				Domain:      "https://trulyao.dev",
				Link:        "https://trulyao.dev/posts/update-001-robin",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.url, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Title != tt.want.Title {
				t.Errorf("Parse(title) got = %v, want %v", got.Title, tt.want.Title)
			}

			if got.Description != tt.want.Description {
				t.Errorf(
					"Parse(description) got = %v, want %v",
					got.Description,
					tt.want.Description,
				)
			}

			if got.Favicon != tt.want.Favicon {
				t.Errorf("Parse(favicon) got = %v, want %v", got.Favicon, tt.want.Favicon)
			}

			if got.Author != tt.want.Author {
				t.Errorf("Parse(author) got = %v, want %v", got.Author, tt.want.Author)
			}

			if got.Thumbnail != tt.want.Thumbnail {
				t.Errorf("Parse(thumbnail) got = %v, want %v", got.Thumbnail, tt.want.Thumbnail)
			}

			if got.SiteType != tt.want.SiteType {
				t.Errorf("Parse(siteType) got = %v, want %v", got.SiteType, tt.want.SiteType)
			}

			if got.Domain != tt.want.Domain {
				t.Errorf("Parse(domain) got = %v, want %v", got.Domain, tt.want.Domain)
			}

			if got.Link != tt.want.Link {
				t.Errorf("Parse(link) got = %v, want %v", got.Link, tt.want.Link)
			}
		})
	}
}
