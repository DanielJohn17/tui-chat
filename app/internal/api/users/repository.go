package users

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DanielJohn17/tui-chat/app/internal/api/database"
)

type UserRepositoryInt interface {
	CreateUser(ctx context.Context, user CreateUserType) (*CreateUserResponseType, error)
	GetUserByUsername(ctx context.Context, username string) (*GetUserResponseAuthType, error)
	GetUserByID(ctx context.Context, id int64) (*GetUserResponseType, error)
	DeleteUser(ctx context.Context, id int64) (*DeleteUserResponseType, error)
}

type UserQuerier interface {
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.CreateUserRow, error)
	GetUserByUsername(ctx context.Context, username string) (database.GetUserByUsernameRow, error)
	GetUserById(ctx context.Context, id int64) (database.GetUserByIdRow, error)
	DeleteUser(ctx context.Context, id int64) (database.DeleteUserRow, error)
}

type UserRepository struct {
	q UserQuerier
}

func NewUserRepository(q UserQuerier) *UserRepository {
	return &UserRepository{q: q}
}

var _ UserRepositoryInt = (*UserRepository)(nil)

func (r *UserRepository) CreateUser(
	ctx context.Context,
	user CreateUserType,
) (*CreateUserResponseType, error) {
	userParams := database.CreateUserParams{
		Name:     user.Name,
		Username: user.Username,
		Password: user.Password,
	}

	createdUser, err := r.q.CreateUser(ctx, userParams)
	if err != nil {
		log.Printf("==>CreateUser: %v\n", err)
		return nil, fmt.Errorf("error creating user")
	}

	return &CreateUserResponseType{
		ID:        createdUser.ID,
		Name:      createdUser.Name,
		Username:  createdUser.Username,
		CreatedAt: createdUser.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: createdUser.UpdatedAt.Time.Format(time.RFC3339),
	}, nil
}

func (r *UserRepository) GetUserByUsername(
	ctx context.Context,
	username string,
) (*GetUserResponseAuthType, error) {
	user, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		log.Printf("==>GetUserByUsername: %v\n", err)
		return nil, fmt.Errorf("user not found")
	}

	return &GetUserResponseAuthType{
		ID:       user.ID,
		Name:     user.Name,
		Password: user.Password,
		Username: user.Username,
	}, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*GetUserResponseType, error) {
	user, err := r.q.GetUserById(ctx, id)
	if err != nil {
		log.Printf("=>>GetUserById: %v\n", err)
		return nil, fmt.Errorf("user not found")
	}

	return &GetUserResponseType{
		ID:        user.ID,
		Name:      user.Name,
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Time.Format(time.RFC3339),
	}, nil
}

func (r *UserRepository) DeleteUser(
	ctx context.Context,
	id int64,
) (*DeleteUserResponseType, error) {
	deletedUser, err := r.q.DeleteUser(ctx, id)
	if err != nil {
		log.Printf("==>DeleteUser: %v\n", err)
		return nil, fmt.Errorf("error deleting user")
	}

	return &DeleteUserResponseType{
		ID:       deletedUser.ID,
		Name:     deletedUser.Name,
		Username: deletedUser.Username,
	}, nil
}
