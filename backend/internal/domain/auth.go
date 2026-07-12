package domain

import "time"

type AuthenticatedUser struct {
	ID    string
	Name  string
	Email string
	Role  UserRole
}

type UserAuthentication struct {
	UserID            string
	Name              string
	Email             string
	Role              UserRole
	Active            bool
	DeletedAt         *time.Time
	PasswordHash      string
	PasswordChangedAt time.Time
}

func (a *UserAuthentication) IsDeleted() bool {
	return a.DeletedAt != nil
}
