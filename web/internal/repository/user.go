package repository

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/matthewhartstonge/argon2"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/internal/kv"
	"go.trulyao.dev/hubble/web/internal/otp"
	"go.trulyao.dev/hubble/web/internal/models"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
	"go.trulyao.dev/seer"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

var (
	ErrAccountWithEmailNotFound = apperrors.New(
		"no account with this email found",
		http.StatusNotFound)

	ErrAccountWithIDNotFound = apperrors.New(
		"no account with this ID found",
		http.StatusNotFound)
)

type UpdateProfileParams struct {
	FirstName string
	LastName  string
	Username  string
}

type UserRepository interface {
	// EmailExists checks if an email already exists.
	EmailExists(email string) (bool, error)

	// UsernameExists checks if a username already exists.
	UsernameExists(username string) (bool, error)

	// CreateUser creates a new user account.
	CreateUser(queries.CreateUserParams) (*models.User, error)

	// ChangePassword changes a user's password (obviously)
	ChangePassword(id int32, newPassword string) (*models.User, error)

	// VerifyPassword validates a user's password.
	VerifyPassword(id int32, password string) (bool, error)

	// FindUserByEmail finds a user by email.
	FindUserByEmail(email string) (*models.User, error)

	// FindUserByID finds a user by ID. This is safe to expose to the API.
	FindUserByID(id int32) (*models.User, error)

	// FindUserByPublicID finds a user by ID. This is safe to expose to the API.
	FindUserByPublicID(id string) (*models.User, error)

	// VerifyEmail verifies an email.
	VerifyEmail(id int32) (*models.User, error)

	// RequestEmailChange requests an email change.
	RequestEmailChange(id int32, newEmail string) (code string, err error)

	// VerifyEmailChange verifies an email change.
	VerifyEmailChange(id int32, newEmail, code string) error

	// UpdateProfile updates a user's profile.
	UpdateProfile(id int32, data UpdateProfileParams) (*models.User, error)
}

type userRepo struct {
	*baseRepo
}

// RequestEmailChange implements UserRepository.
func (u *userRepo) RequestEmailChange(id int32, newEmail string) (code string, err error) {
	eg := new(errgroup.Group)

	// Check if the email is already in use
	eg.Go(func() (err error) {
		if emailExists, _ := u.EmailExists(newEmail); emailExists {
			return apperrors.New(
				"email is already in use by another account",
				http.StatusConflict,
			)
		}

		return nil
	})

	// Check if the new email has already been reserved
	eg.Go(func() (err error) {
		hasBeenReserved, err := u.store.Exists(kv.KeyEmailReservation(newEmail))
		if err != nil {
			return err
		}

		if !hasBeenReserved {
			return nil
		}

		// If the email has already been reserved, we need to check if it was reserved by the same user
		reservedBy, err := u.store.Get(kv.KeyEmailReservation(newEmail))
		if err != nil {
			return err
		}

		if string(reservedBy) == fmt.Sprintf("%d", id) {
			return nil
		}

		return apperrors.New(
			"unable to use this email at this time, this email is most likely reserved by another user",
			http.StatusConflict,
		)
	})

	if err := eg.Wait(); err != nil {
		return "", err
	}

	// ==== Generate a new email change code ==== //
	eg = new(errgroup.Group)

	// Generate a new email change code
	eg.Go(func() (err error) {
		code, err = u.otpManager.GenerateToken(otp.TokenUserEmailChange, id)
		if err != nil {
			return err
		}

		return nil
	})

	// We will store the user who requested the email change to prevent two users from trying to change to the same email
	eg.Go(func() (err error) {
		if exists, _ := u.store.Exists(kv.KeyEmailReservation(newEmail)); exists {
			return nil
		}

		key := kv.KeyEmailReservation(newEmail)
		if err := u.store.SetWithTTL(key, fmt.Append(nil, id), time.Hour); err != nil {
			return err
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return "", err
	}

	return code, nil
}

// VerifyEmailChange implements UserRepository.
func (u *userRepo) VerifyEmailChange(id int32, newEmail, code string) error {
	eg := new(errgroup.Group)

	// Check if the email is already in use
	eg.Go(func() (err error) {
		if emailExists, _ := u.EmailExists(newEmail); emailExists {
			return apperrors.New(
				"email is already in use by another account",
				http.StatusConflict,
			)
		}

		return nil
	})

	// Check if the new email has already been reserved
	eg.Go(func() (err error) {
		hasBeenReserved, err := u.store.Exists(kv.KeyEmailReservation(newEmail))
		if err != nil {
			return err
		}

		if !hasBeenReserved {
			return apperrors.New(
				"no email change request found for this email",
				http.StatusNotFound,
			)
		}

		// If the email has already been reserved, we need to check if it was reserved by the same user
		reservedBy, err := u.store.Get(kv.KeyEmailReservation(newEmail))
		if err != nil {
			return err
		}

		if string(reservedBy) == fmt.Sprintf("%d", id) {
			return nil
		}

		return apperrors.New(
			"unable to use this email at this time, this email is most likely reserved by another user",
			http.StatusConflict,
		)
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	// ==== Verify the email change code ==== //
	if err := u.otpManager.VerifyToken(otp.VerifyTokenParams{
		Type:                otp.TokenUserEmailChange,
		Identifier:          id,
		ProvidedToken:       code,
		RetainAfterVerified: false,
	}); err != nil {
		return err
	}

	eg = new(errgroup.Group)

	// Update the user's email
	eg.Go(func() (err error) {
		if _, err := u.queries.UpdateEmail(context.TODO(), queries.UpdateEmailParams{
			Email: newEmail,
			ID:    id,
		}); err != nil {
			return err
		}

		return nil
	})

	// Remove the email reservation
	eg.Go(func() (err error) {
		if err := u.store.Delete(kv.KeyEmailReservation(newEmail)); err != nil {
			return err
		}

		return nil
	})

	return eg.Wait()
}

// EmailExists checks if an email already exists.
func (u *userRepo) EmailExists(email string) (bool, error) {
	if lib.Empty(email) {
		return false, apperrors.New("email is required", http.StatusBadRequest)
	}

	return u.queries.UserExistsByEmail(context.TODO(), email)
}

// UsernameExists checks if a username already exists.
func (u *userRepo) UsernameExists(username string) (bool, error) {
	if lib.Empty(username) {
		return false, apperrors.New("username is required", http.StatusBadRequest)
	}

	return u.queries.UserExistsByUsername(context.TODO(), username)
}

// CreateUser creates a new user account.
func (u *userRepo) CreateUser(data queries.CreateUserParams) (*models.User, error) {
	var (
		usernameExists, emailExists bool
		err                         error
	)

	if usernameExists, err = u.UsernameExists(data.Username); err != nil {
		return nil, err
	}

	if usernameExists {
		return nil, apperrors.New(
			"an account with this username already exists",
			http.StatusConflict,
		)
	}

	if emailExists, err = u.EmailExists(data.Email); err != nil {
		return nil, err
	}

	if emailExists {
		return nil, apperrors.New("an account with this email already exists", http.StatusConflict)
	}

	user, err := u.queries.CreateUser(context.TODO(), data)
	if err != nil {
		return nil, err
	}

	return ToUser(&user), nil
}

func (u *userRepo) FindUserByID(id int32) (*models.User, error) {
	user, err := u.queries.FindUserById(context.TODO(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrAccountWithIDNotFound
		}

		return nil, err
	}

	return ToUser(&user), nil
}

// FindUserByEmail finds a user by email.
func (u *userRepo) FindUserByEmail(email string) (*models.User, error) {
	user, err := u.queries.FindUserByEmail(context.TODO(), email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrAccountWithEmailNotFound
		}

		return nil, err
	}

	return ToUser(&user), nil
}

func (u *userRepo) FindUserByPublicID(publicId string) (*models.User, error) {
	id := new(pgtype.UUID)

	if err := id.Scan(publicId); err != nil {
		return nil, seer.Wrap("find_user_by_public_id", err, "invalid user ID provided")
	}

	user, err := u.queries.FindUserByPublicId(context.TODO(), pgtype.UUID{
		Bytes: id.Bytes,
		Valid: true,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrAccountWithIDNotFound
		}

		return nil, err
	}

	return ToUser(&user), nil
}

// ChangePassword changes a user's password.
func (u *userRepo) ChangePassword(id int32, newPassword string) (*models.User, error) {
	argon := argon2.DefaultConfig()
	hashedPassword, err := argon.HashEncoded([]byte(newPassword))
	if err != nil {
		return nil, seer.Wrap("change_password", err)
	}

	user, err := u.queries.UpdatePassword(context.TODO(), queries.UpdatePasswordParams{
		ID:             id,
		HashedPassword: string(hashedPassword),
	})
	if err != nil {
		return nil, err
	}

	return ToUser(&user), nil
}

// VerifyPassword validates a user's password.
func (u *userRepo) VerifyPassword(id int32, password string) (bool, error) {
	user, err := u.queries.FindUserById(context.TODO(), id)
	if err != nil {
		return false, err
	}

	return lib.VerifyPassword(lib.VerifyPasswordParams{
		Password: password,
		Hash:     user.HashedPassword,
	})
}

// VerifyEmail verifies an email.
func (u *userRepo) VerifyEmail(id int32) (*models.User, error) {
	user, err := u.queries.VerifyEmail(context.TODO(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &models.User{}, apperrors.New(
				"no account with this ID found",
				http.StatusNotFound,
			)
		}

		return &models.User{}, err
	}

	return ToUser(&user), nil
}

func (u *userRepo) UpdateProfile(id int32, data UpdateProfileParams) (*models.User, error) {
	user, err := u.queries.UpdateProfile(context.TODO(), queries.UpdateProfileParams{
		ID:        id,
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Username:  data.Username,
	})
	if err != nil {
		return nil, err
	}

	return ToUser(&user), nil
}

func ToUser(u *queries.User) *models.User {
	return &models.User{
		ID:             u.ID,
		PublicID:       u.PublicID.String(),
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		HashedPassword: u.HashedPassword,
		Email:          u.Email,
		EmailVerified:  u.EmailVerified,
		Username:       u.Username,
		AvatarID:       u.AvatarID.String,
		CreatedAt:      u.CreatedAt.Time,
		LastUpdatedAt:  u.UpdatedAt.Time,
	}
}
