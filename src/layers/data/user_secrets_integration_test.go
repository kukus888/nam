//go:build integration

package data

import (
	"fmt"
	"testing"
	"time"
)

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func TestUserCRUDIntegration(t *testing.T) {
	pool := setupIntegrationDB(t)

	username := uniqueName("user")
	userDTO := UserDTO{
		Username: username,
		Email:    uniqueName("mail") + "@example.com",
		Password: "Start123!",
	}

	userID, err := CreateUser(pool, userDTO, 1)
	if err != nil {
		t.Fatalf("CreateUser() returned error: %v", err)
	}
	if userID == nil || *userID == 0 {
		t.Fatal("expected CreateUser() to return a valid user id")
	}

	byID, err := GetUserById(pool, *userID)
	if err != nil {
		t.Fatalf("GetUserById() returned error: %v", err)
	}
	if byID == nil {
		t.Fatal("expected user to exist after CreateUser()")
	}
	if byID.Username != username {
		t.Fatalf("username mismatch, got %q want %q", byID.Username, username)
	}

	byUsername, err := GetUserByUsername(pool, username)
	if err != nil {
		t.Fatalf("GetUserByUsername() returned error: %v", err)
	}
	if byUsername == nil || byUsername.Id != *userID {
		t.Fatal("expected lookup by username to return created user")
	}

	updatedEmail := uniqueName("mail-upd") + "@example.com"
	byID.Email = updatedEmail
	byID.Color = "blue"
	byID.RoleId = 2
	if err := byID.UpdateWithoutPassword(pool); err != nil {
		t.Fatalf("UpdateWithoutPassword() returned error: %v", err)
	}

	updatedUser, err := GetUserById(pool, *userID)
	if err != nil {
		t.Fatalf("GetUserById() after update returned error: %v", err)
	}
	if updatedUser == nil {
		t.Fatal("expected updated user to exist")
	}
	if updatedUser.Email != updatedEmail || updatedUser.RoleId != 2 || updatedUser.Color != "blue" {
		t.Fatal("updated user fields do not match expected values")
	}

	if err := updatedUser.UpdatePassword(pool, "NewPassword456!"); err != nil {
		t.Fatalf("UpdatePassword() returned error: %v", err)
	}

	withNewPassword, err := GetUserById(pool, *userID)
	if err != nil {
		t.Fatalf("GetUserById() after password update returned error: %v", err)
	}
	if withNewPassword == nil {
		t.Fatal("expected user to exist after password update")
	}
	if err := VerifyPassword(withNewPassword.PasswordHash, "NewPassword456!"); err != nil {
		t.Fatalf("expected new password hash to verify, got error: %v", err)
	}

	if err := withNewPassword.Delete(pool); err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}

	afterDelete, err := GetUserById(pool, *userID)
	if err != nil {
		t.Fatalf("GetUserById() after delete returned error: %v", err)
	}
	if afterDelete != nil {
		t.Fatal("expected deleted user to be absent")
	}
}

func TestSecretsCRUDIntegration(t *testing.T) {
	pool := setupIntegrationDB(t)

	userID, err := CreateUser(pool, UserDTO{
		Username: uniqueName("author"),
		Email:    uniqueName("author-mail") + "@example.com",
		Password: "Author123!",
	}, 1)
	if err != nil {
		t.Fatalf("CreateUser() for secret owner returned error: %v", err)
	}

	secret := &SecretDAO{
		Type:        "password",
		Name:        uniqueName("secret"),
		Description: "integration secret",
		Data:        []byte("encrypted-by-caller"),
		Metadata: map[string]interface{}{
			"scope": "integration",
		},
		CreatedBy: userID,
		UpdatedBy: userID,
	}

	secretID, err := secret.DbInsert(pool)
	if err != nil {
		t.Fatalf("DbInsert() returned error: %v", err)
	}
	if secretID == nil || *secretID == 0 {
		t.Fatal("expected DbInsert() to return a valid secret id")
	}

	byID, err := GetSecretById(pool, *secretID)
	if err != nil {
		t.Fatalf("GetSecretById() returned error: %v", err)
	}
	if byID.Name != secret.Name || byID.Type != secret.Type {
		t.Fatal("secret read by id does not match inserted values")
	}

	byName, err := GetSecretByName(pool, secret.Name)
	if err != nil {
		t.Fatalf("GetSecretByName() returned error: %v", err)
	}
	if byName.Id != *secretID {
		t.Fatal("secret lookup by name did not return expected id")
	}

	updated := *byID
	updated.Type = "ssh_key"
	updated.Description = "updated secret"
	updated.Data = []byte("updated-encrypted-data")
	updated.Metadata = map[string]interface{}{"scope": "updated"}
	updated.UpdatedBy = userID

	if err := UpdateSecret(pool, *secretID, &updated); err != nil {
		t.Fatalf("UpdateSecret() returned error: %v", err)
	}

	afterUpdate, err := GetSecretById(pool, *secretID)
	if err != nil {
		t.Fatalf("GetSecretById() after update returned error: %v", err)
	}
	if afterUpdate.Type != "ssh_key" || afterUpdate.Description != "updated secret" {
		t.Fatal("updated secret fields do not match expected values")
	}

	allSecrets, err := GetAllSecrets(pool)
	if err != nil {
		t.Fatalf("GetAllSecrets() returned error: %v", err)
	}
	if len(allSecrets) == 0 {
		t.Fatal("expected at least one secret in GetAllSecrets()")
	}

	sshSecrets, err := GetSshSecrets(pool)
	if err != nil {
		t.Fatalf("GetSshSecrets() returned error: %v", err)
	}
	if len(sshSecrets) == 0 {
		t.Fatal("expected updated ssh_key secret to be returned by GetSshSecrets()")
	}

	if err := DeleteSecret(pool, *secretID); err != nil {
		t.Fatalf("DeleteSecret() returned error: %v", err)
	}

	if _, err := GetSecretById(pool, *secretID); err == nil {
		t.Fatal("expected GetSecretById() to fail after deletion")
	}
}
