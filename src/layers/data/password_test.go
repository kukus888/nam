package data

import "testing"

func TestHashPasswordAndVerifyPassword(t *testing.T) {
	t.Parallel()

	password := "CorrectHorseBatteryStaple!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() returned error: %v", err)
	}
	if hash == password {
		t.Fatal("expected password hash to differ from plaintext password")
	}

	if err := VerifyPassword(hash, password); err != nil {
		t.Fatalf("VerifyPassword() returned error for valid password: %v", err)
	}
}

func TestHashPasswordUsesSalt(t *testing.T) {
	t.Parallel()

	password := "same-password"
	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() first call returned error: %v", err)
	}
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() second call returned error: %v", err)
	}

	if hash1 == hash2 {
		t.Fatal("expected different hashes for the same password")
	}
}

func TestVerifyPasswordRejectsInvalidPasswordAndHash(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("good-password")
	if err != nil {
		t.Fatalf("HashPassword() returned error: %v", err)
	}

	if err := VerifyPassword(hash, "bad-password"); err == nil {
		t.Fatal("expected VerifyPassword() to fail for wrong password")
	}

	if err := VerifyPassword("not-a-bcrypt-hash", "whatever"); err == nil {
		t.Fatal("expected VerifyPassword() to fail for malformed bcrypt hash")
	}
}

func TestUserHasRoleHierarchy(t *testing.T) {
	t.Parallel()

	admin := User{RoleId: 1}
	operator := User{RoleId: 2}
	viewer := User{RoleId: 3}

	if !admin.HasRole("Admin") || !admin.HasRole("Operator") || !admin.HasRole("Viewer") {
		t.Fatal("admin role hierarchy check failed")
	}
	if operator.HasRole("Admin") || !operator.HasRole("Operator") || !operator.HasRole("Viewer") {
		t.Fatal("operator role hierarchy check failed")
	}
	if viewer.HasRole("Admin") || viewer.HasRole("Operator") || !viewer.HasRole("Viewer") {
		t.Fatal("viewer role hierarchy check failed")
	}

	unknownRoleUser := User{RoleId: 999}
	if unknownRoleUser.HasRole("Viewer") {
		t.Fatal("expected user with unknown role id to not match known roles")
	}
	if admin.HasRole("NonExistentRole") {
		t.Fatal("expected unknown required role to return false")
	}
}
