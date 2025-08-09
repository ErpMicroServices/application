package directives

import (
	"context"
	"errors"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Auth directive implementation
func Auth(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	// Check if user is authenticated
	// This would integrate with your authentication system
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return nil, &gqlerror.Error{
			Message: "Access denied. Authentication required.",
			Extensions: map[string]interface{}{
				"code": "UNAUTHENTICATED",
			},
		}
	}

	return next(ctx)
}

// HasRole directive implementation  
func HasRole(ctx context.Context, obj interface{}, next graphql.Resolver, roles []string) (interface{}, error) {
	userRoles := getUserRolesFromContext(ctx)
	
	// Check if user has required role
	for _, requiredRole := range roles {
		for _, userRole := range userRoles {
			if userRole == requiredRole {
				return next(ctx)
			}
		}
	}

	return nil, &gqlerror.Error{
		Message: fmt.Sprintf("Access denied. Required roles: %v", roles),
		Extensions: map[string]interface{}{
			"code": "INSUFFICIENT_PERMISSIONS",
		},
	}
}

// ReadOnly directive implementation
func ReadOnly(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	// This could be used to mark fields as read-only in different contexts
	return next(ctx)
}

// Complexity directive implementation
func Complexity(ctx context.Context, obj interface{}, next graphql.Resolver, multipliers []string, maximum *int) (interface{}, error) {
	// This would integrate with gqlgen's complexity analysis
	return next(ctx)
}

// Helper functions - these would integrate with your auth system
func getUserIDFromContext(ctx context.Context) string {
	// Extract user ID from context
	// This is a placeholder - implement based on your auth system
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

func getUserRolesFromContext(ctx context.Context) []string {
	// Extract user roles from context  
	// This is a placeholder - implement based on your auth system
	if roles := ctx.Value("user_roles"); roles != nil {
		if roleList, ok := roles.([]string); ok {
			return roleList
		}
	}
	return []string{}
}