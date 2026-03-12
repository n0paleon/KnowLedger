package service

import (
	"KnowLedger/internal/config"
	"KnowLedger/internal/repository"
	"KnowLedger/pkg/utils"
	"context"
	"time"

	"go.uber.org/zap"
)

func SeedAdminIfEmpty(repo *repository.AdminRepository, cfg *config.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := repo.Count(ctx)
	if err != nil || count > 0 {
		if err != nil {
			zap.L().Error("admin repository.Count error", zap.Error(err))
		}
		return
	}

	if cfg.Admin.Username == "" || cfg.Admin.Password == "" {
		zap.L().Warn("No admin account seeded. Set ADMIN_USERNAME and ADMIN_PASSWORD env vars.")
		zap.L().Warn("admin seeder skipped")
		return
	}

	hashed, err := utils.GeneratePasswordHash(cfg.Admin.Password)
	if err != nil {
		zap.L().Fatal("failed to generate password hash", zap.Error(err))
	}

	if _, err := repo.Create(ctx, cfg.Admin.Username, hashed); err != nil {
		zap.L().Fatal("failed to create admin user", zap.Error(err))
	}

	zap.L().Info("admin user created", zap.String("username", cfg.Admin.Username))
}
