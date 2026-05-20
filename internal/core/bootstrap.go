package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saegusamayumi1234/hsb-data-sync/internal/config"
	"github.com/saegusamayumi1234/hsb-data-sync/internal/constant"
	"github.com/saegusamayumi1234/hsb-data-sync/internal/domain/hypixel"
	"github.com/saegusamayumi1234/hsb-data-sync/internal/store/postgres"
	"github.com/saegusamayumi1234/hsb-data-sync/internal/store/postgres/repository"
	"github.com/saegusamayumi1234/hsb-data-sync/internal/store/redis"
	"github.com/saegusamayumi1234/hsb-data-sync/internal/store/shared"
)

type App struct {
	Logger     *slog.Logger
	Config     *config.Config
	Redis      *redis.RedisClient
	DB         *pgxpool.Pool
	SharedData *shared.SharedData
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

	if err := app.prepareSharedData(ctx); err != nil {
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

func (a *App) prepareSharedData(ctx context.Context) error {
	a.SharedData = shared.NewSharedData(map[string]any{})

	systemKVRepository := repository.NewPostgresSystemKVRepository(a.DB)

	systemKVRows, err := systemKVRepository.GetValuesByKeys(ctx, []string{
		constant.SystemKVKeys.NeuRepoConstantsPets, 
		constant.SystemKVKeys.HypixelResourcesSkyblockItems,
		constant.SystemKVKeys.HypixelResourcesSkyblockSkills,
	})

	if err != nil {
		a.Logger.Error("error fetching system KV values from database", "error", err)
	} else {
		for _, item := range systemKVRows {
			switch item.Key {
			case constant.SystemKVKeys.NeuRepoConstantsPets:
				shared.Set(a.SharedData, shared.KeyNeuRepoConstantsPets, item.Value)
			case constant.SystemKVKeys.HypixelResourcesSkyblockSkills:
				var skyblockSkillsResponse hypixel.SkyblockSkillsResponse
				if err := json.Unmarshal(item.Value, &skyblockSkillsResponse); err != nil {
					return fmt.Errorf("error unmarshaling skyblock skills response: %w", err)
				}
				shared.Set(a.SharedData, shared.KeySkyblockVersion, skyblockSkillsResponse.Version)
			case constant.SystemKVKeys.HypixelResourcesSkyblockItems:
				lookup, err := hypixel.BuildSkyblockItemsReferenceLookupMap(item.Value)
				if err != nil {
					return fmt.Errorf("error building skyblock items reference lookup map: %w", err)
				}
				shared.Set(a.SharedData, shared.KeySkyblockItemsReferenceLookupMap, lookup)
			default:
				a.Logger.Warn("encountered unknown system KV key", "key", item.Key)
			}
			a.Logger.Info("loaded system KV value into shared data", "key", item.Key)
		}
	}

	a.Logger.Info("shared data store initialized successfully")
	return nil
}
