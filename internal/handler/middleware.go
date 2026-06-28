package handler

import (
	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/internal/apperror"
	"context"
	"net/http"
	"strings"
)

// AuthRequired はアクセストークンを検証し、context に userID を注入する
func (h *Handler) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := h.extractTokenFromRequest(r)

		// トークンがない場合は認証エラー
		if token == "" {
			h.respondError(w, apperror.ErrUnauthorized, http.StatusUnauthorized)
			return
		}
		// トークンを検証し、claims から userID を取得
		if h.authBundleService == nil {
			h.respondError(w, apperror.ErrUnauthorized, http.StatusUnauthorized)
			return
		}

		// トークンを検証し、claims から userID を取得
		claims, err := h.authBundleService.ValidateAccessToken(token)
		if err != nil {
			h.respondError(w, apperror.ErrUnauthorized, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), authbundle.UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Authorization ヘッダまたは Cookie からトークンを取得する
func (h *Handler) extractTokenFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	cookie, err := r.Cookie("access_token")
	if err == nil {
		return cookie.Value
	}

	return ""
}
