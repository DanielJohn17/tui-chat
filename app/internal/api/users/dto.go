// Package users
package users

type CreateUser struct {
	Name     string
	Username string
	Password string
}

type CreateUserResponse struct {
	ID        int64
	Name      string
	Username  string
	CreatedAt string
	UpdatedAT string
}

type GetUserResponse struct {
	ID       int64
	Name     string
	Username string
}
