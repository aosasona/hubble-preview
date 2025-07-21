package rbac

import (
	"errors"
	"strings"
)

type Role uint32

const (
	// RoleOwner is the highest role and has all permissions
	RoleOwner Role = (1 << 30) // as high as 32-bit signed integer can go

	// RoleGuest is the lowest role and has the least permissions
	RoleGuest Role = (1 << iota)

	// RoleUser is the role for authenticated users and members of a workspace
	RoleUser

	// RoleAdmin is the role for users who have admin permissions in a workspace
	RoleAdmin
)

var RolesMapping = map[Role]string{
	RoleGuest: "guest",
	RoleUser:  "user",
	RoleAdmin: "admin",
	RoleOwner: "owner",
}

func RoleFromString(role string) Role {
	role = strings.ToLower(role)
	for k, v := range RolesMapping {
		if v == role {
			return k
		}
	}
	return 0
}

func (r *Role) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return errors.New("empty role")
	}

	if text[0] == '"' && text[len(text)-1] == '"' {
		text = text[1 : len(text)-1]
	}

	if len(text) == 0 {
		return errors.New("empty role")
	}

	for k, v := range RolesMapping {
		if v == string(text) {
			*r = k
			return nil
		}
	}

	return errors.New("invalid role")
}

func (r Role) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r Role) String() string {
	if name, ok := RolesMapping[r]; ok {
		return name
	}
	return "unknown"
}

func (role Role) Can(perm Permission) bool {
	required, ok := permissions[perm]
	if !ok {
		return false
	}

	return role&required != 0
}

func (role Role) Has(target Role) bool {
	return role&target == target
}

func CombineRoles(roles ...Role) Role {
	var result Role
	for _, role := range roles {
		result |= role
	}
	return result
}
