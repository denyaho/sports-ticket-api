package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/config"
	"42tokyo-road-to-dena-server/internal/handler"
	"42tokyo-road-to-dena-server/internal/repository"
	"42tokyo-road-to-dena-server/internal/service"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("Error running server", "error", err)
		os.Exit(1)
	}
	logger.Info("Server exited gracefully")
}

func setupDatabase(cfg *config.Config) (*sql.DB, error) {
	dbDriver := cfg.Database.Driver
	dBcfg := cfg.Database
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dBcfg.Host,
		dBcfg.Port,
		dBcfg.User,
		dBcfg.Password,
		dBcfg.Name,
	)

	db, err := sql.Open(dbDriver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

func run(logger *slog.Logger) error {
	// 設定の読み込み
	cfg, err := config.Load(logger)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	// DB接続の初期化
	db, err := setupDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}
	authConfig := &authbundle.AuthConfig{
		JWTSecret:    cfg.Auth.JWTSecret,
		JWTIssuer:    cfg.Auth.JWTIssuer,
		JWTAudience:  cfg.Auth.JWTAudience,
		AccessTTL:    cfg.Auth.AccessTokenTTL,
		RefreshTTL:   cfg.Auth.RefreshTokenTTL,
		CookieDomain: cfg.Auth.CookieDomain,
		CookieSecure: cfg.Auth.CookieSecure,
	}
	// ハンドラーの初期化
	userrepo := repository.NewUserRepository(db)
	userservice := service.NewUserService(userrepo)

	gamerepo := repository.NewGameRepository(db)
	gameService := service.NewGameService(gamerepo)

	seatsrepo := repository.NewSeatsRepository(db)
	seatsService := service.NewSeatsService(seatsrepo)

	reservationRepo := repository.NewReservationRepository(db)
	reservationService := service.NewReservationService(reservationRepo)

	store := authbundle.NewRefreshTokenStore(sqlx.NewDb(db, cfg.Database.Driver))
	authbundle := authbundle.NewAuthBundle(authConfig, store)

	h := handler.New(
		authbundle,
		authConfig,
		userservice,
		gameService,
		seatsService,
		reservationService,
		logger,
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
	errCh := make(chan error, 1)
	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if errServ := srv.ListenAndServe(); errServ != nil && !errors.Is(errServ, http.ErrServerClosed) {
			errCh <- fmt.Errorf("failed to start server: %w", errServ)
		}
	}()

	// シグナルハンドリング
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM) // 監視すべきシグナルを列挙する

	defer stop()
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if errRes := reservationService.ExpiredReservations(ctx); errRes != nil {
					log.Printf("Error checking expired reservations: %v", errRes)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("Shutting down server...")
		stop()
	}
	// グレースフルシャットダウン
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	select {
	case err := <-errCh:
		return err
	default:
	}
	log.Println("Server exited")
	return nil
}
