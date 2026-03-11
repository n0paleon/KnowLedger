package server

import (
	"KnowLedger/internal/config"
	"KnowLedger/web"
	"io/fs"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/dustin/go-humanize"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/storage/redis/v3"
	"github.com/gofiber/template/django/v4"
	redigo "github.com/redis/go-redis/v9"
)

type StructValidator struct {
	validate *validator.Validate
}

func (v *StructValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

func NewStructValidator() *StructValidator {
	return &StructValidator{
		validate: validator.New(),
	}
}

func humanizeBytes(v interface{}) string {
	if size, ok := v.(int64); ok {
		return humanize.Bytes(uint64(size))
	}
	return ""
}

func NewHttpServer(cfg *config.Config, redisClient redigo.UniversalClient) *fiber.App {
	engine := django.New("web/views", ".html")
	engine.Reload(true)

	engine.AddFunc("humanize_bytes", humanizeBytes)

	if !cfg.App.Dev {
		viewsFS := web.FS
		views, _ := fs.Sub(viewsFS, "views")
		engine = django.NewFileSystem(http.FS(views), ".html")
		engine.Reload(false)
	}

	serv := fiber.New(fiber.Config{
		CaseSensitive:   true,
		Views:           engine,
		ViewsLayout:     "layouts/main",
		AppName:         cfg.App.Name,
		JSONDecoder:     sonic.Unmarshal,
		JSONEncoder:     sonic.Marshal,
		StructValidator: NewStructValidator(),
		BodyLimit:       30 * 1024 * 1024,
	})

	if !cfg.App.Dev {
		staticFS := web.FS
		staticSub, _ := fs.Sub(staticFS, "static")
		serv.Use(static.New("", static.Config{
			FS:       staticSub,
			Browse:   false,
			Compress: true,
			MaxAge:   3600,
		}))

		setupCompressionMiddleware(serv)
	} else {
		serv.Use(static.New("./web/static", static.Config{
			CacheDuration: 0,
		}))
	}

	setupSessionStorage(serv, redisClient)

	return serv
}

func setupCompressionMiddleware(app *fiber.App) {
	compressor := compress.New(compress.Config{
		Level: compress.LevelDefault,
	})
	app.Use(compressor)
}

func setupSessionStorage(app *fiber.App, redisClient redigo.UniversalClient) {
	store := redis.NewFromConnection(redisClient)

	app.Use(session.New(session.Config{
		Storage:         store,
		CookieSecure:    true,           // HTTPS only
		CookieHTTPOnly:  true,           // Prevent XSS
		CookieSameSite:  "Lax",          // CSRF protection
		IdleTimeout:     8 * time.Hour,  // Session timeout, after N-minute of inactivity, session will be auto expire
		AbsoluteTimeout: 48 * time.Hour, // Maximum session life, force expire after N-hours regardless of activity
		Extractor:       extractors.FromCookie("__Host-session_id"),
	}))
}
