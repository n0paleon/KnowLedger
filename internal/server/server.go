package server

import (
	"KnowLedger/internal/config"
	"KnowLedger/web"
	"io/fs"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/dustin/go-humanize"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/template/django/v4"
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

func NewHttpServer(cfg *config.Config) *fiber.App {
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
		serv.Use("/*", static.New("", static.Config{
			FS:     staticSub,
			Browse: false,
		}))
	} else {
		serv.Use("/*", static.New("./web/static", static.Config{
			CacheDuration: 0,
		}))
	}

	return serv
}
