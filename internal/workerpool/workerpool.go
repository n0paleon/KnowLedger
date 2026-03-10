package workerpool

import (
	"KnowLedger/internal/config"

	"github.com/panjf2000/ants/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Pool struct {
	*ants.Pool
}

type Params struct {
	fx.In

	Config *config.Config
	Log    *zap.Logger
}

func NewWorkerPool(p Params) (*Pool, error) {
	pool, err := ants.NewPool(
		p.Config.Worker.PoolSize,
		ants.WithPanicHandler(func(err interface{}) {
			p.Log.Error("worker pool panic recovered", zap.Any("error", err))
		}),
		ants.WithLogger(&antsLogger{log: p.Log}),
		ants.WithPreAlloc(true),
	)
	if err != nil {
		return nil, err
	}
	return &Pool{pool}, nil
}

// antsLogger bridges ants logger ke zap
type antsLogger struct {
	log *zap.Logger
}

func (l *antsLogger) Printf(format string, args ...interface{}) {
	l.log.Sugar().Infof(format, args...)
}
