package meta

import (
	httprouter "github.com/jdgonzalez907/1channel/internal/http"

	"github.com/jdgonzalez907/1channel/internal/config"
)

func RegisterRoutes(r *httprouter.Router, cfg config.Configuration) {
	r.Handle("GET /meta/webhook", NewVerificationHandler(cfg).Handle)
	r.Handle("POST /meta/webhook", NewEventHandler(cfg).Handle)
}
