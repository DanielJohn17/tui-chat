package users

import (
	"context"
	"fmt"
	"time"

	"github.com/DanielJohn17/tui-chat/app/internal/api/database"
)

type UserRepositoryInt interface {
	CreateUser(ctx context.Context, user CreateUser) (*CreateUserResponse, error)
	GetUserByUsername(ctx context.Context, username string) (*GetUserResponse, error)
}

type UserRepository struct {
	q *database.Queries
}

func NewUserRepository(q *database.Queries) *UserRepository {
	return &UserRepository{q: q}
}

var _ UserRepositoryInt = (*UserRepository)(nil)

func (r *UserRepository) CreateUser(ctx context.Context, user CreateUser) (*CreateUserResponse, error) {
	userParams := database.CreateUserParams{
		Name:     user.Name,
		Username: user.Username,
		Password: user.Password,
	}

	createdUser, err := r.q.CreateUser(ctx, userParams)
	if err != nil {
		fmt.Printf("CreateUser: %v\n", err)
		return nil, fmt.Errorf("error creating user")
	}

	return &CreateUserResponse{
		ID:        createdUser.ID,
		Name:      createdUser.Name,
		Username:  createdUser.Username,
		CreatedAt: createdUser.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAT: createdUser.UpdatedAt.Time.Format(time.RFC3339),
	}, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*GetUserResponse, error) {
	user, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		fmt.Printf("GetUserByUsername: %v\n", err)
		return nil, fmt.Errorf("user not found")
	}

	return &GetUserResponse{ID: user.ID, Name: user.Name, Username: user.Username}, nil
}
