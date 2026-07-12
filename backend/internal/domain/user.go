package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleAdmin    UserRole = "admin"
)

func (r UserRole) IsValid() bool {
	switch r {
	case RoleCustomer, RoleAdmin:
		return true
	default:
		return false
	}
}

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type User struct {
	id        string
	name      string
	email     string
	avatarURL *string
	role      UserRole

	emailVerifiedAt *time.Time
	lastLoginAt     *time.Time

	Timestamps
	SoftDelete
	Activatable
}

func NewUser(name, email string, role UserRole, avatarURL *string) (*User, error) {
	if err := validateUser(name, email, role); err != nil {
		return nil, err
	}

	user := &User{}

	user.setData(name, email, role, avatarURL)

	return user, nil
}

func RestoreUser(
	id, name, email string,
	avatarURL *string,
	role UserRole,
	active bool,
	emailVerifiedAt, lastLoginAt *time.Time,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) (*User, error) {
	if err := validateUser(name, email, role); err != nil {
		return nil, err
	}

	return &User{
		id:              id,
		name:            strings.TrimSpace(name),
		email:           strings.TrimSpace(email),
		avatarURL:       avatarURL,
		role:            role,
		emailVerifiedAt: emailVerifiedAt,
		lastLoginAt:     lastLoginAt,
		Timestamps:      NewTimestampsFrom(createdAt, updatedAt),
		SoftDelete:      NewSoftDeleteFrom(deletedAt),
		Activatable:     NewActivatableFrom(active),
	}, nil
}

func (u *User) setData(name, email string, role UserRole, avatarURL *string) {
	u.name = strings.TrimSpace(name)
	u.email = strings.TrimSpace(email)
	u.role = role
	u.avatarURL = avatarURL
}

func validateUser(name, email string, role UserRole) error {
	var errs []error

	if strings.TrimSpace(name) == "" {
		errs = append(errs, ErrInvalidUserName)
	}
	if !emailRegex.MatchString(strings.TrimSpace(email)) {
		errs = append(errs, ErrInvalidUserEmail)
	}
	if !role.IsValid() {
		errs = append(errs, ErrInvalidUserRole)
	}

	return errors.Join(errs...)
}

func (u *User) Update(name, email string, role UserRole, avatarURL *string) error {
	if err := validateUser(name, email, role); err != nil {
		return err
	}

	u.setData(name, email, role, avatarURL)

	return nil
}

func (u *User) ID() string {
	return u.id
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Email() string {
	return u.email
}

func (u *User) AvatarURL() *string {
	return u.avatarURL
}

func (u *User) Role() UserRole {
	return u.role
}

func (u *User) EmailVerifiedAt() *time.Time {
	return u.emailVerifiedAt
}

func (u *User) LastLoginAt() *time.Time {
	return u.lastLoginAt
}
