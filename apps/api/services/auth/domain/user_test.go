package domain

import (
	"testing"
)

func TestNewUser(t *testing.T) {
	userID := GenerateUserID()
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	email := "user@example.com"
	password := "SecurePass123!"
	name := "John Doe"

	user, err := NewUser(userID, tenantID, email, password, name)
	if err != nil {
		t.Fatalf("NewUser failed: %v", err)
	}

	if !user.ID().Equals(userID) {
		t.Errorf("ID mismatch")
	}
	if !user.TenantID().Equals(tenantID) {
		t.Errorf("Tenant ID mismatch")
	}
	if user.Email() != email {
		t.Errorf("Email mismatch")
	}
	if user.FullName() != name {
		t.Errorf("Name mismatch")
	}
	if !user.IsActive() {
		t.Errorf("User should be active")
	}
	if user.EmailVerified() {
		t.Errorf("Email should not be verified")
	}
}

func TestUserVerifyPassword(t *testing.T) {
	userID := GenerateUserID()
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	password := "SecurePass123!"

	user, _ := NewUser(userID, tenantID, "user@example.com", password, "John")

	// Correct password should verify
	if err := user.VerifyPassword(password); err != nil {
		t.Errorf("Password verification failed: %v", err)
	}

	// Wrong password should fail
	if err := user.VerifyPassword("WrongPass123!"); err == nil {
		t.Errorf("Wrong password should not verify")
	}
}

func TestUserChangePassword(t *testing.T) {
	userID := GenerateUserID()
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	oldPassword := "SecurePass123!"
	newPassword := "NewSecurePass456!"

	user, _ := NewUser(userID, tenantID, "user@example.com", oldPassword, "John")

	// Change password
	if err := user.ChangePassword(oldPassword, newPassword); err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	// Old password should not work
	if err := user.VerifyPassword(oldPassword); err == nil {
		t.Errorf("Old password should not work after change")
	}

	// New password should work
	if err := user.VerifyPassword(newPassword); err != nil {
		t.Errorf("New password should work: %v", err)
	}
}

func TestUserDeactivate(t *testing.T) {
	userID := GenerateUserID()
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	user, _ := NewUser(userID, tenantID, "user@example.com", "SecurePass123!", "John")

	if !user.IsActive() {
		t.Fatalf("User should be active initially")
	}

	user.Deactivate()
	if user.IsActive() {
		t.Errorf("User should be inactive after deactivate")
	}

	user.Activate()
	if !user.IsActive() {
		t.Errorf("User should be active after activate")
	}
}

func TestUserVerifyEmail(t *testing.T) {
	userID := GenerateUserID()
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	user, _ := NewUser(userID, tenantID, "user@example.com", "SecurePass123!", "John")

	if user.EmailVerified() {
		t.Fatalf("Email should not be verified initially")
	}

	user.VerifyEmail()
	if !user.EmailVerified() {
		t.Errorf("Email should be verified after VerifyEmail")
	}
}

func TestUserRecordLogin(t *testing.T) {
	userID := GenerateUserID()
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	user, _ := NewUser(userID, tenantID, "user@example.com", "SecurePass123!", "John")

	if user.LastLoginAt() != nil {
		t.Fatalf("LastLoginAt should be nil initially")
	}

	ipAddr := "192.168.1.1"
	user.RecordLogin(ipAddr)

	if user.LastLoginAt() == nil {
		t.Errorf("LastLoginAt should be set after RecordLogin")
	}
	if user.LastLoginIPAddr() != ipAddr {
		t.Errorf("LastLoginIPAddr mismatch")
	}
}

func TestUserSetRoles(t *testing.T) {
	userID := GenerateUserID()
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	user, _ := NewUser(userID, tenantID, "user@example.com", "SecurePass123!", "John")

	roleID1 := MustNewRoleID("role-1")
	roleID2 := MustNewRoleID("role-2")
	roles := []RoleID{roleID1, roleID2}

	user.SetRoles(roles)

	if len(user.RoleIDs()) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(user.RoleIDs()))
	}
}

func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		password  string
		shouldErr bool
		reason    string
	}{
		{"Short1!", true, "too short"},
		{"NoDigits!", true, "no digits"},
		{"nouppercasesymbols1!", true, "no uppercase"},
		{"NOLOWERCASESYMBOLS1!", true, "no lowercase"},
		{"NoSymbols1", true, "no symbols"},
		{"ValidPass123!", false, "valid password"},
	}

	userID := GenerateUserID()
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")

	for _, tt := range tests {
		_, err := NewUser(userID, tenantID, "user@example.com", tt.password, "John")
		if (err != nil) != tt.shouldErr {
			t.Errorf("Password %s: expected err=%v, got err=%v. Reason: %s",
				tt.password, tt.shouldErr, err != nil, tt.reason)
		}
	}
}

func TestEmailValidation(t *testing.T) {
	tests := []struct {
		email     string
		shouldErr bool
	}{
		{"valid@example.com", false},
		{"invalid-email", true},
		{"@nodomain.com", true},
		{"noat.com", true},
	}

	userID := GenerateUserID()
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")

	for _, tt := range tests {
		_, err := NewUser(userID, tenantID, tt.email, "ValidPass123!", "John")
		if (err != nil) != tt.shouldErr {
			t.Errorf("Email %s: expected err=%v, got err=%v",
				tt.email, tt.shouldErr, err != nil)
		}
	}
}
