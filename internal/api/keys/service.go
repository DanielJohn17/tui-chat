package keys

import (
	"context"
	"encoding/hex"

	"github.com/DanielJohn17/tui-chat/internal/api/apierrs"
	"github.com/DanielJohn17/tui-chat/internal/api/users"
)

type KeyServiceInt interface {
	UploadPublicKey(ctx context.Context, userID int64, hexKey string) error
	GetPublicKey(ctx context.Context, userID int64) (*PublicKeyResponse, error)
}

type KeyService struct {
	u users.UserServiceInt
}

func NewKeyService(u users.UserServiceInt) *KeyService {
	return &KeyService{u: u}
}

var _ KeyServiceInt = (*KeyService)(nil)

func (s *KeyService) UploadPublicKey(ctx context.Context, userID int64, hexKey string) error {
	bytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return apierrs.NewBadRequestError("public key must be a valid hex-encoded 32-byte value")
	}
	return s.u.SetPublicKey(ctx, userID, bytes)
}

func (s *KeyService) GetPublicKey(ctx context.Context, userID int64) (*PublicKeyResponse, error) {
	pkBytes, err := s.u.GetPublicKey(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &PublicKeyResponse{
		UserID:    userID,
		PublicKey: hex.EncodeToString(pkBytes),
	}, nil
}
