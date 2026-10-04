package crypto

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestPair is a helper that returns a fresh keypair for test use.
func generateTestPair(t *testing.T) *IdentityKey {
	t.Helper()
	key, err := GenerateIdentityKey()
	require.NoError(t, err)
	return key
}

func TestDeriveConvKey_Symmetric(t *testing.T) {
	alice := generateTestPair(t)
	bob := generateTestPair(t)

	keyAlice, err := DeriveConvKey(alice.PrivateKey, bob.PublicKey)
	require.NoError(t, err)

	keyBob, err := DeriveConvKey(bob.PrivateKey, alice.PublicKey)
	require.NoError(t, err)

	// Both sides must derive the same key
	assert.Equal(t, keyAlice, keyBob, "ECDH must be symmetric: Alice and Bob derive same key")
}

func TestDeriveConvKey_DifferentPeers(t *testing.T) {
	alice := generateTestPair(t)
	bob := generateTestPair(t)
	charlie := generateTestPair(t)

	keyAB, err := DeriveConvKey(alice.PrivateKey, bob.PublicKey)
	require.NoError(t, err)

	keyAC, err := DeriveConvKey(alice.PrivateKey, charlie.PublicKey)
	require.NoError(t, err)

	// Keys for different conversations must differ
	assert.NotEqual(t, keyAB, keyAC, "keys for different conversations must differ")
}

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	alice := generateTestPair(t)
	bob := generateTestPair(t)

	keyAlice, err := DeriveConvKey(alice.PrivateKey, bob.PublicKey)
	require.NoError(t, err)
	keyBob, err := DeriveConvKey(bob.PrivateKey, alice.PublicKey)
	require.NoError(t, err)

	tests := []string{
		"hello, world",
		"",
		"🔒 Unicode safe emoji test ✅",
		"A very long message: " + string(make([]byte, 4096)),
		"special chars: <>&\"'\\n\\t",
	}

	for _, plaintext := range tests {
		ciphertext, err := Encrypt(keyAlice, plaintext)
		require.NoError(t, err, "Encrypt must not fail")
		assert.NotEmpty(t, ciphertext, "ciphertext must not be empty")
		assert.NotEqual(t, plaintext, ciphertext, "ciphertext must differ from plaintext")

		decrypted, err := Decrypt(keyBob, ciphertext)
		require.NoError(t, err, "Decrypt must not fail")
		assert.Equal(t, plaintext, decrypted, "decrypted text must match original")
	}
}

func TestEncrypt_FreshNonceEachCall(t *testing.T) {
	alice := generateTestPair(t)
	bob := generateTestPair(t)
	key, _ := DeriveConvKey(alice.PrivateKey, bob.PublicKey)

	c1, _ := Encrypt(key, "same plaintext")
	c2, _ := Encrypt(key, "same plaintext")

	// Ciphertexts from different calls must differ (different nonces)
	assert.NotEqual(t, c1, c2, "each Encrypt call must produce a distinct ciphertext")
}

func TestDecrypt_LegacyPlaintext(t *testing.T) {
	alice := generateTestPair(t)
	bob := generateTestPair(t)
	key, _ := DeriveConvKey(alice.PrivateKey, bob.PublicKey)

	// A pre-E2EE plaintext message is not valid base64
	legacyMessages := []string{
		"hey whats up",
		"Hello World!",
		"This is a message from before E2EE was enabled.",
	}

	for _, msg := range legacyMessages {
		_, err := Decrypt(key, msg)
		assert.True(t, errors.Is(err, ErrLegacyMessage),
			"legacy plaintext %q must return ErrLegacyMessage, got: %v", msg, err)
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	alice := generateTestPair(t)
	bob := generateTestPair(t)
	charlie := generateTestPair(t)

	// Alice encrypts for Bob
	keyAB, _ := DeriveConvKey(alice.PrivateKey, bob.PublicKey)
	ciphertext, err := Encrypt(keyAB, "secret message")
	require.NoError(t, err)

	// Charlie tries to decrypt with his key — must fail
	keyBC, _ := DeriveConvKey(bob.PrivateKey, charlie.PublicKey)
	_, err = Decrypt(keyBC, ciphertext)
	assert.True(t, errors.Is(err, ErrDecryptFailed),
		"decryption with wrong key must return ErrDecryptFailed")
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	alice := generateTestPair(t)
	bob := generateTestPair(t)
	key, _ := DeriveConvKey(alice.PrivateKey, bob.PublicKey)

	ciphertext, err := Encrypt(key, "tamper me")
	require.NoError(t, err)

	// Flip a byte in the middle of the base64 string
	bs := []byte(ciphertext)
	midpoint := len(bs) / 2
	bs[midpoint] ^= 0xFF
	tampered := string(bs)

	_, err = Decrypt(key, tampered)
	assert.Error(t, err, "tampered ciphertext must fail to decrypt")
}
