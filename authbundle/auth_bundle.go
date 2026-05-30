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
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// エラー定数
var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidInput = errors.New("invalid input")
	ErrInternal     = errors.New("internal server error")
)

// ============================================================================
// 認証機能統合ファイル
// このファイルは、認証関連の全機能を1つにまとめたものです。
// このままでも利用可能ですが、アーキテクチャに合わせて適切に分割することを推奨します
// ============================================================================

// コンテキストで使用するキー
type contextKey string

const UserIDKey contextKey = "userID"

// リクエスト/レスポンス型
type AuthLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

type AuthTokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// エンティティ
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

// リフレッシュトークンの有効期限切れ判定
func (r *RefreshToken) IsExpired() bool {
	return r.ExpiresAt.Before(time.Now())
}

// リフレッシュトークンの失効済み判定
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

// RefreshTokenStore のコンストラクタ
func NewRefreshTokenStore(db *sqlx.DB) *RefreshTokenStore {
	return &RefreshTokenStore{db: db}
}

// リフレッシュトークンの保存
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

// ハッシュでリフレッシュトークンを取得
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
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

// ハッシュでリフレッシュトークンを失効
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

// ユーザーIDの全リフレッシュトークンを失効
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

// AuthBundle のコンストラクタ
func NewAuthBundle(cfg *AuthConfig, refreshTokenStore *RefreshTokenStore) *AuthBundle {
	return &AuthBundle{
		cfg:               cfg,
		refreshTokenStore: refreshTokenStore,
	}
}

// JWTクレーム
type AuthClaims struct {
	UserID uuid.UUID `json:"sub"`
	jwt.RegisteredClaims
}

// アクセストークン生成
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
		return "", fmt.Errorf("%w: failed to sign token: %v", ErrInternal, err)
	}

	return tokenString, nil
}

// アクセストークン検証
func (a *AuthBundle) ValidateAccessToken(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrUnauthorized, token.Header["alg"])
		}
		return []byte(a.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrUnauthorized
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !ok {
		return nil, ErrUnauthorized
	}

	// Issuer検証
	if claims.Issuer != a.cfg.JWTIssuer {
		return nil, ErrUnauthorized
	}

	// Audience検証
	validAudience := false
	for _, aud := range claims.Audience {
		if aud == a.cfg.JWTAudience {
			validAudience = true
			break
		}
	}
	if !validAudience {
		return nil, ErrUnauthorized
	}

	return claims, nil
}

// リフレッシュトークン生成
func (a *AuthBundle) GenerateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// ランダムトークン生成
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("%w: failed to generate random token: %v", ErrInternal, err)
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
		return "", fmt.Errorf("%w: failed to save refresh token: %v", ErrInternal, err)
	}

	return tokenString, nil
}

// リフレッシュトークン検証
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
		return nil, ErrUnauthorized
	}

	if token.IsExpired() {
		return nil, ErrUnauthorized
	}

	if token.IsRevoked() {
		return nil, ErrUnauthorized
	}

	return token, nil
}

// リフレッシュトークンローテーション
func (a *AuthBundle) RotateRefreshToken(ctx context.Context, oldTokenString string) (string, error) {
	// 旧トークン検証
	oldToken, err := a.ValidateRefreshToken(ctx, oldTokenString)
	if err != nil {
		return "", err
	}

	// 旧トークン失効
	hash := sha256.Sum256([]byte(oldTokenString))
	tokenHash := hex.EncodeToString(hash[:])
	if err := a.refreshTokenStore.RevokeByTokenHash(ctx, tokenHash); err != nil {
		return "", fmt.Errorf("%w: failed to revoke old token: %v", ErrInternal, err)
	}

	// 新トークン生成
	newToken, err := a.GenerateRefreshToken(ctx, oldToken.UserID)
	if err != nil {
		return "", fmt.Errorf("%w: failed to generate new token: %v", ErrInternal, err)
	}

	return newToken, nil
}

// パスワードハッシュ生成
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// パスワード検証
func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// Cookie設定
func SetAuthCookies(w http.ResponseWriter, accessToken, refreshToken string, cfg *AuthConfig) {
	// アクセストークンCookie
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		MaxAge:   int(cfg.AccessTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	// リフレッシュトークンCookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		MaxAge:   int(cfg.RefreshTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func ReadAuthCookies(w http.ResponseWriter, r *http.Request) (string, string) {
	accessToken, err := r.Cookie("access_token")
	if err != nil {
		accessToken = &http.Cookie{}
	}
	refreshToken, err := r.Cookie("refresh_token")
	if err != nil {
		refreshToken = &http.Cookie{}
	}
	return accessToken.Value, refreshToken.Value
}

// コンテキストからユーザーIDを取得
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// コンテキストにユーザーIDを設定
func SetUserIDInContext(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// ドキュメント（Swagger/OpenAPI）向けのヘルパー
const (
	SwaggerDir  = "./docs/swagger"
	OpenAPIPath = "./docs/openapi.yaml"
)

// Swagger UI と OpenAPI YAML を提供するハンドラを mux に登録する
func RegisterDocsRoutes(mux *http.ServeMux) {
	mux.Handle("GET /swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir(SwaggerDir))))
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, OpenAPIPath)
	})
}
