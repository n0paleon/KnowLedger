package main

import (
	"KnowLedger/internal/config"
	"KnowLedger/internal/database"
	"KnowLedger/internal/logger"
	"KnowLedger/internal/repository"
	"KnowLedger/internal/server"
	"KnowLedger/internal/server/handler"
	"KnowLedger/internal/service"
	"KnowLedger/internal/storage"
	"KnowLedger/internal/storage/cache"
	"KnowLedger/internal/storage/r2"
	"KnowLedger/internal/workerpool"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var configFile = flag.String("config", "config.yaml", "--config [CONFIG_FILE]")

func init() {
	flag.Parse()
}

func main() {
	app := fx.New(
		fx.StopTimeout(15*time.Minute),
		fx.Provide(
			func() (*config.Config, error) {
				return config.LoadConfig(*configFile)
			},
			func(cfg *config.Config) (*zap.Logger, error) {
				l, err := logger.SetupLogger(cfg.Log.Level, cfg.App.Dev, cfg.Log.Output)
				if err != nil {
					return nil, err
				}
				zap.ReplaceGlobals(l)
				return l, nil
			},

			workerpool.NewWorkerPool,
			func(cfg *config.Config) (*gorm.DB, error) {
				return database.Connect(cfg.Database.DSN)
			},
			func(cfg *config.Config) (storage.FileStorage, error) {
				return r2.NewR2CASStorage(
					cfg.Storage.R2.BucketName, cfg.Storage.R2.AccessKey,
					cfg.Storage.R2.SecretKey, cfg.Storage.R2.APIEndpoint, cfg.Storage.R2.PublicEndpoint,
				)
			},
			cache.NewRedisUniversalClient,

			repository.NewFactRepository,
			repository.NewTagRepository,
			repository.NewAdminRepository,
			repository.NewGCJobRepository,

			service.NewMediaService,
			service.NewFactService,
			func(
				factRepo *repository.FactRepository,
				jobRepo *repository.GCJobRepository,
				storage storage.FileStorage,
				pool *workerpool.Pool,
				log *zap.Logger,
				cfg *config.Config,
			) *service.GCService {
				return service.NewGCService(service.GCServiceConfig{
					FactRepository:      factRepo,
					GCJobRepository:     jobRepo,
					Storage:             storage,
					Pool:                pool,
					Log:                 log,
					Interval:            time.Duration(cfg.GC.IntervalSeconds) * time.Second,
					SimilarityThreshold: cfg.GC.ContentSimilarityThreshold,
					LogRetention:        time.Duration(cfg.GC.LogRetentionDays) * 24 * time.Hour,
				})
			},
			service.NewProfileService,
			service.NewInternalApiService,

			handler.NewAdminHandler,
			handler.NewAdminApiHandler,
			handler.NewPublicHandler,
			handler.NewAuthHandler,
			handler.NewInternalAPIHandler,

			server.NewHttpServer,
		),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.WithOptions(zap.IncreaseLevel(zap.ErrorLevel))}
		}),
		fx.Invoke(registerHooks),
	)

	if err := app.Start(context.Background()); err != nil {
		log.Fatal("failed to start application", zap.Error(err))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	sig := <-quit

	zap.L().Info("signal received, initiating graceful shutdown", zap.String("signal", sig.String()))

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer stopCancel()

	if err := app.Stop(stopCtx); err != nil {
		zap.L().Error("error during shutdown", zap.Error(err))
		os.Exit(1)
	}
}

func registerHooks(
	lc fx.Lifecycle,
	serv *fiber.App,
	cfg *config.Config,
	log *zap.Logger,
	adminHandler *handler.AdminHandler,
	adminApiHandler *handler.AdminApiHandler,
	publicHandler *handler.PublicHandler,
	authHandler *handler.AuthHandler,
	internalApiHandler *handler.InternalAPIHandler,
	pool *workerpool.Pool,
	db *gorm.DB,
	gcService *service.GCService,
	adminRepo *repository.AdminRepository,
	redisClient redis.UniversalClient,
) {
	server.SetupRoutes(serv, adminHandler, adminApiHandler, publicHandler, authHandler, internalApiHandler)

	gracefulCtx, gracefulCancel := context.WithCancel(context.Background())
	monitorCtx, monitorCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(onStartCtx context.Context) error {
			if err := redisClient.Ping(onStartCtx).Err(); err != nil {
				return err
			}

			service.SeedAdminIfEmpty(adminRepo, cfg)

			go gcService.Start()

			go func() {
				log.Info("http server started", zap.String("endpoint", cfg.HTTP.ListenAddr))
				if err := serv.Listen(cfg.HTTP.ListenAddr, fiber.ListenConfig{
					GracefulContext: gracefulCtx,
				}); err != nil {
					log.Error("http server error", zap.Error(err))
				}
			}()

			if cfg.App.Dev {
				go func() {
					ticker := time.NewTicker(10 * time.Second)
					defer ticker.Stop()
					for {
						select {
						case <-ticker.C:
							log.Debug("workerpool stats",
								zap.Int("capacity", pool.Cap()),
								zap.Int("free", pool.Free()),
								zap.Int("running", pool.Running()),
							)
						case <-monitorCtx.Done():
							log.Debug("workerpool monitor stopped")
							return
						}
					}
				}()
			}

			return nil
		},
		OnStop: func(onStopCtx context.Context) error {
			log.Info("shutting down gracefully...")

			monitorCancel()
			gracefulCancel()
			gcService.Stop()
			_ = pool.ReleaseTimeout(5 * time.Second)

			sqlDB, _ := db.DB()
			_ = sqlDB.Close()

			log.Info("shutdown complete")
			return nil
		},
	})
}
