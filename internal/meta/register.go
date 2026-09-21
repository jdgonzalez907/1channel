package meta

import (
	httprouter "github.com/jdgonzalez907/1channel/internal/http"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations"
)

func RegisterRoutes(r *httprouter.Router, cfg config.Configuration, conversationsAPI conversations.ConversationsAPI) {
	r.Handle("GET /meta/webhook", NewVerificationHandler(cfg).Handle)
	r.Handle("POST /meta/webhook", NewMessageEventHandler(cfg, conversationsAPI).Handle)
}
