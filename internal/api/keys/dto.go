// Package keys handles public key exchange for E2EE.
package keys

type UploadPublicKeyRequest struct {
	PublicKey string `json:"public_key" binding:"required"`
}

type PublicKeyResponse struct {
	UserID    int64  `json:"user_id"`
	PublicKey string `json:"public_key"`
}
