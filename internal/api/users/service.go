package users

import (
	"context"

	"github.com/DanielJohn17/tui-chat/app/internal/api/errors"
)

type UserServiceInt interface {
	CreateUser(
		ctx context.Context,
		input CreateUserType,
	) (*CreateUserResponseType, error)

	GetUserByUsername(
		ctx context.Context,
		input string,
	) (*GetUserResponseType, error)

	GetUserByID(
		ctx context.Context,
		input int64,
	) (*GetUserResponseType, error)

	DeleteUser(
		ctx context.Context,
		input int64,
	) (*DeleteUserResponseType, error)
}

type UserService struct {
	r UserRepositoryInt
}

func NewUserService(r UserRepositoryInt) *UserService {
	return &UserService{r: r}
}

var _ UserServiceInt = (*UserService)(nil)

func (s *UserService) CreateUser(
	ctx context.Context,
	input CreateUserType,
) (*CreateUserResponseType, error) {
	_, err := s.r.GetUserByUsername(ctx, input.Username)
	if err == nil {
		return nil, errors.NewConflictError("user already exists")
	}

	createdUser, err := s.r.CreateUser(ctx, input)
	if err != nil {
		return nil, errors.NewInternalServerError(err.Error(), err)
	}

	return createdUser, nil
}

func (s *UserService) GetUserByUsername(
	ctx context.Context,
	input string,
) (*GetUserResponseType, error) {
	user, err := s.r.GetUserByUsername(ctx, input)
	if err != nil {
		return nil, errors.NewNotFoundError(err.Error())
	}

	userResp := &GetUserResponseType{
		ID:       user.ID,
		Name:     user.Name,
		Password: user.Password,
		Username: user.Username,
	}

	return userResp, nil
}

func (s *UserService) GetUserByID(
	ctx context.Context,
	input int64,
) (*GetUserResponseType, error) {
	user, err := s.r.GetUserByID(ctx, input)
	if err != nil {
		return nil, errors.NewNotFoundError(err.Error())
	}

	return user, nil
}

func (s *UserService) DeleteUser(
	ctx context.Context,
	input int64,
) (*DeleteUserResponseType, error) {
	user, err := s.r.GetUserByID(ctx, input)
	if err != nil {
		return nil, errors.NewNotFoundError(err.Error())
	}

	deletedUser, err := s.r.DeleteUser(ctx, user.ID)
	if err != nil {
		return nil, errors.NewInternalServerError(err.Error(), err)
	}

	return deletedUser, nil
}
