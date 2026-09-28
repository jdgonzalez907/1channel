package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httputil"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb/sqlc"
)

type UserReadHandler struct {
	queries *sqlc.Queries
}

func NewUserReadHandler(queries *sqlc.Queries) *UserReadHandler {
	return &UserReadHandler{queries: queries}
}

func (h *UserReadHandler) Register(r chi.Router) {
	r.Get("/users/{id}", h.handleGetUser)
}

func (h *UserReadHandler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := httputil.RequirePathUUID(w, r, "id", "invalid user id")
	if !ok {
		return
	}

	user, err := h.queries.FindUserByID(r.Context(), pgdb.UUID(userID))
	if err != nil {
		httputil.LookupError(w, r, err, "user not found")
		return
	}

	httputil.JSON(w, http.StatusOK, UserDetailResponse{
		ID:        pgdb.UUIDString(user.ID),
		CreatedAt: pgdb.FromTimestamp(user.CreatedAt),
	})
}
