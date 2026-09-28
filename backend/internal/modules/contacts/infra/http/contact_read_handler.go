package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httputil"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb/sqlc"
)

type ContactReadHandler struct {
	queries *sqlc.Queries
}

func NewContactReadHandler(queries *sqlc.Queries) *ContactReadHandler {
	return &ContactReadHandler{queries: queries}
}

func (h *ContactReadHandler) Register(r chi.Router) {
	r.Get("/contacts/{id}", h.handleGetContact)
}

func (h *ContactReadHandler) handleGetContact(w http.ResponseWriter, r *http.Request) {
	contactID, ok := httputil.RequirePathUUID(w, r, "id", "invalid contact id")
	if !ok {
		return
	}

	contact, err := h.queries.FindContactByID(r.Context(), pgdb.UUID(contactID))
	if err != nil {
		httputil.LookupError(w, r, err, "contact not found")
		return
	}

	httputil.JSON(w, http.StatusOK, ContactResponse{
		ID:         pgdb.UUIDString(contact.ID),
		ExternalID: contact.ExternalContactID,
		CreatedAt:  pgdb.FromTimestamp(contact.CreatedAt),
	})
}
