package http

import (
	"errors"
	"net/http"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"

	"github.com/jdgonzalez907/1channel/internal/modules/users/app"
	"github.com/jdgonzalez907/1channel/internal/modules/users/domain"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httperror"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httputil"
)

type UserHandler struct {
	createUser app.CreateUser
}

func NewUserHandler(createUser app.CreateUser) *UserHandler {
	return &UserHandler{createUser: createUser}
}

func (h *UserHandler) Register(r chi.Router) {
	r.Post("/users", h.handleCreateUser)
}

func (h *UserHandler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.createUser.Execute(r.Context(), app.CreateUserInput{
		ID:        uuid.NewV7(),
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, UserResponse{ID: user.ID().String()})
}

func (h *UserHandler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrUserInvalidID):
		httperror.Unprocessable(w, r, err.Error())
	default:
		httperror.Generic(w, r, err)
	}
}
