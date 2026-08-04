package authbundle

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"42tokyo-road-to-dena-server/internal/apperror"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================================
// 認証機能統合ファイル
// このファイルは、認証関連の全機能を1つにまとめたものです。
// このままでも利用可能ですが、アーキテクチャに合わせて適切に分割することを推奨します
// ============================================================================

// A contextKey is a custom type to avoid context key collisions.
type contextKey string

const UserIDKey contextKey = "userID"

// A AuthLoginRequest はログインリクエストの型です
type AuthLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

const (
	CookieAccessToken  = "access_token"
	CookieRefreshToken = "refresh_token"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

type AuthTokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken はリフレッシュトークンの型です
type RefreshToken struct {
	ID        uuid.UUID    `db:"id"`
	UserID    uuid.UUID    `db:"user_id"`
	TokenHash string       `db:"token_hash"`
	ExpiresAt time.Time    `db:"expires_at"`
	RevokedAt *time.Time   `db:"revoked_at"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

// IsExpired returns true if the refresh token is expired, false otherwise.
func (r *RefreshToken) IsExpired() bool {
	return r.ExpiresAt.Before(time.Now())
}

// IsRevoked returns true if the refresh token is revoked, false otherwise.
func (r *RefreshToken) IsRevoked() bool {
	return r.RevokedAt != nil
}

type AuthConfig struct {
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
	CookieDomain string
	CookieSecure bool
}

type RefreshTokenStore struct {
	db *sqlx.DB
}

// NewRefreshTokenStore returns a new instance of RefreshTokenStore.
func NewRefreshTokenStore(db *sqlx.DB) *RefreshTokenStore {
	return &RefreshTokenStore{db: db}
}

// Create saves a new refresh token.
func (st *RefreshTokenStore) Create(ctx context.Context, token *RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := st.db.ExecContext(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
		token.UpdatedAt,
	)
	return err
}

// GetByTokenHash retrieves a refresh token by its hash.
func (st *RefreshTokenStore) GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	var token RefreshToken
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at, updated_at, deleted_at
		FROM refresh_tokens
		WHERE token_hash = $1
		  AND deleted_at IS NULL
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
	`
	err := st.db.GetContext(ctx, &token, query, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	return &token, nil
}

// RevokeByTokenHash revokes a refresh token by its hash.
func (st *RefreshTokenStore) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE token_hash = $1
		  AND deleted_at IS NULL
		  AND revoked_at IS NULL
	`
	_, err := st.db.ExecContext(ctx, query, tokenHash)
	return err
}

// RevokeByUserID revokes all refresh tokens for a user.
func (st *RefreshTokenStore) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE user_id = $1
		  AND deleted_at IS NULL
		  AND revoked_at IS NULL
	`
	_, err := st.db.ExecContext(ctx, query, userID)
	return err
}

type AuthBundle struct {
	cfg               *AuthConfig
	refreshTokenStore *RefreshTokenStore
}

// NewAuthBundle returns a new instance of AuthBundle.
func NewAuthBundle(cfg *AuthConfig, refreshTokenStore *RefreshTokenStore) *AuthBundle {
	return &AuthBundle{
		cfg:               cfg,
		refreshTokenStore: refreshTokenStore,
	}
}

// A AuthClaims represents the JWT claims for authentication.
type AuthClaims struct {
	UserID uuid.UUID `json:"sub"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new access token for the given user ID.
func (a *AuthBundle) GenerateAccessToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := AuthClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(a.cfg.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    a.cfg.JWTIssuer,
			Audience:  []string{a.cfg.JWTAudience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("%w: failed to sign token: %v", apperror.ErrInternal, err.Error())
	}

	return tokenString, nil
}

// ValidateAccessToken validates the access token and returns the claims if valid.
func (a *AuthBundle) ValidateAccessToken(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", apperror.ErrUnauthorized, token.Header["alg"])
		}
		return []byte(a.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, apperror.ErrUnauthorized
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !ok {
		return nil, apperror.ErrUnauthorized
	}

	// Issuer検証
	if claims.Issuer != a.cfg.JWTIssuer {
		return nil, apperror.ErrUnauthorized
	}

	// Audience検証
	validAudience := slices.Contains(claims.Audience, a.cfg.JWTAudience)
	if !validAudience {
		return nil, apperror.ErrUnauthorized
	}

	return claims, nil
}

// GenerateRefreshToken generates a new refresh token for the given user ID.
func (a *AuthBundle) GenerateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// ランダムトークン生成
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("%w: failed to generate random token: %v", apperror.ErrInternal, err.Error())
	}
	tokenString := hex.EncodeToString(tokenBytes)

	// ハッシュ化
	hash := sha256.Sum256([]byte(tokenString))
	tokenHash := hex.EncodeToString(hash[:])

	// DB保存
	refreshToken := &RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(a.cfg.RefreshTTL),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := a.refreshTokenStore.Create(ctx, refreshToken); err != nil {
		return "", fmt.Errorf("%w: failed to save refresh token: %v", apperror.ErrInternal, err.Error())
	}

	return tokenString, nil
}

// ValidateRefreshToken validates the refresh token and returns the corresponding RefreshToken if valid.
func (a *AuthBundle) ValidateRefreshToken(ctx context.Context, tokenString string) (*RefreshToken, error) {
	// ハッシュ化
	hash := sha256.Sum256([]byte(tokenString))
	tokenHash := hex.EncodeToString(hash[:])

	// DB検索
	token, err := a.refreshTokenStore.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, apperror.ErrUnauthorized
	}

	if token.IsExpired() {
		return nil, apperror.ErrUnauthorized
	}

	if token.IsRevoked() {
		return nil, apperror.ErrUnauthorized
	}

	return token, nil
}

// RotateRefreshToken rotates the refresh token: it validates the old token, revokes it, and generates a new one.
func (a *AuthBundle) RotateRefreshToken(ctx context.Context, oldTokenString string) (string, error) {
	// 旧トークン検証
	oldToken, err := a.ValidateRefreshToken(ctx, oldTokenString)
	if err != nil {
		return "", err
	}

	// 旧トークン失効
	hash := sha256.Sum256([]byte(oldTokenString))
	tokenHash := hex.EncodeToString(hash[:])
	if err = a.refreshTokenStore.RevokeByTokenHash(ctx, tokenHash); err != nil {
		return "", fmt.Errorf("%w: failed to revoke old token: %v", apperror.ErrInternal, err.Error())
	}

	// 新トークン生成
	newToken, err := a.GenerateRefreshToken(ctx, oldToken.UserID)
	if err != nil {
		return "", fmt.Errorf("%w: failed to generate new token: %v", apperror.ErrInternal, err.Error())
	}

	return newToken, nil
}

// HashPassword hashes the given password using bcrypt.
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%w: failed to hash password: %v", apperror.ErrInternal, err.Error())
	}
	return string(hashedBytes), nil
}

// SetAuthCookies sets the access and refresh tokens as HTTP cookies.
//
//nolint:gosec // G124: Secure は環境設定で制御（本番=true、ローカルHTTP開発=false）。HttpOnly/SameSite は静的に設定済み
func SetAuthCookies(w http.ResponseWriter, accessToken, refreshToken string, cfg *AuthConfig) {
	// アクセストークンCookie
	http.SetCookie(w, &http.Cookie{
		Name:     CookieAccessToken,
		Value:    accessToken,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		MaxAge:   int(cfg.AccessTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     CookieRefreshToken,
		Value:    refreshToken,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		MaxAge:   int(cfg.RefreshTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func ReadAuthCookies(_ http.ResponseWriter, r *http.Request) (string, string) {
	accessToken, err := r.Cookie(CookieAccessToken)
	if err != nil {
		accessToken = &http.Cookie{}
	}
	refreshToken, err := r.Cookie(CookieRefreshToken)
	if err != nil {
		refreshToken = &http.Cookie{}
	}
	return accessToken.Value, refreshToken.Value
}

// GetUserIDFromContext retrieves the user ID from the context.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// SetUserIDInContext sets the user ID in the context.
func SetUserIDInContext(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// SwaggerDir はSwagger UIのディレクトリパスです
const (
	SwaggerDir  = "./docs/swagger"
	OpenAPIPath = "./docs/openapi.yaml"
)

// RegisterDocsRoutes registers the routes for serving Swagger UI and OpenAPI YAML.
func RegisterDocsRoutes(mux *http.ServeMux) {
	mux.Handle("GET /swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir(SwaggerDir))))
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, OpenAPIPath)
	})
}
