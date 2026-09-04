package users

import (
	"context"
	"net/http"

	"github.com/DanielJohn17/tui-chat/app/internal/api/types"
)

type UserServiceInt interface {
	CreateUser(
		ctx context.Context,
		input CreateUserType,
	) (*types.APIResponse[CreateUserResponseType], *types.APIErrorResponse)

	GetUserByUsername(
		ctx context.Context,
		input string,
	) (*types.APIResponse[GetUserResponseType], *types.APIErrorResponse)

	GetUserByID(
		ctx context.Context,
		input int64,
	) (*types.APIResponse[GetUserResponseType], *types.APIErrorResponse)

	DeleteUser(
		ctx context.Context,
		input int64,
	) (*types.APIResponse[DeleteUserResponseType], *types.APIErrorResponse)
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
) (*types.APIResponse[CreateUserResponseType], *types.APIErrorResponse) {
	_, err := s.r.GetUserByUsername(ctx, input.Username)
	if err == nil {
		return nil, &types.APIErrorResponse{
			Status:  http.StatusConflict,
			Success: false,
			Message: "user already created",
		}
	}

	createdUser, err := s.r.CreateUser(ctx, input)
	if err != nil {
		return nil, &types.APIErrorResponse{
			Status:  http.StatusInternalServerError,
			Success: false,
			Message: err.Error(),
		}
	}

	return &types.APIResponse[CreateUserResponseType]{
		Status:  http.StatusCreated,
		Success: true,
		Data:    *createdUser,
	}, nil
}

func (s *UserService) GetUserByUsername(
	ctx context.Context,
	input string,
) (*types.APIResponse[GetUserResponseType], *types.APIErrorResponse) {
	user, err := s.r.GetUserByUsername(ctx, input)
	if err != nil {
		return nil, &types.APIErrorResponse{
			Status:  http.StatusNotFound,
			Success: false,
			Message: err.Error(),
		}
	}

	return &types.APIResponse[GetUserResponseType]{
		Status:  http.StatusFound,
		Success: true,
		Data:    *user,
	}, nil
}

func (s *UserService) GetUserByID(
	ctx context.Context,
	input int64,
) (*types.APIResponse[GetUserResponseType], *types.APIErrorResponse) {
	user, err := s.r.GetUserByID(ctx, input)
	if err != nil {
		return nil, &types.APIErrorResponse{
			Status:  http.StatusNotFound,
			Success: false,
			Message: err.Error(),
		}
	}

	return &types.APIResponse[GetUserResponseType]{
		Status:  http.StatusFound,
		Success: true,
		Data:    *user,
	}, nil
}

func (s *UserService) DeleteUser(
	ctx context.Context,
	input int64,
) (*types.APIResponse[DeleteUserResponseType], *types.APIErrorResponse) {
	user, err := s.r.GetUserByID(ctx, input)
	if err != nil {
		return nil, &types.APIErrorResponse{
			Status:  http.StatusNotFound,
			Success: false,
			Message: err.Error(),
		}
	}

	deletedUser, err := s.r.DeleteUser(ctx, user.ID)
	if err != nil {
		return nil, &types.APIErrorResponse{
			Status:  http.StatusInternalServerError,
			Success: false,
			Message: err.Error(),
		}

	}

	return &types.APIResponse[DeleteUserResponseType]{
		Status:  http.StatusOK,
		Success: true,
		Data:    *deletedUser,
	}, nil
}
