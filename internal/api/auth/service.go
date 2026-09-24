// Package auth
package auth

import (
	"context"

	"github.com/DanielJohn17/tui-chat/internal/api/errors"
	"github.com/DanielJohn17/tui-chat/internal/api/helpers"
	"github.com/DanielJohn17/tui-chat/internal/api/users"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceInt interface {
	Register(
		ctx context.Context,
		user RegisterUserType,
	) (*UserResponseType, error)

	Login(
		ctx context.Context,
		input LoginUserType,
	) (*UserResponseType, error)
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
) (*UserResponseType, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, errors.NewInternalServerError("password error, try again", err)
	}

	createUser := users.CreateUserType{
		Name:     input.Name,
		Username: input.Username,
		Password: string(hashPassword),
	}

	userRegisterd, apiError := s.u.CreateUser(ctx, createUser)
	if apiError != nil {
		return nil, apiError
	}

	token, err := helpers.CreateToken(
		helpers.UserToken{ID: userRegisterd.ID, Username: userRegisterd.Username},
	)
	if err != nil {
		return nil, errors.NewInternalServerError("error creating token", err)
	}

	newData := &UserResponseType{
		ID:        userRegisterd.ID,
		Name:      userRegisterd.Name,
		Username:  userRegisterd.Username,
		Token:     token,
		CreatedAt: userRegisterd.CreatedAt,
		UpdatedAt: userRegisterd.UpdatedAt,
	}
	return newData, nil
}

func (s *AuthService) Login(
	ctx context.Context,
	input LoginUserType,
) (*UserResponseType, error) {
	user, err := s.u.GetUserByUsername(ctx, input.Username)
	if err != nil {
		return nil, errors.NewNotFoundError("incorrect username or password")
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(input.Password),
	); err != nil {
		return nil, errors.NewNotFoundError("incorrect username or password")
	}

	token, err := helpers.CreateToken(helpers.UserToken{ID: user.ID, Username: user.Username})
	if err != nil {
		return nil, errors.NewInternalServerError("error creating token", err)
	}

	userLogin := &UserResponseType{
		ID:       user.ID,
		Name:     user.Name,
		Username: user.Username,
		Token:    token,
	}

	return userLogin, nil
}
