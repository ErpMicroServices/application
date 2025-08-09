package directives

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/erpmicroservices/auth-go/pkg/middleware"
	"github.com/erpmicroservices/auth-go/pkg/oauth2"
)

// HasRoleDirective implements the @hasRole GraphQL directive
type HasRoleDirective struct {
	oauth2Client *oauth2.Client
}

// NewHasRoleDirective creates a new hasRole directive handler
func NewHasRoleDirective(oauth2Client *oauth2.Client) *HasRoleDirective {
	return &HasRoleDirective{
		oauth2Client: oauth2Client,
	}
}

// HasRoleArgs defines the arguments for the @hasRole directive
type HasRoleArgs struct {
	Role      *string   `json:"role"`
	Roles     *[]string `json:"roles"`
	RequireAll *bool    `json:"requireAll"`
}

// HasRole is the GraphQL directive handler for @hasRole
func HasRole(oauth2Client *oauth2.Client) func(ctx context.Context, obj interface{}, next graphql.Resolver, role *string, roles *[]string, requireAll *bool) (interface{}, error) {
	directive := NewHasRoleDirective(oauth2Client)
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, role *string, roles *[]string, requireAll *bool) (interface{}, error) {
		args := &HasRoleArgs{
			Role:      role,
			Roles:     roles,
			RequireAll: requireAll,
		}
		return directive.Handle(ctx, obj, next, args)
	}
}

// Handle processes the @hasRole directive
func (d *HasRoleDirective) Handle(ctx context.Context, obj interface{}, next graphql.Resolver, args *HasRoleArgs) (interface{}, error) {
	// First ensure user is authenticated
	authCtx, authenticated := middleware.GetAuthContext(ctx)
	if !authenticated || authCtx == nil || !authCtx.Authenticated {
		log.Debug().
			Str("operation", graphql.GetOperationContext(ctx).OperationName).
			Str("field", graphql.GetFieldContext(ctx).Field.Name).
			Msg("GraphQL @hasRole directive: Authentication required")
		
		return nil, &gqlerror.Error{
			Message: "Authentication required",
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": "UNAUTHENTICATED",
			},
		}
	}

	// Build list of required roles
	var requiredRoles []string
	
	if args.Role != nil && *args.Role != "" {
		requiredRoles = append(requiredRoles, *args.Role)
	}
	
	if args.Roles != nil {
		for _, role := range *args.Roles {
			if role != "" {
				requiredRoles = append(requiredRoles, role)
			}
		}
	}

	if len(requiredRoles) == 0 {
		log.Warn().
			Str("operation", graphql.GetOperationContext(ctx).OperationName).
			Str("field", graphql.GetFieldContext(ctx).Field.Name).
			Msg("GraphQL @hasRole directive: No roles specified")
		
		return nil, &gqlerror.Error{
			Message: "No roles specified in @hasRole directive",
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": "INVALID_DIRECTIVE",
			},
		}
	}

	// Check role requirements
	requireAll := args.RequireAll != nil && *args.RequireAll
	var hasAccess bool
	var accessType string

	if requireAll {
		// User must have ALL specified roles
		hasAccess = d.hasAllRoles(authCtx, requiredRoles)
		accessType = "all"
	} else {
		// User must have ANY of the specified roles (default behavior)
		hasAccess = d.hasAnyRole(authCtx, requiredRoles)
		accessType = "any"
	}

	if !hasAccess {
		log.Debug().
			Str("operation", graphql.GetOperationContext(ctx).OperationName).
			Str("field", graphql.GetFieldContext(ctx).Field.Name).
			Str("subject", authCtx.Subject).
			Strs("required_roles", requiredRoles).
			Strs("user_roles", authCtx.Roles).
			Str("access_type", accessType).
			Msg("GraphQL @hasRole directive: Access denied")
		
		return nil, &gqlerror.Error{
			Message: d.buildAccessDeniedMessage(requiredRoles, requireAll),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code":          "INSUFFICIENT_PRIVILEGES",
				"required_roles": requiredRoles,
				"user_roles":     authCtx.Roles,
				"require_all":    requireAll,
			},
		}
	}

	log.Debug().
		Str("operation", graphql.GetOperationContext(ctx).OperationName).
		Str("field", graphql.GetFieldContext(ctx).Field.Name).
		Str("subject", authCtx.Subject).
		Strs("required_roles", requiredRoles).
		Strs("user_roles", authCtx.Roles).
		Str("access_type", accessType).
		Msg("GraphQL @hasRole directive: Access granted")

	return next(ctx)
}

// hasAnyRole checks if the user has any of the specified roles
func (d *HasRoleDirective) hasAnyRole(authCtx *middleware.AuthContext, requiredRoles []string) bool {
	return authCtx.HasAnyRole(requiredRoles...)
}

// hasAllRoles checks if the user has all of the specified roles
func (d *HasRoleDirective) hasAllRoles(authCtx *middleware.AuthContext, requiredRoles []string) bool {
	for _, requiredRole := range requiredRoles {
		if !authCtx.HasRole(requiredRole) {
			return false
		}
	}
	return true
}

// buildAccessDeniedMessage creates a user-friendly access denied message
func (d *HasRoleDirective) buildAccessDeniedMessage(requiredRoles []string, requireAll bool) string {
	if len(requiredRoles) == 1 {
		return "Required role: " + requiredRoles[0]
	}
	
	rolesStr := strings.Join(requiredRoles, ", ")
	if requireAll {
		return "Required roles (all): " + rolesStr
	} else {
		return "Required roles (any): " + rolesStr
	}
}

// HasAuthorityArgs defines the arguments for the @hasAuthority directive
type HasAuthorityArgs struct {
	Authority   *string   `json:"authority"`
	Authorities *[]string `json:"authorities"`
	RequireAll  *bool     `json:"requireAll"`
}

// HasAuthority is the GraphQL directive handler for @hasAuthority
func HasAuthority(oauth2Client *oauth2.Client) func(ctx context.Context, obj interface{}, next graphql.Resolver, authority *string, authorities *[]string, requireAll *bool) (interface{}, error) {
	directive := NewHasRoleDirective(oauth2Client) // Reuse the same directive struct
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, authority *string, authorities *[]string, requireAll *bool) (interface{}, error) {
		args := &HasAuthorityArgs{
			Authority:   authority,
			Authorities: authorities,
			RequireAll:  requireAll,
		}
		return directive.HandleAuthority(ctx, obj, next, args)
	}
}

// HandleAuthority processes the @hasAuthority directive
func (d *HasRoleDirective) HandleAuthority(ctx context.Context, obj interface{}, next graphql.Resolver, args *HasAuthorityArgs) (interface{}, error) {
	// First ensure user is authenticated
	authCtx, authenticated := middleware.GetAuthContext(ctx)
	if !authenticated || authCtx == nil || !authCtx.Authenticated {
		log.Debug().
			Str("operation", graphql.GetOperationContext(ctx).OperationName).
			Str("field", graphql.GetFieldContext(ctx).Field.Name).
			Msg("GraphQL @hasAuthority directive: Authentication required")
		
		return nil, &gqlerror.Error{
			Message: "Authentication required",
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": "UNAUTHENTICATED",
			},
		}
	}

	// Build list of required authorities
	var requiredAuthorities []string
	
	if args.Authority != nil && *args.Authority != "" {
		requiredAuthorities = append(requiredAuthorities, *args.Authority)
	}
	
	if args.Authorities != nil {
		for _, authority := range *args.Authorities {
			if authority != "" {
				requiredAuthorities = append(requiredAuthorities, authority)
			}
		}
	}

	if len(requiredAuthorities) == 0 {
		log.Warn().
			Str("operation", graphql.GetOperationContext(ctx).OperationName).
			Str("field", graphql.GetFieldContext(ctx).Field.Name).
			Msg("GraphQL @hasAuthority directive: No authorities specified")
		
		return nil, &gqlerror.Error{
			Message: "No authorities specified in @hasAuthority directive",
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": "INVALID_DIRECTIVE",
			},
		}
	}

	// Check authority requirements
	requireAll := args.RequireAll != nil && *args.RequireAll
	var hasAccess bool
	var accessType string

	if requireAll {
		// User must have ALL specified authorities
		hasAccess = d.hasAllAuthorities(authCtx, requiredAuthorities)
		accessType = "all"
	} else {
		// User must have ANY of the specified authorities (default behavior)
		hasAccess = d.hasAnyAuthority(authCtx, requiredAuthorities)
		accessType = "any"
	}

	if !hasAccess {
		log.Debug().
			Str("operation", graphql.GetOperationContext(ctx).OperationName).
			Str("field", graphql.GetFieldContext(ctx).Field.Name).
			Str("subject", authCtx.Subject).
			Strs("required_authorities", requiredAuthorities).
			Strs("user_authorities", authCtx.Authorities).
			Str("access_type", accessType).
			Msg("GraphQL @hasAuthority directive: Access denied")
		
		return nil, &gqlerror.Error{
			Message: d.buildAuthorityAccessDeniedMessage(requiredAuthorities, requireAll),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code":               "INSUFFICIENT_PRIVILEGES",
				"required_authorities": requiredAuthorities,
				"user_authorities":     authCtx.Authorities,
				"require_all":          requireAll,
			},
		}
	}

	log.Debug().
		Str("operation", graphql.GetOperationContext(ctx).OperationName).
		Str("field", graphql.GetFieldContext(ctx).Field.Name).
		Str("subject", authCtx.Subject).
		Strs("required_authorities", requiredAuthorities).
		Strs("user_authorities", authCtx.Authorities).
		Str("access_type", accessType).
		Msg("GraphQL @hasAuthority directive: Access granted")

	return next(ctx)
}

// hasAnyAuthority checks if the user has any of the specified authorities
func (d *HasRoleDirective) hasAnyAuthority(authCtx *middleware.AuthContext, requiredAuthorities []string) bool {
	return authCtx.HasAnyAuthority(requiredAuthorities...)
}

// hasAllAuthorities checks if the user has all of the specified authorities
func (d *HasRoleDirective) hasAllAuthorities(authCtx *middleware.AuthContext, requiredAuthorities []string) bool {
	for _, requiredAuthority := range requiredAuthorities {
		if !authCtx.HasAuthority(requiredAuthority) {
			return false
		}
	}
	return true
}

// buildAuthorityAccessDeniedMessage creates a user-friendly access denied message for authorities
func (d *HasRoleDirective) buildAuthorityAccessDeniedMessage(requiredAuthorities []string, requireAll bool) string {
	if len(requiredAuthorities) == 1 {
		return "Required authority: " + requiredAuthorities[0]
	}
	
	authoritiesStr := strings.Join(requiredAuthorities, ", ")
	if requireAll {
		return "Required authorities (all): " + authoritiesStr
	} else {
		return "Required authorities (any): " + authoritiesStr
	}
}

// HasPermissionArgs defines the arguments for a combined role/authority directive
type HasPermissionArgs struct {
	Roles       *[]string `json:"roles"`
	Authorities *[]string `json:"authorities"`
	RequireAll  *bool     `json:"requireAll"`
}

// HasPermission is a combined GraphQL directive that checks both roles and authorities
func HasPermission(oauth2Client *oauth2.Client) func(ctx context.Context, obj interface{}, next graphql.Resolver, roles *[]string, authorities *[]string, requireAll *bool) (interface{}, error) {
	directive := NewHasRoleDirective(oauth2Client)
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, roles *[]string, authorities *[]string, requireAll *bool) (interface{}, error) {
		args := &HasPermissionArgs{
			Roles:       roles,
			Authorities: authorities,
			RequireAll:  requireAll,
		}
		return directive.HandlePermission(ctx, obj, next, args)
	}
}

// HandlePermission processes the @hasPermission directive (combined roles and authorities)
func (d *HasRoleDirective) HandlePermission(ctx context.Context, obj interface{}, next graphql.Resolver, args *HasPermissionArgs) (interface{}, error) {
	// First ensure user is authenticated
	authCtx, authenticated := middleware.GetAuthContext(ctx)
	if !authenticated || authCtx == nil || !authCtx.Authenticated {
		return nil, &gqlerror.Error{
			Message: "Authentication required",
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": "UNAUTHENTICATED",
			},
		}
	}

	// Build lists of required roles and authorities
	var requiredRoles []string
	var requiredAuthorities []string
	
	if args.Roles != nil {
		requiredRoles = *args.Roles
	}
	
	if args.Authorities != nil {
		requiredAuthorities = *args.Authorities
	}

	if len(requiredRoles) == 0 && len(requiredAuthorities) == 0 {
		return nil, &gqlerror.Error{
			Message: "No roles or authorities specified in @hasPermission directive",
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": "INVALID_DIRECTIVE",
			},
		}
	}

	requireAll := args.RequireAll != nil && *args.RequireAll
	
	// Check if user has required permissions
	hasRoleAccess := len(requiredRoles) == 0 || 
		(requireAll && d.hasAllRoles(authCtx, requiredRoles)) ||
		(!requireAll && d.hasAnyRole(authCtx, requiredRoles))
	
	hasAuthorityAccess := len(requiredAuthorities) == 0 || 
		(requireAll && d.hasAllAuthorities(authCtx, requiredAuthorities)) ||
		(!requireAll && d.hasAnyAuthority(authCtx, requiredAuthorities))

	var hasAccess bool
	if requireAll {
		hasAccess = hasRoleAccess && hasAuthorityAccess
	} else {
		hasAccess = hasRoleAccess || hasAuthorityAccess
	}

	if !hasAccess {
		log.Debug().
			Str("operation", graphql.GetOperationContext(ctx).OperationName).
			Str("field", graphql.GetFieldContext(ctx).Field.Name).
			Str("subject", authCtx.Subject).
			Bool("require_all", requireAll).
			Msg("GraphQL @hasPermission directive: Access denied")
		
		return nil, &gqlerror.Error{
			Message: "Insufficient permissions",
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code":                 "INSUFFICIENT_PRIVILEGES",
				"required_roles":       requiredRoles,
				"required_authorities": requiredAuthorities,
				"user_roles":          authCtx.Roles,
				"user_authorities":    authCtx.Authorities,
				"require_all":         requireAll,
			},
		}
	}

	return next(ctx)
}