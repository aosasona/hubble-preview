package spec

import "slices"

//go:generate go tool github.com/abice/go-enum --marshal

// ENUM(noop,log::debug,log::error,log::warn,entries::update,chunks::create,network::request,transform::chunk_with_overlap,transform::chunk_by_sentence,transform::url_to_markdown,transform::html_to_markdown,store::get,store::set,store::delete,store::all,store::clear,crypto::rand)
type Perm string

type (
	Privilege struct {
		// Identifier is the identifier of the privilege
		Identifier Perm `json:"identifier"  toml:"identifier"`
		// Description is the reason why this privilege is needed
		Description string `json:"description" toml:"description"`
	}

	Privileges []Privilege
)

func (p *Privileges) Add(identifier Perm, description string) {
	*p = append(*p, Privilege{
		Identifier:  identifier,
		Description: description,
	})
}

func (p *Privileges) Remove(identifier Perm) {
	for i, privilege := range *p {
		if privilege.Identifier == identifier {
			*p = slices.Delete(*p, i, i+1)
			return
		}
	}
}

func (p Privileges) Has(identifier Perm) bool {
	for _, privilege := range p {
		if privilege.Identifier == identifier {
			return true
		}
	}
	return false
}
