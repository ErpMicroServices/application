package directives

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/erpmicroservices/auth-go/pkg/middleware"
	"github.com/erpmicroservices/auth-go/pkg/oauth2"
)

// AuthDirective implements the @auth GraphQL directive
type AuthDirective struct {
	oauth2Client *oauth2.Client
}

// NewAuthDirective creates a new auth directive handler
func NewAuthDirective(oauth2Client *oauth2.Client) *AuthDirective {
	return &AuthDirective{
		oauth2Client: oauth2Client,
	}
}

// Auth is the GraphQL directive handler for @auth
func Auth(oauth2Client *oauth2.Client) func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	directive := NewAuthDirective(oauth2Client)
	return directive.Handle
}

// Handle processes the @auth directive
func (d *AuthDirective) Handle(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	// Get authentication context
	authCtx, authenticated := middleware.GetAuthContext(ctx)
	if !authenticated || authCtx == nil {
		log.Debug().
			Str("operation", graphql.GetOperationContext(ctx).OperationName).
			Str("field", graphql.GetFieldContext(ctx).Field.Name).
			Msg("GraphQL @auth directive: Authentication required")
		
		return nil, &gqlerror.Error{
			Message: "Authentication required",
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": "UNAUTHENTICATED",
			},
		}
	}

	if !authCtx.Authenticated {
		log.Debug().
			Str("operation", graphql.GetOperationContext(ctx).OperationName).
			Str("field", graphql.GetFieldContext(ctx).Field.Name).
			Msg("GraphQL @auth directive: User not authenticated")
		
		return nil, &gqlerror.Error{
			Message: "Authentication required",
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": "UNAUTHENTICATED",
			},
		}
	}

	log.Debug().
		Str("operation", graphql.GetOperationContext(ctx).OperationName).
		Str("field", graphql.GetFieldContext(ctx).Field.Name).
		Str("subject", authCtx.Subject).
		Msg("GraphQL @auth directive: Access granted")

	return next(ctx)
}

// AuthWithRequiredArgs handles @auth directive with additional arguments
type AuthWithRequiredArgs struct {
	RequireEmailVerified *bool `json:"requireEmailVerified"`
	RequireServiceAccount *bool `json:"requireServiceAccount"`
}

// AuthWithArgs is the GraphQL directive handler for @auth with arguments
func AuthWithArgs(oauth2Client *oauth2.Client) func(ctx context.Context, obj interface{}, next graphql.Resolver, args *AuthWithRequiredArgs) (interface{}, error) {
	directive := NewAuthDirective(oauth2Client)
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, args *AuthWithRequiredArgs) (interface{}, error) {
		return directive.HandleWithArgs(ctx, obj, next, args)
	}
}

// HandleWithArgs processes the @auth directive with additional validation arguments
func (d *AuthDirective) HandleWithArgs(ctx context.Context, obj interface{}, next graphql.Resolver, args *AuthWithRequiredArgs) (interface{}, error) {
	// First perform basic authentication check
	authCtx, authenticated := middleware.GetAuthContext(ctx)
	if !authenticated || authCtx == nil || !authCtx.Authenticated {
		log.Debug().
			Str("operation", graphql.GetOperationContext(ctx).OperationName).
			Str("field", graphql.GetFieldContext(ctx).Field.Name).
			Msg("GraphQL @auth directive: Authentication required")
		
		return nil, &gqlerror.Error{
			Message: "Authentication required",
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": "UNAUTHENTICATED",
			},
		}
	}

	// Additional validations based on arguments
	if args != nil {
		// Check if email verification is required
		if args.RequireEmailVerified != nil && *args.RequireEmailVerified {
			// Try to get user info if we don't have it in the auth context
			if authCtx.ValidationResult != nil && authCtx.ValidationResult.Claims != nil {
				// Check JWT claims for email verification
				if !authCtx.ValidationResult.Claims.EmailVerified {
					return nil, &gqlerror.Error{
						Message: "Email verification required",
						Path:    graphql.GetPath(ctx),
						Extensions: map[string]interface{}{
							"code": "EMAIL_NOT_VERIFIED",
						},
					}
				}
			} else {
				// Fetch user info from OAuth2 provider
				userInfo, err := d.oauth2Client.GetUserInfo(ctx, authCtx.Token)
				if err != nil {
					log.Error().Err(err).Msg("Failed to get user info for email verification check")
					return nil, &gqlerror.Error{
						Message: "Unable to verify email status",
						Path:    graphql.GetPath(ctx),
						Extensions: map[string]interface{}{
							"code": "VERIFICATION_ERROR",
						},
					}
				}
				
				if !userInfo.EmailVerified {
					return nil, &gqlerror.Error{
						Message: "Email verification required",
						Path:    graphql.GetPath(ctx),
						Extensions: map[string]interface{}{
							"code": "EMAIL_NOT_VERIFIED",
						},
					}
				}
			}
		}

		// Check if service account is required
		if args.RequireServiceAccount != nil && *args.RequireServiceAccount {
			if !authCtx.IsServiceAccount() {
				return nil, &gqlerror.Error{
					Message: "Service account required",
					Path:    graphql.GetPath(ctx),
					Extensions: map[string]interface{}{
						"code": "SERVICE_ACCOUNT_REQUIRED",
					},
				}
			}
		}
	}

	log.Debug().
		Str("operation", graphql.GetOperationContext(ctx).OperationName).
		Str("field", graphql.GetFieldContext(ctx).Field.Name).
		Str("subject", authCtx.Subject).
		Msg("GraphQL @auth directive with args: Access granted")

	return next(ctx)
}

// GetAuthenticatedUser is a helper function to get authenticated user info from GraphQL context
func GetAuthenticatedUser(ctx context.Context) (*middleware.AuthContext, error) {
	authCtx, authenticated := middleware.GetAuthContext(ctx)
	if !authenticated || authCtx == nil || !authCtx.Authenticated {
		return nil, fmt.Errorf("user not authenticated")
	}
	return authCtx, nil
}

// MustGetAuthenticatedUser is a helper function that panics if user is not authenticated
func MustGetAuthenticatedUser(ctx context.Context) *middleware.AuthContext {
	authCtx, err := GetAuthenticatedUser(ctx)
	if err != nil {
		panic(err)
	}
	return authCtx
}

// GetCurrentUserID is a helper function to get the current user's ID from GraphQL context
func GetCurrentUserID(ctx context.Context) (string, error) {
	authCtx, err := GetAuthenticatedUser(ctx)
	if err != nil {
		return "", err
	}
	return authCtx.Subject, nil
}

// IsCurrentUser checks if the provided user ID matches the authenticated user
func IsCurrentUser(ctx context.Context, userID string) bool {
	currentUserID, err := GetCurrentUserID(ctx)
	if err != nil {
		return false
	}
	return currentUserID == userID
}

// GetCurrentUserRoles returns the roles of the authenticated user
func GetCurrentUserRoles(ctx context.Context) ([]string, error) {
	authCtx, err := GetAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	return authCtx.Roles, nil
}

// GetCurrentUserAuthorities returns the authorities of the authenticated user
func GetCurrentUserAuthorities(ctx context.Context) ([]string, error) {
	authCtx, err := GetAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	return authCtx.Authorities, nil
}

// HasCurrentUserRole checks if the authenticated user has a specific role
func HasCurrentUserRole(ctx context.Context, role string) bool {
	authCtx, err := GetAuthenticatedUser(ctx)
	if err != nil {
		return false
	}
	return authCtx.HasRole(role)
}

// HasCurrentUserAnyRole checks if the authenticated user has any of the specified roles
func HasCurrentUserAnyRole(ctx context.Context, roles ...string) bool {
	authCtx, err := GetAuthenticatedUser(ctx)
	if err != nil {
		return false
	}
	return authCtx.HasAnyRole(roles...)
}

// HasCurrentUserAuthority checks if the authenticated user has a specific authority
func HasCurrentUserAuthority(ctx context.Context, authority string) bool {
	authCtx, err := GetAuthenticatedUser(ctx)
	if err != nil {
		return false
	}
	return authCtx.HasAuthority(authority)
}

// HasCurrentUserAnyAuthority checks if the authenticated user has any of the specified authorities
func HasCurrentUserAnyAuthority(ctx context.Context, authorities ...string) bool {
	authCtx, err := GetAuthenticatedUser(ctx)
	if err != nil {
		return false
	}
	return authCtx.HasAnyAuthority(authorities...)
}