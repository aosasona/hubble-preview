package rbac_test

import (
	"testing"

	"go.trulyao.dev/hubble/web/pkg/rbac"
)

type Test struct {
	Role   rbac.Role
	Target rbac.Role
	Result bool
}

func TestHasRole(t *testing.T) {
	tests := []Test{
		{rbac.RoleGuest, rbac.RoleGuest, true},
		{rbac.RoleGuest, rbac.RoleUser, false},
		{rbac.RoleGuest, rbac.RoleAdmin, false},
		{rbac.RoleGuest | rbac.RoleUser, rbac.RoleUser, true},
		{rbac.RoleGuest | rbac.RoleUser, rbac.RoleAdmin, false},
		{rbac.RoleGuest | rbac.RoleUser, rbac.RoleGuest, true},
		{rbac.RoleGuest | rbac.RoleUser, rbac.RoleGuest | rbac.RoleUser, true},
		{rbac.RoleGuest | rbac.RoleUser, rbac.RoleGuest | rbac.RoleAdmin, false},
	}

	for _, test := range tests {
		if test.Role.Has(test.Target) != test.Result {
			t.Errorf("expected %v to have role %v: %v", test.Role, test.Target, test.Result)
		}
	}
}

func TestCombineRoles(t *testing.T) {
	tests := []struct {
		Roles  []rbac.Role
		Result rbac.Role
	}{
		{[]rbac.Role{rbac.RoleGuest, rbac.RoleUser}, rbac.RoleGuest | rbac.RoleUser},
		{
			[]rbac.Role{rbac.RoleGuest, rbac.RoleUser, rbac.RoleAdmin},
			rbac.RoleGuest | rbac.RoleUser | rbac.RoleAdmin,
		},
		{
			[]rbac.Role{rbac.RoleGuest, rbac.RoleUser, rbac.RoleAdmin, rbac.RoleOwner},
			rbac.RoleGuest | rbac.RoleUser | rbac.RoleAdmin | rbac.RoleOwner,
		},
	}

	for _, test := range tests {
		if rbac.CombineRoles(test.Roles...) != test.Result {
			t.Errorf("expected %v to be combined to %v", test.Roles, test.Result)
		}
	}
}
