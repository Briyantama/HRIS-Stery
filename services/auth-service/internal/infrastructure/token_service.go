package infrastructure

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/hris-stery/hris-stery/services/auth-service/internal/application"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	refreshTokenKeyPrefix   = "refresh_token:"
	revocationListKeyPrefix = "revoked_token:"
)

// TokenServiceImpl implements application.TokenService with JWT + Redis.
type TokenServiceImpl struct {
	signer *JWTSigner
	redis  *redis.Client
	ttl    time.Duration
}

// NewTokenService creates a new token service.
func NewTokenService(signer *JWTSigner, redis *redis.Client, refreshTokenTTL time.Duration) *TokenServiceImpl {
	return &TokenServiceImpl{
		signer: signer,
		redis:  redis,
		ttl:    refreshTokenTTL,
	}
}

// GenerateTokenPair creates both access and refresh tokens.
func (s *TokenServiceImpl) GenerateTokenPair(
	ctx context.Context,
	userID domain.UserID,
	tenantID domain.TenantID,
	email string,
	roles []string,
) (string, string, int32, error) {
	accessToken, ttl, err := s.signer.GenerateAccessToken(userID, tenantID, email, roles)
	if err != nil {
		return "", "", 0, fmt.Errorf("generate access token: %w", err)
	}

	tokenHash, err := s.storeRefreshToken(ctx)
	if err != nil {
		return "", "", 0, err
	}

	refreshToken, err := s.signer.GenerateRefreshToken(userID, tenantID, email, roles, tokenHash)
	if err != nil {
		return "", "", 0, fmt.Errorf("generate refresh token: %w", err)
	}

	return accessToken, refreshToken, ttl, nil
}

// ValidateAccessToken verifies an access token.
func (s *TokenServiceImpl) ValidateAccessToken(ctx context.Context, token string) (*application.TokenClaims, error) {
	claims, err := s.signer.ValidateAccessToken(token)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// RefreshAccessToken rotates the refresh token and issues a new access token.
func (s *TokenServiceImpl) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, int32, error) {
	claims, err := s.signer.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid refresh token: %w", err)
	}

	isRevoked, err := s.IsRefreshTokenRevoked(ctx, claims.TokenHash)
	if err != nil {
		return "", "", 0, fmt.Errorf("check revocation: %w", err)
	}
	if isRevoked {
		return "", "", 0, fmt.Errorf("refresh token is revoked")
	}

	activeKey := refreshTokenKeyPrefix + claims.TokenHash
	exists, err := s.redis.Exists(ctx, activeKey).Result()
	if err != nil {
		return "", "", 0, fmt.Errorf("check active refresh token: %w", err)
	}
	if exists == 0 {
		return "", "", 0, fmt.Errorf("refresh token not found or expired")
	}

	if err := s.redis.Del(ctx, activeKey).Err(); err != nil {
		return "", "", 0, fmt.Errorf("remove active refresh token: %w", err)
	}
	if err := s.redis.Set(ctx, revocationListKeyPrefix+claims.TokenHash, "revoked", s.ttl).Err(); err != nil {
		return "", "", 0, fmt.Errorf("revoke old refresh token: %w", err)
	}

	newTokenHash, err := s.storeRefreshToken(ctx)
	if err != nil {
		return "", "", 0, err
	}

	newRefreshToken, err := s.signer.GenerateRefreshToken(
		claims.UserID,
		claims.TenantID,
		claims.Email,
		claims.Roles,
		newTokenHash,
	)
	if err != nil {
		return "", "", 0, fmt.Errorf("generate rotated refresh token: %w", err)
	}

	accessToken, ttl, err := s.signer.GenerateAccessToken(
		claims.UserID,
		claims.TenantID,
		claims.Email,
		claims.Roles,
	)
	if err != nil {
		return "", "", 0, fmt.Errorf("generate new access token: %w", err)
	}

	return accessToken, newRefreshToken, ttl, nil
}

// RevokeRefreshToken marks a refresh token as revoked.
func (s *TokenServiceImpl) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	claims, err := s.signer.ValidateRefreshToken(refreshToken)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	key := revocationListKeyPrefix + claims.TokenHash
	if err := s.redis.Set(ctx, key, "revoked", s.ttl).Err(); err != nil {
		return fmt.Errorf("store revocation: %w", err)
	}

	if err := s.redis.Del(ctx, refreshTokenKeyPrefix+claims.TokenHash).Err(); err != nil {
		return fmt.Errorf("delete active token: %w", err)
	}

	return nil
}

// IsRefreshTokenRevoked checks if a refresh token has been revoked.
func (s *TokenServiceImpl) IsRefreshTokenRevoked(ctx context.Context, tokenHash string) (bool, error) {
	key := revocationListKeyPrefix + tokenHash
	exists := s.redis.Exists(ctx, key).Val()
	return exists > 0, nil
}

func (s *TokenServiceImpl) storeRefreshToken(ctx context.Context) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate random token: %w", err)
	}
	tokenValue := hex.EncodeToString(tokenBytes)
	tokenHash := HashToken(tokenValue)

	key := refreshTokenKeyPrefix + tokenHash
	if err := s.redis.Set(ctx, key, tokenValue, s.ttl).Err(); err != nil {
		return "", fmt.Errorf("store refresh token: %w", err)
	}
	return tokenHash, nil
}
