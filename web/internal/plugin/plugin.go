package plugin

import (
	"strings"

	"github.com/BurntSushi/toml"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	"go.trulyao.dev/hubble/web/pkg/document"
	"go.trulyao.dev/hubble/web/pkg/lib"
)

type PluginV1 struct {
	PluginName        string               `json:"name"        toml:"name"        validate:"required,mixed_name,min=2"`
	PluginDescription string               `json:"description" toml:"description" validate:"required,ascii"`
	PluginModes       []queries.PluginMode `json:"modes"       toml:"modes"       validate:"required,dive,oneof=on_create background" mirror:"type:Array<'on_create' | 'background'>"`
	PluginTargets     []document.EntryType `json:"targets"     toml:"targets"     validate:"required,dive,alpha"                      mirror:"type:'link' | 'audio' | 'video' | 'image' | 'pdf' | 'interchange' | 'epub' | 'word_document' | 'presentation' | 'spreadsheet' | 'html' | 'markdown' | 'plain_text' | 'archive' | 'code' | 'comment' | 'other' | '*'"`
	PluginPrivileges  spec.Privileges      `json:"privileges"  toml:"privileges"`

	// OutputChecksum is the checksum of the plugin binary that we generate internally
	OutputChecksum []byte `json:"-" toml:"-"`
	// ProvidedChecksum is the checksum provided by the plugin author for verification
	ProvidedChecksum []byte `json:"-" toml:"-"`

	AbsolutePathInCache string `json:"-"`
}

// IsDifferentVersion implements spec.Plugin.
func (p *PluginV1) IsDifferentTo(other spec.ComparablePlugin) bool {
	// Compare checksums
	if other.Checksum() != p.Checksum() {
		return true
	}

	trim := func(s string) string {
		return strings.TrimSpace(strings.ToLower(s))
	}

	// Find changes in name
	if trim(p.PluginName) != trim(other.Name()) {
		return true
	}

	// Find changes in description
	if trim(p.PluginDescription) != trim(other.Description()) {
		return true
	}

	// Find changes in targets
	if len(p.PluginTargets) != len(other.Targets()) {
		return true
	} else if len(lib.Diff(p.PluginTargets, other.Targets())) > 0 {
		return true
	}

	// Find changes in privileges
	if len(p.PluginPrivileges) != len(other.Privileges()) {
		return true
	}

	for _, privilege := range other.Privileges() {
		if !p.PluginPrivileges.Has(privilege.Identifier) {
			return true
		}
	}

	for _, privilege := range p.PluginPrivileges {
		if !other.Privileges().Has(privilege.Identifier) {
			return true
		}
	}

	// Find changes in modes
	if len(p.PluginModes) != len(other.Modes()) {
		return true
	} else if len(lib.Diff(p.PluginModes, other.Modes())) > 0 {
		return true
	}

	return false
}

// Checksum implements spec.Plugin.
func (p *PluginV1) Checksum() string {
	return string(p.OutputChecksum)
}

func ParsePluginV1(data []byte) (*PluginV1, error) {
	plugin := PluginV1{}
	if err := toml.Unmarshal(data, &plugin); err != nil {
		return nil, err
	}

	if err := lib.ValidateStruct(&plugin); err != nil {
		return nil, err
	}

	return &plugin, nil
}

func (p *PluginV1) CachePath() string {
	return p.AbsolutePathInCache
}

// HasVerificationSHA256 implements spec.Plugin.
func (p *PluginV1) HasVerificationSHA256() bool {
	return len(p.ProvidedChecksum) > 0
}

// Description implements spec.Plugin.
func (p *PluginV1) Description() string {
	return p.PluginDescription
}

// Modes implements spec.Plugin.
func (p *PluginV1) Modes() []queries.PluginMode {
	return p.PluginModes
}

// Name implements spec.Plugin.
func (p *PluginV1) Name() string {
	return p.PluginName
}

// Privileges implements spec.Plugin.
func (p *PluginV1) Privileges() []spec.Privilege {
	return p.PluginPrivileges
}

// Targets implements spec.Plugin.
func (p *PluginV1) Targets() []document.EntryType {
	return p.PluginTargets
}

var _ spec.Plugin = (*PluginV1)(nil)
