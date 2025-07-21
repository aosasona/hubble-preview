package plugin

import (
	"github.com/BurntSushi/toml"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	"go.trulyao.dev/hubble/web/pkg/lib"
	"go.trulyao.dev/seer"
)

func ParseSourceV1(data []byte) (*SourceV1, error) {
	source := SourceV1{}
	if err := toml.Unmarshal(data, &source); err != nil {
		return nil, err
	}

	_, err := spec.ParseRemoteSource(source.RemoteURL)
	if err != nil {
		return nil, seer.Wrap("parse_remote_source", err)
	}

	if err := lib.ValidateStruct(&source); err != nil {
		return nil, err
	}

	return &source, nil
}

type SourceV1 struct {
	SourceName               string                  `json:"name"                toml:"name"                validate:"required,mixed_name,min=2"`
	SourceAuthor             string                  `json:"author"              toml:"author"              validate:"required,mixed_name,min=2"`
	SourceDescription        string                  `json:"description"         toml:"description"         validate:"required,ascii"`
	SourcePlugins            []string                `json:"plugins"             toml:"plugins"             validate:"required,dive,ascii"`
	SourceVersioningStrategy spec.VersioningStrategy `json:"versioning_strategy" toml:"versioning-strategy"                                      mirror:"type:'on_create' | 'background'"`
	RemoteURL                string                  `json:"url"                 toml:"url"                 validate:"required"`
}

// Description implements spec.Source.
func (s *SourceV1) Description() string {
	return s.SourceDescription
}

// Plugins implements spec.Source.
func (s *SourceV1) Plugins() []string {
	return s.SourcePlugins
}

// URL implements spec.Source.
func (s *SourceV1) URL() string {
	return s.RemoteURL
}

// VersioningStrategy implements spec.Source.
func (s *SourceV1) VersioningStrategy() spec.VersioningStrategy {
	return s.SourceVersioningStrategy
}

// Author implements spec.Source.
func (s *SourceV1) Author() string {
	return s.SourceAuthor
}

// Name implements spec.Source.
func (s *SourceV1) Name() string {
	return s.SourceName
}

var _ spec.Source = (*SourceV1)(nil)
