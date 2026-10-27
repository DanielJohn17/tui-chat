package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"unicode/utf8"

	"github.com/DanielJohn17/tui-chat/internal/api/apierrs"
)

type UserServiceInt interface {
	SearchUsers(ctx context.Context, userID int64, query string) (*SearchUsersResponse, error)
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

	GetUserByInviteToken(
		ctx context.Context,
		inviteToken string,
	) (*PublicUser, error)

	UpdateProfile(
		ctx context.Context,
		userID int64,
		input UpdateUserProfileRequest,
	) (*GetUserResponseType, error)

	RegenerateInviteToken(
		ctx context.Context,
		userID int64,
	) (*GetUserResponseType, error)

	DeleteUser(
		ctx context.Context,
		input int64,
	) (*DeleteUserResponseType, error)

	SetPublicKey(ctx context.Context, userID int64, pubKey []byte) error
	GetPublicKey(ctx context.Context, userID int64) ([]byte, error)
}

type UserService struct {
	r UserRepositoryInt
}

func NewUserService(r UserRepositoryInt) *UserService {
	return &UserService{r: r}
}

var _ UserServiceInt = (*UserService)(nil)

func (s *UserService) SearchUsers(ctx context.Context, userID int64, query string) (*SearchUsersResponse, error) {
	query = strings.TrimSpace(query)
	if n := utf8.RuneCountInString(query); !utf8.ValidString(query) || n < 2 || n > 20 {
		return nil, apierrs.NewBadRequestError("query must be between 2 and 20 characters")
	}
	result, err := s.r.SearchUsers(ctx, userID, query)
	if err != nil {
		return nil, apierrs.NewInternalServerError("error searching users", err)
	}
	if result == nil {
		result = []PublicUser{}
	}
	return &SearchUsersResponse{Users: result}, nil
}

func (s *UserService) CreateUser(
	ctx context.Context,
	input CreateUserType,
) (*CreateUserResponseType, error) {
	_, err := s.r.GetUserByUsername(ctx, input.Username)
	if err == nil {
		return nil, apierrs.NewConflictError("user already exists")
	}

	createdUser, err := s.r.CreateUser(ctx, input)
	if err != nil {
		return nil, apierrs.NewInternalServerError(err.Error(), err)
	}

	return createdUser, nil
}

func (s *UserService) GetUserByUsername(
	ctx context.Context,
	input string,
) (*GetUserResponseType, error) {
	user, err := s.r.GetUserByUsername(ctx, input)
	if err != nil {
		return nil, apierrs.NewNotFoundError(err.Error())
	}

	return user, nil
}

func (s *UserService) GetUserByID(
	ctx context.Context,
	input int64,
) (*GetUserResponseType, error) {
	user, err := s.r.GetUserByID(ctx, input)
	if err != nil {
		return nil, apierrs.NewNotFoundError(err.Error())
	}

	return user, nil
}

func (s *UserService) GetUserByInviteToken(ctx context.Context, inviteToken string) (*PublicUser, error) {
	inviteToken = strings.TrimSpace(inviteToken)
	if inviteToken == "" {
		return nil, apierrs.NewBadRequestError("invite token is required")
	}
	user, err := s.r.GetUserByInviteToken(ctx, inviteToken)
	if err != nil {
		return nil, apierrs.NewNotFoundError("invalid or expired invite link")
	}
	return &PublicUser{
		ID:       user.ID,
		Name:     user.Name,
		Username: user.Username,
		Bio:      user.Bio,
	}, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, input UpdateUserProfileRequest) (*GetUserResponseType, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apierrs.NewBadRequestError("name cannot be empty")
	}
	if utf8.RuneCountInString(name) > 150 {
		return nil, apierrs.NewBadRequestError("name must be at most 150 characters")
	}
	bio := strings.TrimSpace(input.Bio)
	if utf8.RuneCountInString(bio) > 255 {
		return nil, apierrs.NewBadRequestError("bio must be at most 255 characters")
	}
	if input.IsSearchable == nil {
		return nil, apierrs.NewBadRequestError("is_searchable is required")
	}

	user, err := s.r.UpdateUserProfile(ctx, userID, name, bio, *input.IsSearchable)
	if err != nil {
		return nil, apierrs.NewInternalServerError("failed to update profile", err)
	}
	return user, nil
}

func (s *UserService) RegenerateInviteToken(ctx context.Context, userID int64) (*GetUserResponseType, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, apierrs.NewInternalServerError("failed to generate invite token", err)
	}
	token := hex.EncodeToString(b)

	user, err := s.r.RegenerateInviteToken(ctx, userID, token)
	if err != nil {
		return nil, apierrs.NewInternalServerError("failed to save new invite token", err)
	}
	return user, nil
}

func (s *UserService) DeleteUser(
	ctx context.Context,
	input int64,
) (*DeleteUserResponseType, error) {
	user, err := s.r.GetUserByID(ctx, input)
	if err != nil {
		return nil, apierrs.NewNotFoundError(err.Error())
	}

	deletedUser, err := s.r.DeleteUser(ctx, user.ID)
	if err != nil {
		return nil, apierrs.NewInternalServerError(err.Error(), err)
	}

	return deletedUser, nil
}

func (s *UserService) SetPublicKey(ctx context.Context, userID int64, pubKey []byte) error {
	if len(pubKey) != 32 {
		return apierrs.NewBadRequestError("invalid public key: must be 32 bytes")
	}
	return s.r.SetPublicKey(ctx, userID, pubKey)
}

func (s *UserService) GetPublicKey(ctx context.Context, userID int64) ([]byte, error) {
	pk, err := s.r.GetPublicKey(ctx, userID)
	if err != nil || len(pk) == 0 {
		return nil, apierrs.NewNotFoundError("public key not found")
	}
	return pk, nil
}
