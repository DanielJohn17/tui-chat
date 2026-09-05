// Package users
package users

type CreateUserType struct {
	Name     string
	Username string
	Password string
}

type CreateUserResponseType struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type GetUserResponseAuthType struct {
	ID        int64
	Name      string
	Username  string
	Password  string
	CreatedAt string
	UpdatedAt string
}

type GetUserResponseType struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type DeleteUserResponseType struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}
