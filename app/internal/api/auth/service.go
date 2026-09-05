// Package auth
package auth

import (
	"context"
	"net/http"

	"github.com/DanielJohn17/tui-chat/app/internal/api/types"
	"github.com/DanielJohn17/tui-chat/app/internal/api/users"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceInt interface {
	Register(
		ctx context.Context,
		user RegisterUserType,
	) (*types.APIResponse[UserResponseType], *types.APIErrorResponse)

	Login(
		ctx context.Context,
		input LoginUserType,
	) (*types.APIResponse[UserResponseType], *types.APIErrorResponse)
}

type AuthService struct {
	u users.UserServiceInt
}

func NewAuthService(u users.UserServiceInt) *AuthService {
	return &AuthService{u: u}
}

var _ AuthServiceInt = (*AuthService)(nil)

func (s *AuthService) Register(ctx context.Context,
	input RegisterUserType,
) (*types.APIResponse[UserResponseType], *types.APIErrorResponse) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, &types.APIErrorResponse{
			Status:  http.StatusInternalServerError,
			Success: false,
			Message: "password error, try again",
		}
	}

	createUser := users.CreateUserType{
		Name:     input.Name,
		Username: input.Username,
		Password: string(hashPassword),
	}

	apiResponse, apiError := s.u.CreateUser(ctx, createUser)
	if apiError != nil {
		return nil, apiError
	}

	data := apiResponse.Data
	newData := UserResponseType{
		ID:        data.ID,
		Name:      data.Name,
		Username:  data.Username,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
	}
	return &types.APIResponse[UserResponseType]{
		Status:  apiResponse.Status,
		Success: true,
		Data:    newData,
	}, nil
}

func (s *AuthService) Login(
	ctx context.Context,
	input LoginUserType,
) (*types.APIResponse[UserResponseType], *types.APIErrorResponse) {
	user, err := s.u.GetUserAuth(ctx, input.Username)
	if err != nil {
		return nil, &types.APIErrorResponse{
			Status:  404,
			Success: false,
			Message: "incorrect username or password",
		}
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(input.Password),
		[]byte(user.Password),
	); err != nil {
		return nil, &types.APIErrorResponse{
			Status:  404,
			Success: false,
			Message: "incorrect username or password",
		}
	}

	data := UserResponseType{
		ID:       user.ID,
		Name:     user.Name,
		Username: user.Username,
	}

	return &types.APIResponse[UserResponseType]{
		Status:  http.StatusFound,
		Success: true,
		Data:    data,
	}, nil
}
