package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

// RBACMiddleware provides role-based access control middleware
type RBACMiddleware struct {
	config *RBACConfig
}

// RBACConfig holds configuration for RBAC middleware
type RBACConfig struct {
	DefaultDenyAll      bool
	CaseSensitive      bool
	RoleHierarchy      map[string][]string // Maps higher roles to lower roles they inherit
	AuthorityHierarchy map[string][]string // Maps higher authorities to lower authorities they inherit
}

// DefaultRBACConfig returns default RBAC configuration
func DefaultRBACConfig() *RBACConfig {
	return &RBACConfig{
		DefaultDenyAll: true,
		CaseSensitive:  true,
		RoleHierarchy: map[string][]string{
			"ADMIN":      {"MANAGER", "USER"},
			"MANAGER":    {"USER"},
			"MODERATOR":  {"USER"},
		},
		AuthorityHierarchy: map[string][]string{
			"ADMIN":         {"WRITE", "READ"},
			"WRITE":         {"READ"},
			"SERVICE":       {"READ", "WRITE"},
			"SYSTEM_ADMIN":  {"ADMIN", "WRITE", "READ", "SERVICE"},
		},
	}
}

// NewRBACMiddleware creates a new RBAC middleware
func NewRBACMiddleware(config *RBACConfig) *RBACMiddleware {
	if config == nil {
		config = DefaultRBACConfig()
	}
	
	return &RBACMiddleware{
		config: config,
	}
}

// RequireRole creates middleware that requires one or more specific roles
func (rm *RBACMiddleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := GetAuthContext(r.Context())
			if !ok || !authCtx.Authenticated {
				log.Warn().Str("path", r.URL.Path).Msg("RBAC: No authentication context found")
				rm.forbidden(w, "Authentication required")
				return
			}

			if rm.hasAnyRole(authCtx, roles...) {
				log.Debug().
					Str("path", r.URL.Path).
					Str("subject", authCtx.Subject).
					Strs("required_roles", roles).
					Strs("user_roles", authCtx.Roles).
					Msg("RBAC: Role check passed")
				next.ServeHTTP(w, r)
				return
			}

			log.Warn().
				Str("path", r.URL.Path).
				Str("subject", authCtx.Subject).
				Strs("required_roles", roles).
				Strs("user_roles", authCtx.Roles).
				Msg("RBAC: Role check failed")
			
			rm.forbidden(w, fmt.Sprintf("Required roles: %s", strings.Join(roles, ", ")))
		})
	}
}

// RequireAllRoles creates middleware that requires all specified roles
func (rm *RBACMiddleware) RequireAllRoles(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := GetAuthContext(r.Context())
			if !ok || !authCtx.Authenticated {
				log.Warn().Str("path", r.URL.Path).Msg("RBAC: No authentication context found")
				rm.forbidden(w, "Authentication required")
				return
			}

			if rm.hasAllRoles(authCtx, roles...) {
				log.Debug().
					Str("path", r.URL.Path).
					Str("subject", authCtx.Subject).
					Strs("required_roles", roles).
					Strs("user_roles", authCtx.Roles).
					Msg("RBAC: All roles check passed")
				next.ServeHTTP(w, r)
				return
			}

			log.Warn().
				Str("path", r.URL.Path).
				Str("subject", authCtx.Subject).
				Strs("required_roles", roles).
				Strs("user_roles", authCtx.Roles).
				Msg("RBAC: All roles check failed")
			
			rm.forbidden(w, fmt.Sprintf("Required all roles: %s", strings.Join(roles, ", ")))
		})
	}
}

// RequireAuthority creates middleware that requires one or more specific authorities
func (rm *RBACMiddleware) RequireAuthority(authorities ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := GetAuthContext(r.Context())
			if !ok || !authCtx.Authenticated {
				log.Warn().Str("path", r.URL.Path).Msg("RBAC: No authentication context found")
				rm.forbidden(w, "Authentication required")
				return
			}

			if rm.hasAnyAuthority(authCtx, authorities...) {
				log.Debug().
					Str("path", r.URL.Path).
					Str("subject", authCtx.Subject).
					Strs("required_authorities", authorities).
					Strs("user_authorities", authCtx.Authorities).
					Msg("RBAC: Authority check passed")
				next.ServeHTTP(w, r)
				return
			}

			log.Warn().
				Str("path", r.URL.Path).
				Str("subject", authCtx.Subject).
				Strs("required_authorities", authorities).
				Strs("user_authorities", authCtx.Authorities).
				Msg("RBAC: Authority check failed")
			
			rm.forbidden(w, fmt.Sprintf("Required authorities: %s", strings.Join(authorities, ", ")))
		})
	}
}

// RequireAllAuthorities creates middleware that requires all specified authorities
func (rm *RBACMiddleware) RequireAllAuthorities(authorities ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := GetAuthContext(r.Context())
			if !ok || !authCtx.Authenticated {
				log.Warn().Str("path", r.URL.Path).Msg("RBAC: No authentication context found")
				rm.forbidden(w, "Authentication required")
				return
			}

			if rm.hasAllAuthorities(authCtx, authorities...) {
				log.Debug().
					Str("path", r.URL.Path).
					Str("subject", authCtx.Subject).
					Strs("required_authorities", authorities).
					Strs("user_authorities", authCtx.Authorities).
					Msg("RBAC: All authorities check passed")
				next.ServeHTTP(w, r)
				return
			}

			log.Warn().
				Str("path", r.URL.Path).
				Str("subject", authCtx.Subject).
				Strs("required_authorities", authorities).
				Strs("user_authorities", authCtx.Authorities).
				Msg("RBAC: All authorities check failed")
			
			rm.forbidden(w, fmt.Sprintf("Required all authorities: %s", strings.Join(authorities, ", ")))
		})
	}
}

// RequireRoleOrAuthority creates middleware that requires either a role OR an authority
func (rm *RBACMiddleware) RequireRoleOrAuthority(roles []string, authorities []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := GetAuthContext(r.Context())
			if !ok || !authCtx.Authenticated {
				log.Warn().Str("path", r.URL.Path).Msg("RBAC: No authentication context found")
				rm.forbidden(w, "Authentication required")
				return
			}

			hasRole := len(roles) == 0 || rm.hasAnyRole(authCtx, roles...)
			hasAuthority := len(authorities) == 0 || rm.hasAnyAuthority(authCtx, authorities...)

			if hasRole || hasAuthority {
				log.Debug().
					Str("path", r.URL.Path).
					Str("subject", authCtx.Subject).
					Strs("required_roles", roles).
					Strs("required_authorities", authorities).
					Msg("RBAC: Role or authority check passed")
				next.ServeHTTP(w, r)
				return
			}

			log.Warn().
				Str("path", r.URL.Path).
				Str("subject", authCtx.Subject).
				Strs("required_roles", roles).
				Strs("required_authorities", authorities).
				Msg("RBAC: Role or authority check failed")
			
			rm.forbidden(w, fmt.Sprintf("Required roles: %s OR authorities: %s", 
				strings.Join(roles, ", "), strings.Join(authorities, ", ")))
		})
	}
}

// RequireOwnershipOrRole creates middleware that allows access if user is the owner OR has a specific role
func (rm *RBACMiddleware) RequireOwnershipOrRole(getOwnerIDFromRequest func(*http.Request) string, roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := GetAuthContext(r.Context())
			if !ok || !authCtx.Authenticated {
				log.Warn().Str("path", r.URL.Path).Msg("RBAC: No authentication context found")
				rm.forbidden(w, "Authentication required")
				return
			}

			// Check ownership
			ownerID := getOwnerIDFromRequest(r)
			if ownerID != "" && authCtx.Subject == ownerID {
				log.Debug().
					Str("path", r.URL.Path).
					Str("subject", authCtx.Subject).
					Str("owner_id", ownerID).
					Msg("RBAC: Ownership check passed")
				next.ServeHTTP(w, r)
				return
			}

			// Check roles
			if rm.hasAnyRole(authCtx, roles...) {
				log.Debug().
					Str("path", r.URL.Path).
					Str("subject", authCtx.Subject).
					Strs("required_roles", roles).
					Msg("RBAC: Role check passed (ownership failed)")
				next.ServeHTTP(w, r)
				return
			}

			log.Warn().
				Str("path", r.URL.Path).
				Str("subject", authCtx.Subject).
				Str("owner_id", ownerID).
				Strs("required_roles", roles).
				Msg("RBAC: Ownership and role check failed")
			
			rm.forbidden(w, "Access denied - not owner and insufficient privileges")
		})
	}
}

// hasAnyRole checks if the user has any of the specified roles (considering hierarchy)
func (rm *RBACMiddleware) hasAnyRole(authCtx *AuthContext, roles ...string) bool {
	for _, requiredRole := range roles {
		for _, userRole := range authCtx.Roles {
			if rm.roleMatches(userRole, requiredRole) {
				return true
			}
		}
	}
	return false
}

// hasAllRoles checks if the user has all of the specified roles (considering hierarchy)
func (rm *RBACMiddleware) hasAllRoles(authCtx *AuthContext, roles ...string) bool {
	for _, requiredRole := range roles {
		hasRole := false
		for _, userRole := range authCtx.Roles {
			if rm.roleMatches(userRole, requiredRole) {
				hasRole = true
				break
			}
		}
		if !hasRole {
			return false
		}
	}
	return true
}

// hasAnyAuthority checks if the user has any of the specified authorities (considering hierarchy)
func (rm *RBACMiddleware) hasAnyAuthority(authCtx *AuthContext, authorities ...string) bool {
	for _, requiredAuthority := range authorities {
		for _, userAuthority := range authCtx.Authorities {
			if rm.authorityMatches(userAuthority, requiredAuthority) {
				return true
			}
		}
	}
	return false
}

// hasAllAuthorities checks if the user has all of the specified authorities (considering hierarchy)
func (rm *RBACMiddleware) hasAllAuthorities(authCtx *AuthContext, authorities ...string) bool {
	for _, requiredAuthority := range authorities {
		hasAuthority := false
		for _, userAuthority := range authCtx.Authorities {
			if rm.authorityMatches(userAuthority, requiredAuthority) {
				hasAuthority = true
				break
			}
		}
		if !hasAuthority {
			return false
		}
	}
	return true
}

// roleMatches checks if a user role matches a required role (considering hierarchy)
func (rm *RBACMiddleware) roleMatches(userRole, requiredRole string) bool {
	// Direct match
	if rm.stringEquals(userRole, requiredRole) {
		return true
	}

	// Check if user role inherits the required role through hierarchy
	if inheritedRoles, exists := rm.config.RoleHierarchy[userRole]; exists {
		for _, inheritedRole := range inheritedRoles {
			if rm.stringEquals(inheritedRole, requiredRole) {
				return true
			}
			// Recursive check for multi-level hierarchy
			if rm.roleMatches(inheritedRole, requiredRole) {
				return true
			}
		}
	}

	return false
}

// authorityMatches checks if a user authority matches a required authority (considering hierarchy)
func (rm *RBACMiddleware) authorityMatches(userAuthority, requiredAuthority string) bool {
	// Direct match
	if rm.stringEquals(userAuthority, requiredAuthority) {
		return true
	}

	// Check if user authority inherits the required authority through hierarchy
	if inheritedAuthorities, exists := rm.config.AuthorityHierarchy[userAuthority]; exists {
		for _, inheritedAuthority := range inheritedAuthorities {
			if rm.stringEquals(inheritedAuthority, requiredAuthority) {
				return true
			}
			// Recursive check for multi-level hierarchy
			if rm.authorityMatches(inheritedAuthority, requiredAuthority) {
				return true
			}
		}
	}

	return false
}

// stringEquals compares strings considering case sensitivity configuration
func (rm *RBACMiddleware) stringEquals(a, b string) bool {
	if rm.config.CaseSensitive {
		return a == b
	}
	return strings.EqualFold(a, b)
}

// forbidden sends a forbidden response
func (rm *RBACMiddleware) forbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"error":"forbidden","message":"` + message + `"}`))
}

// CheckAccess checks if the given context has access based on roles and authorities
func (rm *RBACMiddleware) CheckAccess(authCtx *AuthContext, requiredRoles []string, requiredAuthorities []string) bool {
	if authCtx == nil || !authCtx.Authenticated {
		return false
	}

	hasRequiredRole := len(requiredRoles) == 0 || rm.hasAnyRole(authCtx, requiredRoles...)
	hasRequiredAuthority := len(requiredAuthorities) == 0 || rm.hasAnyAuthority(authCtx, requiredAuthorities...)

	return hasRequiredRole && hasRequiredAuthority
}

// GetEffectiveRoles returns all effective roles for a user (including inherited roles)
func (rm *RBACMiddleware) GetEffectiveRoles(userRoles []string) []string {
	effectiveRoles := make(map[string]bool)
	
	for _, role := range userRoles {
		effectiveRoles[role] = true
		rm.addInheritedRoles(role, effectiveRoles)
	}

	result := make([]string, 0, len(effectiveRoles))
	for role := range effectiveRoles {
		result = append(result, role)
	}

	return result
}

// GetEffectiveAuthorities returns all effective authorities for a user (including inherited authorities)
func (rm *RBACMiddleware) GetEffectiveAuthorities(userAuthorities []string) []string {
	effectiveAuthorities := make(map[string]bool)
	
	for _, authority := range userAuthorities {
		effectiveAuthorities[authority] = true
		rm.addInheritedAuthorities(authority, effectiveAuthorities)
	}

	result := make([]string, 0, len(effectiveAuthorities))
	for authority := range effectiveAuthorities {
		result = append(result, authority)
	}

	return result
}

// addInheritedRoles recursively adds inherited roles
func (rm *RBACMiddleware) addInheritedRoles(role string, effectiveRoles map[string]bool) {
	if inheritedRoles, exists := rm.config.RoleHierarchy[role]; exists {
		for _, inheritedRole := range inheritedRoles {
			if !effectiveRoles[inheritedRole] {
				effectiveRoles[inheritedRole] = true
				rm.addInheritedRoles(inheritedRole, effectiveRoles)
			}
		}
	}
}

// addInheritedAuthorities recursively adds inherited authorities
func (rm *RBACMiddleware) addInheritedAuthorities(authority string, effectiveAuthorities map[string]bool) {
	if inheritedAuthorities, exists := rm.config.AuthorityHierarchy[authority]; exists {
		for _, inheritedAuthority := range inheritedAuthorities {
			if !effectiveAuthorities[inheritedAuthority] {
				effectiveAuthorities[inheritedAuthority] = true
				rm.addInheritedAuthorities(inheritedAuthority, effectiveAuthorities)
			}
		}
	}
}