package core

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saegusamayumi1234/hsb-data-sync/internal/config"
	"github.com/saegusamayumi1234/hsb-data-sync/internal/store/postgres"
	"github.com/saegusamayumi1234/hsb-data-sync/internal/store/redis"
)

type App struct {
	Logger     *slog.Logger
	Config     *config.Config
	Redis      *redis.RedisClient
	DB         *pgxpool.Pool
	SharedData *SharedData
	Manager    *Manager
}

func NewApp(ctx context.Context) (*App, error) {
	app := &App{}

	if err := app.prepareLogger(); err != nil {
		return nil, err
	}

	if err := app.prepareConfig(); err != nil {
		return nil, err
	}

	if err := app.prepareRedis(ctx); err != nil {
		return nil, err
	}

	if err := app.preparePostgres(ctx); err != nil {
		return nil, err
	}

	if err := app.prepareManager(); err != nil {
		return nil, err
	}

	return app, nil
}

func (a *App) ShutdownApp(ctx context.Context) error {
	if a == nil {
		return fmt.Errorf("app not initialized")
	}

	a.Manager.Stop()

	// 1. Shutdown manager first (stop accepting requests)
	// if err := a.shutdownManager(ctx); err != nil {
	// 	log.Printf("Error shutting down manager: %v", err)
	// }

	// 2. Close database
	if err := a.closePostgres(); err != nil {
		log.Printf("error closing postgres: %v", err)
	}

	// 3. Close Redis
	if err := a.closeRedis(); err != nil {
		log.Printf("error closing redis: %v", err)
	}

	return nil
}

func (a *App) prepareLogger() error {
	a.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
		// AddSource: true,
    }))

    slog.SetDefault(a.Logger)

	a.Logger.Info("logger initialized successfully")
	return nil
}

func (a *App) prepareConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	a.Config = cfg
	a.Logger.Info("configuration loaded successfully")
	return nil
}

func (a *App) prepareRedis(ctx context.Context) error {
	client, err := redis.New(ctx, a.Config.Redis)
	if err != nil {
		return err
	}

	a.Redis = client
	a.Logger.Info("connected to redis successfully")
	return nil
}

func (a *App) closeRedis() error {
	if a.Redis == nil {
		a.Logger.Info("redis client not initialized, skipping close")
		return nil
	}

	if err := a.Redis.Close(); err != nil {
		return err
	}

	a.Logger.Info("redis client closed successfully")
	return nil
}

func (a *App) preparePostgres(ctx context.Context) error {
	db, err := postgres.New(ctx, a.Config.Postgres)
	if err != nil {
		return err
	}

	a.DB = db.Pool
	a.Logger.Info("connected to postgres successfully")
	return nil
}

func (a *App) closePostgres() error {
	if a.DB != nil {
		a.DB.Close()
	}

	a.Logger.Info("postgres connection pool closed successfully")
	return nil
}


func (a *App) prepareManager() error {
	a.Manager = NewManager(a)

	a.Logger.Info("manager initialized successfully")
	return nil
}
