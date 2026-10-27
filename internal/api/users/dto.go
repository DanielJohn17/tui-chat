// Package users
package users

type PublicUser struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Bio      string `json:"bio"`
}

type SearchUsersResponse struct {
	Users []PublicUser `json:"users"`
}

type CreateUserType struct {
	Name     string
	Username string
	Password string
}

type CreateUserResponseType struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Username     string `json:"username"`
	Bio          string `json:"bio"`
	IsSearchable bool   `json:"is_searchable"`
	InviteToken  string `json:"invite_token"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type GetUserResponseType struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Username     string `json:"username"`
	Bio          string `json:"bio"`
	IsSearchable bool   `json:"is_searchable"`
	InviteToken  string `json:"invite_token,omitempty"`
	Password     string `json:"-"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type UpdateUserProfileRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=150"`
	Bio          string `json:"bio" validate:"max=255"`
	IsSearchable *bool  `json:"is_searchable" validate:"required"`
}

type DeleteUserResponseType struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}
