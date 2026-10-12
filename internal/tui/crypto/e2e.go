package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// Sentinel errors returned by Decrypt.
var (
	// ErrDecryptFailed indicates the ciphertext could not be decrypted.
	// This can happen if the key is wrong or the data is corrupted.
	ErrDecryptFailed = errors.New("e2ee: decryption failed")

	// ErrLegacyMessage indicates the content appears to be a legacy plaintext
	// message that was stored before E2EE was enabled.
	ErrLegacyMessage = errors.New("e2ee: legacy plaintext message")
)

// hkdfInfo is the application-specific context label used in HKDF.
// Changing this value invalidates all existing derived keys.
const hkdfInfo = "tui-chat-e2ee-v1"

// DeriveConvKey computes the shared AES-256 symmetric key for a direct conversation
// using X25519 ECDH and HKDF-SHA256.
//
// Both conversation participants derive the same key:
//
//	Alice: DeriveConvKey(alicePriv, bobPub)  == key
//	Bob:   DeriveConvKey(bobPriv,   alicePub) == key
func DeriveConvKey(myPriv [32]byte, peerPub [32]byte) ([32]byte, error) {
	priv, err := ecdh.X25519().NewPrivateKey(myPriv[:])
	if err != nil {
		return [32]byte{}, fmt.Errorf("e2ee: invalid private key: %w", err)
	}

	pub, err := ecdh.X25519().NewPublicKey(peerPub[:])
	if err != nil {
		return [32]byte{}, fmt.Errorf("e2ee: invalid peer public key: %w", err)
	}

	// Perform ECDH
	shared, err := priv.ECDH(pub)
	if err != nil {
		return [32]byte{}, fmt.Errorf("e2ee: ECDH failed: %w", err)
	}

	// Expand to AES-256 key with HKDF-SHA256
	h := hkdf.New(sha256.New, shared, nil, []byte(hkdfInfo))
	var key [32]byte
	if _, err := io.ReadFull(h, key[:]); err != nil {
		return [32]byte{}, fmt.Errorf("e2ee: HKDF expansion failed: %w", err)
	}

	return key, nil
}

// Encrypt encrypts plaintext with AES-256-GCM using a fresh random 12-byte nonce.
// The returned string is base64(nonce || ciphertext || GCM-tag), safe for JSON transport
// and storage in the messages.content column.
func Encrypt(key [32]byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("e2ee: failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("e2ee: failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize()) // 12 bytes
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("e2ee: failed to generate nonce: %w", err)
	}

	// Seal appends GCM tag to ciphertext: nonce || ciphertext+tag
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a ciphertext string produced by Encrypt.
// Returns the plaintext on success.
//
// Error cases:
//   - ErrLegacyMessage: content is not valid base64 (likely a pre-E2EE plaintext message)
//   - ErrDecryptFailed: valid base64 but GCM authentication failed (wrong key or tampered)
func Decrypt(key [32]byte, ciphertext string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		// Not base64 → it's a legacy plaintext message stored before E2EE.
		return "", ErrLegacyMessage
	}

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", ErrDecryptFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", ErrDecryptFailed
	}

	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize+gcm.Overhead() {
		// Too short to be a valid ciphertext — treat as legacy.
		return "", ErrLegacyMessage
	}

	nonce, ciphertextBytes := raw[:nonceSize], raw[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", ErrDecryptFailed
	}

	return string(plaintext), nil
}
