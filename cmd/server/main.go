package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"sync"

	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/config"
	"42tokyo-road-to-dena-server/internal/handler"
	"42tokyo-road-to-dena-server/internal/repository"
	"42tokyo-road-to-dena-server/internal/service"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/XSAM/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := realMain(); err != nil {
		logger.Error("Error running server", "error", err)
		os.Exit(1)
	}
	logger.Info("Server exited gracefully")
}

func realMain() (err error) {
	ctx := context.Background()
	otelShutdown, err := setupOtelSDK(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup OpenTelemetry SDK: %w", err)
	}
	defer func() {
		err = errors.Join(err, otelShutdown(ctx))
	}()
	logger := slog.New(slogmulti.Fanout(
		otelslog.NewHandler("sports_ticket_app"),
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	))
	slog.SetDefault(logger)

	return run(ctx, logger)
}

func setupDatabase(cfg *config.Config) (*sql.DB, func() error, error) {
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

	db, err := otelsql.Open(dbDriver, dsn, otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	reg, err := otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to register DB stats metrics: %w", err)
	}

	cleanUp := func() error {
		return errors.Join(db.Close(), reg.Unregister())
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(pingCtx); err != nil {
		return nil, cleanUp, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, cleanUp, nil
}


func run(ctx context.Context, logger *slog.Logger) (err error) {
	// 設定の読み込み
	cfg, err := config.Load(logger)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	// DB接続の初期化
	db, cleanUp, err := setupDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}
	defer cleanUp()
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

	gamerepo := repository.NewGameRepository(db, logger)
	gameService := service.NewGameService(gamerepo)

	seatsrepo := repository.NewSeatsRepository(db)
	seatsService := service.NewSeatsService(seatsrepo)

	reservationRepo := repository.NewReservationRepository(db)
	reservationService := service.NewReservationService(
		reservationRepo,
		service.WithHoldTime(cfg.Reservation.ReservationExpiration),
		service.WithMaxSeats(cfg.Reservation.MaxSeats),
	)

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
		logger.InfoContext(ctx, "Starting server", "address", srv.Addr)
		if errServ := srv.ListenAndServe(); errServ != nil && !errors.Is(errServ, http.ErrServerClosed) {
			errCh <- fmt.Errorf("failed to start server: %w", errServ)
		}
	}()

	// シグナルハンドリング
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM) // 監視すべきシグナルを列挙する
	defer stop()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if errRes := reservationService.ExpiredReservations(signalCtx); errRes != nil {
					logger.ErrorContext(ctx, "Error checking expired reservations", "error", errRes)
				}
			case <-signalCtx.Done():
				return
			}
		}
	}()
	select {
	case err = <-errCh:
		return err
	case <-signalCtx.Done():
		logger.InfoContext(ctx, "Shutting down server...")
	}
	// グレースフルシャットダウン
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	select {
	case serveErr := <-errCh:
		return serveErr
	default:
	}
	wg.Wait()
	logger.InfoContext(signalCtx, "Server exited")
	return nil
}
