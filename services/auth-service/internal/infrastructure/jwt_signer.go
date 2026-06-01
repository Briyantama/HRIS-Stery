package infrastructure

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/application"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
)

// JWTSigner handles JWT generation and validation with RSA keys.
type JWTSigner struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	accessTTL  time.Duration // e.g., 15 minutes
	refreshTTL time.Duration // e.g., 7 days
}

// CustomClaims extends jwt.RegisteredClaims with app-specific fields.
type CustomClaims struct {
	UserID   domain.UserID   `json:"user_id"`
	TenantID domain.TenantID `json:"tenant_id"`
	TID      string          `json:"tid"`
	Email    string          `json:"email"`
	Roles    []string        `json:"roles"`
	jwt.RegisteredClaims
}

// RefreshClaims for refresh tokens (minimal data to avoid token bloat).
type RefreshClaims struct {
	UserID    domain.UserID   `json:"user_id"`
	TenantID  domain.TenantID `json:"tenant_id"`
	Email     string          `json:"email"`
	Roles     []string        `json:"roles"`
	TokenHash string          `json:"token_hash"`
	jwt.RegisteredClaims
}

// NewJWTSigner creates a new JWT signer with RSA keys.
func NewJWTSigner(
	privateKeyPEM []byte,
	publicKeyPEM []byte,
	issuer string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) (*JWTSigner, error) {
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return &JWTSigner{
		privateKey: privKey,
		publicKey:  pubKey,
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}, nil
}

// GenerateAccessToken creates a short-lived access token.
func (s *JWTSigner) GenerateAccessToken(
	userID domain.UserID,
	tenantID domain.TenantID,
	email string,
	roles []string,
) (string, int32, error) {
	now := time.Now()
	expiresAt := now.Add(s.accessTTL)

	claims := CustomClaims{
		UserID:   userID,
		TenantID: tenantID,
		TID:      tenantID.String(),
		Email:    email,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", 0, fmt.Errorf("sign access token: %w", err)
	}

	ttlSeconds := int32(s.accessTTL.Seconds())
	return signedToken, ttlSeconds, nil
}

// GenerateRefreshToken creates a long-lived refresh token.
// tokenHash should be the SHA256 hash of a random value (for revocation tracking).
func (s *JWTSigner) GenerateRefreshToken(
	userID domain.UserID,
	tenantID domain.TenantID,
	email string,
	roles []string,
	tokenHash string,
) (string, error) {
	now := time.Now()
	expiresAt := now.Add(s.refreshTTL)

	claims := RefreshClaims{
		UserID:    userID,
		TenantID:  tenantID,
		Email:     email,
		Roles:     roles,
		TokenHash: tokenHash,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign refresh token: %w", err)
	}

	return signedToken, nil
}

// ValidateAccessToken verifies and parses an access token.
func (s *JWTSigner) ValidateAccessToken(tokenString string) (*application.TokenClaims, error) {
	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return &application.TokenClaims{
		UserID:    claims.UserID,
		TenantID:  claims.TenantID,
		Email:     claims.Email,
		Roles:     claims.Roles,
		ExpiresAt: claims.ExpiresAt.Unix(),
	}, nil
}

// ValidateRefreshToken verifies and parses a refresh token.
func (s *JWTSigner) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse refresh token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("refresh token is invalid")
	}

	return claims, nil
}

// HashToken returns the SHA256 hash of a token string.
// Used for revocation tracking (store hash, not raw token).
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
