// Package users
package users

type CreateUserType struct {
	Name     string
	Username string
	Password string
}

type CreateUserResponseType struct {
	ID        int64
	Name      string
	Username  string
	CreatedAt string
	UpdatedAT string
}

type GetUserResponseType struct {
	ID       int64
	Name     string
	Username string
}
