package infrastructure

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/auth/application"
	"github.com/Briyantama/HRIS-Stery/services/auth/domain"
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
	return s.signer.ValidateAccessToken(token)
}

// ConsumeRefreshToken validates a refresh token and atomically spends it, so each token works once.
// The caller issues the replacement pair (GenerateTokenPair) from freshly loaded state.
func (s *TokenServiceImpl) ConsumeRefreshToken(ctx context.Context, refreshToken string) (*application.TokenClaims, error) {
	claims, err := s.signer.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w: %w", domain.ErrUnauthenticated, err)
	}

	isRevoked, err := s.isRefreshTokenRevoked(ctx, claims.TokenHash)
	if err != nil {
		return nil, fmt.Errorf("check revocation: %w", err)
	}
	if isRevoked {
		return nil, fmt.Errorf("refresh token is revoked: %w", domain.ErrUnauthenticated)
	}

	// DEL is atomic: of concurrent callers, exactly one sees n == 1.
	n, err := s.redis.Del(ctx, refreshTokenKeyPrefix+claims.TokenHash).Result()
	if err != nil {
		return nil, fmt.Errorf("consume refresh token: %w", err)
	}
	if n == 0 {
		return nil, fmt.Errorf("refresh token not found, expired or already used: %w", domain.ErrUnauthenticated)
	}

	return &application.TokenClaims{
		UserID:   claims.UserID,
		TenantID: claims.TenantID,
		Email:    claims.Email,
		Roles:    claims.Roles,
	}, nil
}

// RevokeRefreshToken marks a refresh token as revoked.
func (s *TokenServiceImpl) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	claims, err := s.signer.ValidateRefreshToken(refreshToken)
	if err != nil {
		return fmt.Errorf("invalid token: %w: %w", domain.ErrUnauthenticated, err)
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

// isRefreshTokenRevoked checks if a refresh token has been revoked.
func (s *TokenServiceImpl) isRefreshTokenRevoked(ctx context.Context, tokenHash string) (bool, error) {
	key := revocationListKeyPrefix + tokenHash
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
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
