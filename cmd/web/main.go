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
	"KnowLedger/internal/workerpool"
	"context"
	"flag"
	"time"

	"github.com/gofiber/fiber/v3"
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
	fx.New(
		fx.Provide(
			// Config & logger
			func() (*config.Config, error) {
				return config.LoadConfig(*configFile)
			},
			func(cfg *config.Config) (*zap.Logger, error) {
				logOutput := cfg.Log.Output
				l, err := logger.SetupLogger(cfg.Log.Level, cfg.App.Dev, logOutput)
				if err != nil {
					return nil, err
				}
				zap.ReplaceGlobals(l)
				return l, nil
			},

			// Infrastructure
			workerpool.NewWorkerPool, // workerpool
			func(cfg *config.Config) (*gorm.DB, error) {
				return database.Connect(cfg.Database.DSN)
			},
			func(cfg *config.Config) (storage.Storage, error) {
				return storage.NewR2CASStorage(
					cfg.Storage.R2.BucketName, cfg.Storage.R2.AccessKey,
					cfg.Storage.R2.SecretKey, cfg.Storage.R2.APIEndpoint, cfg.Storage.R2.PublicEndpoint,
				)
			},

			// Repositories
			repository.NewFactRepository,
			repository.NewTagRepository,
			repository.NewAdminRepository,

			// Services
			service.NewMediaService,
			service.NewFactService,
			func(factRepo *repository.FactRepository, storage storage.Storage, pool *workerpool.Pool, log *zap.Logger, cfg *config.Config) *service.GCService {
				return service.NewGCService(service.GCServiceConfig{
					FactRepository: factRepo,
					Storage:        storage,
					Pool:           pool,
					Log:            log,
					Interval:       time.Duration(cfg.GC.IntervalSeconds) * time.Second,
				})
			},

			// Handlers
			handler.NewAdminHandler,
			handler.NewAdminApiHandler,
			handler.NewPublicHandler,
			handler.NewAuthHandler,

			// Server
			server.NewHttpServer,
		),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.WithOptions(zap.IncreaseLevel(zap.ErrorLevel))}
		}),
		fx.Invoke(registerHooks),
	).Run()
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
	pool *workerpool.Pool,
	db *gorm.DB,
	gcService *service.GCService,
	adminRepo *repository.AdminRepository,
) {
	server.SetupRoutes(serv, adminHandler, adminApiHandler, publicHandler, authHandler)

	stopMonitor := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			service.SeedAdminIfEmpty(adminRepo, cfg) // seed admin account if empty

			log.Info("pool started", zap.Int("capacity", pool.Cap()))

			go func() {
				log.Info("GC service started")
				gcService.Start()
			}()

			go func() {
				log.Info("http server started", zap.String("endpoint", cfg.HTTP.ListenAddr))
				_ = serv.Listen(cfg.HTTP.ListenAddr, fiber.ListenConfig{
					EnablePrintRoutes: true,
				})
			}()

			go func() {
				if !cfg.App.Dev {
					return
				}

				ticker := time.NewTicker(10 * time.Second)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						log.Debug("workerpool statistics",
							zap.Int("capacity", pool.Cap()),
							zap.Int("free", pool.Free()),
							zap.Int("running", pool.Running()),
						)
					case <-stopMonitor:
						log.Debug("workerpool monitor stopped")
						return
					}
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			close(stopMonitor)

			gcService.Stop()
			_ = pool.ReleaseTimeout(5 * time.Second)
			_ = serv.ShutdownWithTimeout(10 * time.Second)

			sqlDB, _ := db.DB()
			_ = sqlDB.Close()

			log.Info("shutdown complete")
			return nil
		},
	})
}
