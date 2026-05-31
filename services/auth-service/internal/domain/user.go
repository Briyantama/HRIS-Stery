package domain

import (
	"crypto/subtle"
	"fmt"
	"net/mail"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User is the aggregate root for user accounts within a tenant.
type User struct {
	id              UserID
	tenantID        TenantID
	email           string
	passwordHash    string
	fullName        string
	isActive        bool
	emailVerified   bool
	roleIDs         []RoleID // cached role IDs for quick access
	lastLoginAt     *time.Time
	lastLoginIPAddr string
	createdAt       time.Time
	updatedAt       time.Time
}

// NewUser creates a new unverified user with hashed password.
func NewUser(id UserID, tenantID TenantID, email, plainPassword, fullName string) (*User, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(plainPassword); err != nil {
		return nil, err
	}

	hash, err := hashPassword(plainPassword)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	return &User{
		id:           id,
		tenantID:     tenantID,
		email:        email,
		passwordHash: hash,
		fullName:     fullName,
		isActive:     true,
		emailVerified: false,
		roleIDs:      []RoleID{},
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// ID returns the user's unique identifier.
func (u *User) ID() UserID {
	return u.id
}

// TenantID returns the tenant this user belongs to.
func (u *User) TenantID() TenantID {
	return u.tenantID
}

// Email returns the user's email address.
func (u *User) Email() string {
	return u.email
}

// PasswordHash returns the stored bcrypt hash (persistence layer only).
func (u *User) PasswordHash() string {
	return u.passwordHash
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	return u.fullName
}

// IsActive returns whether the user account is active.
func (u *User) IsActive() bool {
	return u.isActive
}

// EmailVerified returns whether the user's email has been verified.
func (u *User) EmailVerified() bool {
	return u.emailVerified
}

// VerifyEmail marks the user's email as verified.
func (u *User) VerifyEmail() {
	u.emailVerified = true
	u.updatedAt = time.Now().UTC()
}

// Deactivate disables the user account. The user cannot log in.
func (u *User) Deactivate() {
	u.isActive = false
	u.updatedAt = time.Now().UTC()
}

// Activate re-enables a deactivated account.
func (u *User) Activate() {
	u.isActive = true
	u.updatedAt = time.Now().UTC()
}

// RehydrateUser reconstructs a persisted user without re-hashing the password.
func RehydrateUser(
	id UserID,
	tenantID TenantID,
	email, passwordHash, fullName string,
	isActive, emailVerified bool,
	createdAt, updatedAt time.Time,
	lastLoginAt *time.Time,
	lastLoginIP string,
) (*User, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	user := &User{
		id:              id,
		tenantID:        tenantID,
		email:           email,
		passwordHash:    passwordHash,
		fullName:        fullName,
		isActive:        isActive,
		emailVerified:   emailVerified,
		roleIDs:         []RoleID{},
		lastLoginAt:     lastLoginAt,
		lastLoginIPAddr: lastLoginIP,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
	return user, nil
}
// Returns nil if the password matches, an error otherwise.
func (u *User) VerifyPassword(plainPassword string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.passwordHash), []byte(plainPassword)); err != nil {
		return fmt.Errorf("password mismatch")
	}
	return nil
}

// ChangePassword updates the user's password. The old password must be verified first.
func (u *User) ChangePassword(oldPlain, newPlain string) error {
	if err := u.VerifyPassword(oldPlain); err != nil {
		return fmt.Errorf("old password invalid")
	}
	if err := validatePassword(newPlain); err != nil {
		return err
	}

	hash, err := hashPassword(newPlain)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	u.passwordHash = hash
	u.updatedAt = time.Now().UTC()
	return nil
}

// SetRoles updates the user's role IDs.
func (u *User) SetRoles(roleIDs []RoleID) {
	u.roleIDs = roleIDs
	u.updatedAt = time.Now().UTC()
}

// RoleIDs returns the user's assigned role IDs.
func (u *User) RoleIDs() []RoleID {
	return u.roleIDs
}

// RecordLogin updates the last login timestamp and IP address.
func (u *User) RecordLogin(ipAddr string) {
	now := time.Now().UTC()
	u.lastLoginAt = &now
	u.lastLoginIPAddr = ipAddr
	u.updatedAt = now
}

// LastLoginAt returns the last login timestamp, or nil if never logged in.
func (u *User) LastLoginAt() *time.Time {
	return u.lastLoginAt
}

// LastLoginIPAddr returns the IP address of the last login.
func (u *User) LastLoginIPAddr() string {
	return u.lastLoginIPAddr
}

// CreatedAt returns the account creation timestamp.
func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

// UpdatedAt returns the last update timestamp.
func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// validateEmail checks email format.
func validateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}
	return nil
}

// validatePassword enforces password strength: min 8 chars, mix of upper/lower/digits/symbols.
func validatePassword(pwd string) error {
	if len(pwd) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(pwd) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}
	// Regex: at least one uppercase, one lowercase, one digit, one special char.
	hasUpper, hasLower, hasDigit, hasSymbol := false, false, false, false
	for _, ch := range pwd {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		case ch < 128 && !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')):
			hasSymbol = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSymbol {
		return fmt.Errorf("password must contain uppercase, lowercase, digit, and symbol")
	}
	return nil
}

// hashPassword hashes a plain-text password using bcrypt.
func hashPassword(plainPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ConstantTimeEqual compares two strings in constant time to prevent timing attacks.
// Used when comparing hashes or tokens.
func ConstantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
