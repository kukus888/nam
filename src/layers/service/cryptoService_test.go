package services

import (
	"bytes"
	"testing"

	"kukus/nam/v2/layers/data"
)

func TestCryptoServiceEncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()

	svc := NewCryptoService("test-master-password", []byte("test-salt"))
	plaintext := []byte("super-secret-value")

	ciphertext, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() returned error: %v", err)
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("expected ciphertext to differ from plaintext")
	}

	decrypted, err := svc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() returned error: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("decrypted plaintext mismatch, got %q want %q", decrypted, plaintext)
	}
}

func TestCryptoServiceEncryptUsesRandomNonce(t *testing.T) {
	t.Parallel()

	svc := NewCryptoService("test-master-password", []byte("test-salt"))
	plaintext := []byte("same plaintext")

	ciphertext1, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() first call returned error: %v", err)
	}
	ciphertext2, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() second call returned error: %v", err)
	}

	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Fatal("expected different ciphertexts for same plaintext due to random nonce")
	}
}

func TestCryptoServiceDecryptRejectsShortCiphertext(t *testing.T) {
	t.Parallel()

	svc := NewCryptoService("test-master-password", []byte("test-salt"))

	if _, err := svc.Decrypt([]byte("short")); err == nil {
		t.Fatal("expected error for ciphertext shorter than nonce")
	}
}

func TestCryptoServiceDecryptWithWrongKeyFails(t *testing.T) {
	t.Parallel()

	correct := NewCryptoService("master-a", []byte("salt"))
	wrong := NewCryptoService("master-b", []byte("salt"))

	ciphertext, err := correct.Encrypt([]byte("secret"))
	if err != nil {
		t.Fatalf("Encrypt() returned error: %v", err)
	}

	if _, err := wrong.Decrypt(ciphertext); err == nil {
		t.Fatal("expected decrypt failure with wrong key")
	}
}

func TestCryptoServiceEncryptSecretAndDecryptDAO(t *testing.T) {
	t.Parallel()

	svc := NewCryptoService("test-master-password", []byte("test-salt"))
	createdBy := uint64(11)
	updatedBy := uint64(22)
	secret := &data.Secret{
		Id:          123,
		Type:        "password",
		Name:        "db-password",
		Description: "database password",
		Data:        []byte("plain-data"),
		Metadata: map[string]interface{}{
			"env": "test",
		},
		CreatedBy: &createdBy,
		UpdatedBy: &updatedBy,
	}

	dao, err := svc.EncryptSecret(secret)
	if err != nil {
		t.Fatalf("EncryptSecret() returned error: %v", err)
	}
	if len(dao.Data) == 0 {
		t.Fatal("expected encrypted data to be set")
	}

	decrypted, err := svc.DecryptDAO(dao)
	if err != nil {
		t.Fatalf("DecryptDAO() returned error: %v", err)
	}

	if !bytes.Equal(decrypted.Data, secret.Data) {
		t.Fatalf("decrypted data mismatch, got %q want %q", decrypted.Data, secret.Data)
	}
	if decrypted.Name != secret.Name || decrypted.Type != secret.Type {
		t.Fatal("decrypted secret metadata fields do not match source")
	}
}
