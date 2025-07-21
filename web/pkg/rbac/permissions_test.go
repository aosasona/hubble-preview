package rbac_test

import (
	"testing"

	"go.trulyao.dev/hubble/web/pkg/rbac"
)

func Test_Can(t *testing.T) {
	type test struct {
		name string
		role rbac.Role
		perm rbac.Permission
		want bool
	}

	tests := []test{
		{
			name: "Owner can delete workspace",
			role: rbac.RoleOwner,
			perm: rbac.PermDeleteWorkspace,
			want: true,
		},
		{
			name: "Owner can create collection",
			role: rbac.RoleOwner,
			perm: rbac.PermCreateCollection,
			want: true,
		},
		{
			name: "Admin can create collection",
			role: rbac.RoleAdmin,
			perm: rbac.PermCreateCollection,
			want: true,
		},
		{
			name: "User can create collection",
			role: rbac.RoleUser,
			perm: rbac.PermCreateCollection,
			want: true,
		},
		{
			name: "Guest cannot create collection",
			role: rbac.RoleGuest,
			perm: rbac.PermCreateCollection,
			want: false,
		},
		{
			name: "owner can invite user to workspace",
			role: rbac.RoleOwner,
			perm: rbac.PermInviteUsersToWorkspace,
			want: true,
		},
		{
			name: "admin can invite user to workspace",
			role: rbac.RoleAdmin,
			perm: rbac.PermInviteUsersToWorkspace,
			want: true,
		},
		{
			name: "user cannot invite user to workspace",
			role: rbac.RoleUser,
			perm: rbac.PermInviteUsersToWorkspace,
			want: false,
		},
		{
			name: "guest cannot invite user to workspace",
			role: rbac.RoleGuest,
			perm: rbac.PermInviteUsersToWorkspace,
			want: false,
		},
	}

	for _, tt := range tests {
		if got := tt.role.Can(tt.perm); got != tt.want {
			t.Errorf(
				"{\nname: %s,\nrole: %d,\nperm: %s\n}: got %t, want %t",
				tt.name,
				tt.role,
				tt.perm,
				got,
				tt.want,
			)
		}
	}
}
