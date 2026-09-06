package meta

import (
	"crypto/subtle"
	"net/http"

	"github.com/jdgonzalez907/1channel/internal/config"
)

const (
	hubModeParam      = "hub.mode"
	hubVerifyToken    = "hub.verify_token"
	hubChallengeParam = "hub.challenge"
	subscribeMode     = "subscribe"
)

type VerificationHandler struct {
	cfg config.Configuration
}

func NewVerificationHandler(cfg config.Configuration) *VerificationHandler {
	return &VerificationHandler{cfg: cfg}
}

func (h *VerificationHandler) Handle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	mode := query.Get(hubModeParam)
	token := query.Get(hubVerifyToken)
	challenge := query.Get(hubChallengeParam)

	modeMatches := subtle.ConstantTimeCompare([]byte(mode), []byte(subscribeMode)) == 1
	tokenMatches := subtle.ConstantTimeCompare([]byte(token), []byte(h.cfg.OneChannelSecret())) == 1

	if modeMatches && tokenMatches {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}

	http.Error(w, "Forbidden", http.StatusForbidden)
}
