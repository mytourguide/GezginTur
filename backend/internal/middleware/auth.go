package middleware

import (
	"net/http"
	"strings"

	"seyahat/backend/internal/service"
)

// Auth, Bearer token dogrular ve claim'leri request context'ine yazar.
func Auth(tokens *service.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				writeAuthError(w)
				return
			}
			claims, err := tokens.ParseAccessToken(strings.TrimPrefix(h, "Bearer "))
			if err != nil {
				writeAuthError(w)
				return
			}
			ctx := service.ContextWithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin, yalnizca admin rolune izin verir (Auth middleware'inden sonra kullanilir).
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := service.ClaimsFrom(r.Context())
		if claims.Role != "admin" {
			http.Error(w, `{"error":"yetkisiz erisim"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeAuthError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	http.Error(w, `{"error":"giris yapmaniz gerekiyor"}`, http.StatusUnauthorized)
}
