package models

import (
	"time"

	"go.trulyao.dev/hubble/web/internal/database/queries"
)

type User struct {
	// ID is the private ID of the user (not exposed to the API)
	ID int32 `json:"id"`

	// PublicID is the public ID of the user, this is a UUID and is safe to expose to the API
	PublicID string `json:"user_id" mirror:"type:string"`

	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	HashedPassword string    `json:"-"`
	Email          string    `json:"email"`
	EmailVerified  bool      `json:"email_verified"`
	Username       string    `json:"username"`
	AvatarID       string    `json:"avatar_id"       mirror:"type:string"`
	CreatedAt      time.Time `json:"created_at"      mirror:"type:number"`
	LastUpdatedAt  time.Time `json:"last_updated_at" mirror:"type:number"`
}

func (u *User) From(user *queries.User) User {
	*u = User{
		ID:             user.ID,
		PublicID:       user.PublicID.String(),
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		HashedPassword: user.HashedPassword,
		Email:          user.Email,
		EmailVerified:  user.EmailVerified,
		Username:       user.Username,
		AvatarID:       user.AvatarID.String,
		CreatedAt:      user.CreatedAt.Time,
		LastUpdatedAt:  user.UpdatedAt.Time,
	}

	return *u
}
