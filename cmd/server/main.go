package main

import (
	"42tokyo-road-to-dena-server/config"
	"42tokyo-road-to-dena-server/internal/handler"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"42tokyo-road-to-dena-server/internal/repository"
	"42tokyo-road-to-dena-server/internal/service"
	_ "github.com/lib/pq"
	"42tokyo-road-to-dena-server/authbundle"
)

func main() {
	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	//DB接続の初期化
	dbDriver := "postgres"
	DBcfg := cfg.Database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",DBcfg.Host, DBcfg.Port, DBcfg.User, DBcfg.Password, DBcfg.Name)

	authConfig := &authbundle.AuthConfig{
		JWTSecret:  cfg.Auth.JWTSecret,
		JWTIssuer:   cfg.Auth.JWTIssuer,
		JWTAudience: cfg.Auth.JWTAudience,
		AccessTTL: cfg.Auth.AccessTokenTTL,
		RefreshTTL: cfg.Auth.RefreshTokenTTL,
		CookieDomain: cfg.Auth.CookieDomain,
		CookieSecure: cfg.Auth.CookieSecure,
	}

	db, err := sql.Open(dbDriver, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	// ハンドラーの初期化
	userrepo := repository.NewUserRepository(db)
	userservice := service.NewUserService(userrepo)

	gamerepo := repository.NewGameRepository(db)
	gameService := service.NewGameService(gamerepo)

	store := authbundle.NewRefreshTokenStore(sqlx.NewDb(db, dbDriver))
	authbundle := authbundle.NewAuthBundle(authConfig, store)

	h := handler.New(
		authbundle,
		authConfig,
		userservice,
		gameService,
	)
	// HTTPサーバーの設定
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      h.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// サーバーの起動（非同期）
	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// シグナルハンドリング
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)//監視すべきシグナルを列挙する
	<-quit

	log.Println("Shutting down server...")

	// グレースフルシャットダウン
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
