// Package users
package users

type URLParam struct {
	Username string `uri:"username" binding:"required,alphanum,min=3,max=20"`
}

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
	UpdatedAt string
}

type GetUserResponseType struct {
	ID        int64
	Name      string
	Username  string
	CreatedAt string
	UpdatedAt string
}

type DeleteUserResponseType struct {
	ID       int64
	Name     string
	Username string
}
