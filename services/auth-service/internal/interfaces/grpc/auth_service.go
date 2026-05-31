package grpc

import (
	"context"
	"time"

	authv1 "github.com/hris-stery/hris-stery/gen/go/hris/auth/v1"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/application/queries"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AuthServiceServer implements authv1.AuthServiceServer.
type AuthServiceServer struct {
	authv1.UnimplementedAuthServiceServer
	loginHandler            *commands.LoginHandler
	refreshTokenHandler     *commands.RefreshTokenHandler
	revokeTokenHandler      *commands.RevokeTokenHandler
	registerTenantHandler   *commands.RegisterTenantHandler
	validateTokenHandler    *queries.ValidateTokenHandler
	getPermissionsHandler   *queries.GetPermissionsHandler
}

// NewAuthServiceServer creates a new auth service gRPC server.
func NewAuthServiceServer(
	loginHandler *commands.LoginHandler,
	refreshTokenHandler *commands.RefreshTokenHandler,
	revokeTokenHandler *commands.RevokeTokenHandler,
	registerTenantHandler *commands.RegisterTenantHandler,
	validateTokenHandler *queries.ValidateTokenHandler,
	getPermissionsHandler *queries.GetPermissionsHandler,
) *AuthServiceServer {
	return &AuthServiceServer{
		loginHandler:          loginHandler,
		refreshTokenHandler:   refreshTokenHandler,
		revokeTokenHandler:    revokeTokenHandler,
		registerTenantHandler: registerTenantHandler,
		validateTokenHandler:  validateTokenHandler,
		getPermissionsHandler: getPermissionsHandler,
	}
}

// Login authenticates a user and returns JWT tokens.
func (s *AuthServiceServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if req.Email == "" || req.Password == "" || req.TenantSlug == "" {
		return nil, status.Error(codes.InvalidArgument, "email, password, and tenant_slug are required")
	}

	result, err := s.loginHandler.Handle(ctx, commands.LoginCommand{
		Email:      req.Email,
		Password:   req.Password,
		TenantSlug: req.TenantSlug,
		IPAddr:     extractIPAddr(ctx), // Extract from context if available
	})
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "login failed: %v", err)
	}

	return &authv1.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.AccessTokenTTL,
		Claims: &authv1.TokenClaims{
			UserId:    result.UserID.String(),
			TenantId:  result.TenantID.String(),
			Email:     result.Email,
			Roles:     result.Roles,
			ExpiresAt: nil, // Set in token service if needed
		},
	}, nil
}

// RefreshToken exchanges a refresh token for a new access token.
func (s *AuthServiceServer) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	result, err := s.refreshTokenHandler.Handle(ctx, commands.RefreshTokenCommand{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "refresh failed: %v", err)
	}

	return &authv1.RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.AccessTokenTTL,
	}, nil
}

// ValidateToken verifies an access token and returns its claims.
func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token is required")
	}

	claims, err := s.validateTokenHandler.Handle(ctx, queries.ValidateTokenQuery{
		AccessToken: req.AccessToken,
	})
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "invalid token: %v", err)
	}

	return &authv1.ValidateTokenResponse{
		Claims: &authv1.TokenClaims{
			UserId:    claims.UserID.String(),
			TenantId:  claims.TenantID.String(),
			Email:     claims.Email,
			Roles:     claims.Roles,
			ExpiresAt: timestamppb.New(time.Unix(claims.ExpiresAt, 0)),
		},
	}, nil
}

// RevokeToken invalidates a refresh token (logout).
func (s *AuthServiceServer) RevokeToken(ctx context.Context, req *authv1.RevokeTokenRequest) (*authv1.RevokeTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	if err := s.revokeTokenHandler.Handle(ctx, commands.RevokeTokenCommand{
		RefreshToken: req.RefreshToken,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "revoke failed: %v", err)
	}

	return &authv1.RevokeTokenResponse{}, nil
}

// GetPermissions returns all permissions for a user.
func (s *AuthServiceServer) GetPermissions(ctx context.Context, req *authv1.GetPermissionsRequest) (*authv1.GetPermissionsResponse, error) {
	if req.UserId == "" || req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and tenant_id are required")
	}

	userID, err := domain.NewUserID(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	tenantID, err := domain.NewTenantID(req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	perms, err := s.getPermissionsHandler.Handle(ctx, queries.GetPermissionsQuery{
		UserID:   userID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get permissions failed: %v", err)
	}

	permStrings := make([]string, len(perms))
	for i, p := range perms {
		permStrings[i] = p.String()
	}

	return &authv1.GetPermissionsResponse{
		Permissions: permStrings,
	}, nil
}

// RegisterTenant creates a new tenant with admin user.
func (s *AuthServiceServer) RegisterTenant(ctx context.Context, req *authv1.RegisterTenantRequest) (*authv1.RegisterTenantResponse, error) {
	if req.CompanyName == "" || req.TenantSlug == "" || req.AdminEmail == "" || req.AdminPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "all fields are required")
	}

	result, err := s.registerTenantHandler.Handle(ctx, commands.RegisterTenantCommand{
		CompanyName:   req.CompanyName,
		TenantSlug:    req.TenantSlug,
		AdminEmail:    req.AdminEmail,
		AdminPassword: req.AdminPassword,
		AdminName:     req.AdminName,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "registration failed: %v", err)
	}

	return &authv1.RegisterTenantResponse{
		TenantId:     result.TenantID.String(),
		UserId:       result.UserID.String(),
		AccessToken:  "", // RegisterTenant does not return tokens; user must login after
		RefreshToken: "",
	}, nil
}

// extractIPAddr retrieves the client IP from the gRPC context.
// In a real scenario, this would extract from gRPC metadata set by middleware.
func extractIPAddr(ctx context.Context) string {
	// TODO: Extract from gRPC metadata (set by middleware interceptor)
	return "0.0.0.0"
}
