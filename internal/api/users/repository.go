package users

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/api/database"
)

type UserRepositoryInt interface {
	SearchUsers(ctx context.Context, userID int64, query string) ([]PublicUser, error)
	CreateUser(ctx context.Context, user CreateUserType) (*CreateUserResponseType, error)
	GetUserByUsername(ctx context.Context, username string) (*GetUserResponseType, error)
	GetUserByID(ctx context.Context, id int64) (*GetUserResponseType, error)
	GetUserByInviteToken(ctx context.Context, inviteToken string) (*GetUserResponseType, error)
	UpdateUserProfile(ctx context.Context, id int64, name, bio string, isSearchable bool) (*GetUserResponseType, error)
	RegenerateInviteToken(ctx context.Context, id int64, token string) (*GetUserResponseType, error)
	DeleteUser(ctx context.Context, id int64) (*DeleteUserResponseType, error)
	SetPublicKey(ctx context.Context, userID int64, pubKey []byte) error
	GetPublicKey(ctx context.Context, userID int64) ([]byte, error)
}

type UserQuerier interface {
	SearchUsers(ctx context.Context, arg database.SearchUsersParams) ([]database.SearchUsersRow, error)
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.CreateUserRow, error)
	GetUserByUsername(ctx context.Context, username string) (database.GetUserByUsernameRow, error)
	GetUserById(ctx context.Context, id int64) (database.GetUserByIdRow, error)
	GetUserByInviteToken(ctx context.Context, inviteToken string) (database.GetUserByInviteTokenRow, error)
	UpdateUserProfile(ctx context.Context, arg database.UpdateUserProfileParams) (database.UpdateUserProfileRow, error)
	RegenerateInviteToken(ctx context.Context, arg database.RegenerateInviteTokenParams) (database.RegenerateInviteTokenRow, error)
	DeleteUser(ctx context.Context, id int64) (database.DeleteUserRow, error)
	SetUserPublicKey(ctx context.Context, arg database.SetUserPublicKeyParams) error
	GetUserPublicKey(ctx context.Context, id int64) ([]byte, error)
}

type UserRepository struct {
	q UserQuerier
}

func NewUserRepository(q UserQuerier) *UserRepository {
	return &UserRepository{q: q}
}

var _ UserRepositoryInt = (*UserRepository)(nil)

func (r *UserRepository) SearchUsers(ctx context.Context, userID int64, query string) ([]PublicUser, error) {
	pattern := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(query) + "%"
	rows, err := r.q.SearchUsers(ctx, database.SearchUsersParams{
		UserID:  userID,
		Pattern: pattern,
		Query:   query,
	})
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	result := make([]PublicUser, 0, len(rows))
	for _, row := range rows {
		result = append(result, PublicUser{
			ID:       row.ID,
			Name:     row.Name,
			Username: row.Username,
			Bio:      row.Bio,
		})
	}
	return result, nil
}

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
		ID:           createdUser.ID,
		Name:         createdUser.Name,
		Username:     createdUser.Username,
		Bio:          createdUser.Bio,
		IsSearchable: createdUser.IsSearchable,
		InviteToken:  createdUser.InviteToken,
		CreatedAt:    createdUser.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    createdUser.UpdatedAt.Time.Format(time.RFC3339),
	}, nil
}

func (r *UserRepository) GetUserByUsername(
	ctx context.Context,
	username string,
) (*GetUserResponseType, error) {
	user, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		log.Printf("==>GetUserByUsername: %v\n", err)
		return nil, fmt.Errorf("user not found")
	}

	return &GetUserResponseType{
		ID:           user.ID,
		Name:         user.Name,
		Password:     user.Password,
		Username:     user.Username,
		Bio:          user.Bio,
		IsSearchable: user.IsSearchable,
		InviteToken:  user.InviteToken,
		CreatedAt:    user.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    user.UpdatedAt.Time.Format(time.RFC3339),
	}, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*GetUserResponseType, error) {
	user, err := r.q.GetUserById(ctx, id)
	if err != nil {
		log.Printf("=>>GetUserById: %v\n", err)
		return nil, fmt.Errorf("user not found")
	}

	return &GetUserResponseType{
		ID:           user.ID,
		Name:         user.Name,
		Username:     user.Username,
		Bio:          user.Bio,
		IsSearchable: user.IsSearchable,
		InviteToken:  user.InviteToken,
		CreatedAt:    user.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    user.UpdatedAt.Time.Format(time.RFC3339),
	}, nil
}

func (r *UserRepository) GetUserByInviteToken(ctx context.Context, inviteToken string) (*GetUserResponseType, error) {
	user, err := r.q.GetUserByInviteToken(ctx, inviteToken)
	if err != nil {
		log.Printf("=>>GetUserByInviteToken: %v\n", err)
		return nil, fmt.Errorf("user not found")
	}

	return &GetUserResponseType{
		ID:           user.ID,
		Name:         user.Name,
		Username:     user.Username,
		Bio:          user.Bio,
		IsSearchable: user.IsSearchable,
		InviteToken:  user.InviteToken,
		CreatedAt:    user.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    user.UpdatedAt.Time.Format(time.RFC3339),
	}, nil
}

func (r *UserRepository) UpdateUserProfile(
	ctx context.Context,
	id int64,
	name string,
	bio string,
	isSearchable bool,
) (*GetUserResponseType, error) {
	updated, err := r.q.UpdateUserProfile(ctx, database.UpdateUserProfileParams{
		ID:           id,
		Name:         name,
		Bio:          bio,
		IsSearchable: isSearchable,
	})
	if err != nil {
		log.Printf("=>>UpdateUserProfile: %v\n", err)
		return nil, fmt.Errorf("error updating profile")
	}

	return &GetUserResponseType{
		ID:           updated.ID,
		Name:         updated.Name,
		Username:     updated.Username,
		Bio:          updated.Bio,
		IsSearchable: updated.IsSearchable,
		InviteToken:  updated.InviteToken,
		CreatedAt:    updated.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    updated.UpdatedAt.Time.Format(time.RFC3339),
	}, nil
}

func (r *UserRepository) RegenerateInviteToken(
	ctx context.Context,
	id int64,
	token string,
) (*GetUserResponseType, error) {
	updated, err := r.q.RegenerateInviteToken(ctx, database.RegenerateInviteTokenParams{
		ID:          id,
		InviteToken: token,
	})
	if err != nil {
		log.Printf("=>>RegenerateInviteToken: %v\n", err)
		return nil, fmt.Errorf("error regenerating invite token")
	}

	return &GetUserResponseType{
		ID:           updated.ID,
		Name:         updated.Name,
		Username:     updated.Username,
		Bio:          updated.Bio,
		IsSearchable: updated.IsSearchable,
		InviteToken:  updated.InviteToken,
		CreatedAt:    updated.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    updated.UpdatedAt.Time.Format(time.RFC3339),
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

func (r *UserRepository) SetPublicKey(ctx context.Context, userID int64, pubKey []byte) error {
	return r.q.SetUserPublicKey(ctx, database.SetUserPublicKeyParams{ID: userID, PublicKey: pubKey})
}

func (r *UserRepository) GetPublicKey(ctx context.Context, userID int64) ([]byte, error) {
	return r.q.GetUserPublicKey(ctx, userID)
}
