package middleware

import (
	"context"
	"net/http"
	"strings"

	"hrms/backend/internal/auth"
	"hrms/backend/internal/httpx"
)

type contextKey string

const claimsKey contextKey = "auth_claims"

func Authenticator(tokens auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				httpx.Error(w, http.StatusUnauthorized, "missing authorization header", nil)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				httpx.Error(w, http.StatusUnauthorized, "invalid authorization header", nil)
				return
			}

			claims, err := tokens.ParseAccessToken(parts[1])
			if err != nil {
				httpx.Error(w, http.StatusUnauthorized, "invalid access token", nil)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsKey, claims)))
		})
	}
}

func RequireRoles(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				httpx.Error(w, http.StatusUnauthorized, "missing auth claims", nil)
				return
			}

			if _, exists := allowed[claims.Role]; !exists {
				httpx.Error(w, http.StatusForbidden, "forbidden", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func ClaimsFromContext(ctx context.Context) (*auth.AccessClaims, bool) {
	claims, ok := ctx.Value(claimsKey).(*auth.AccessClaims)
	return claims, ok
}
