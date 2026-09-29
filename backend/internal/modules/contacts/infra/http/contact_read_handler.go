package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

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

	contact, err := h.queries.FindContactWithPersonalInformationByID(r.Context(), pgdb.UUID(contactID))
	if err != nil {
		httputil.LookupError(w, r, err, "contact not found")
		return
	}

	httputil.JSON(w, http.StatusOK, contactResponseFromRow(contact))
}

func contactResponseFromRow(row sqlc.FindContactWithPersonalInformationByIDRow) ContactResponse {
	return ContactResponse{
		ID:          pgdb.UUIDString(row.ID),
		ExternalID:  row.ExternalContactID,
		Label:       httputil.ContactLabel(row.FirstName, row.LastName, row.DisplayName, row.ExternalContactID),
		DisplayName: row.DisplayName,
		PersonalInformation: personalInformationFromNullable(
			row.PiID,
			row.IdentificationNumber,
			row.FirstName,
			row.LastName,
			row.PhoneNumber,
			row.Email,
			row.Address,
			row.PiCreatedAt,
			row.PiUpdatedAt,
		),
		CreatedAt: pgdb.FromTimestamp(row.CreatedAt),
	}
}

func personalInformationFromNullable(
	id pgtype.UUID,
	identificationNumber string,
	firstName *string,
	lastName *string,
	phoneNumber *string,
	email *string,
	address *string,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
) *PersonalInformationResponse {
	if !id.Valid {
		return nil
	}

	return &PersonalInformationResponse{
		ID:                   pgdb.UUIDString(id),
		IdentificationNumber: identificationNumber,
		FirstName:            firstName,
		LastName:             lastName,
		PhoneNumber:          phoneNumber,
		Email:                email,
		Address:              address,
		CreatedAt:            pgdb.FromTimestamp(createdAt),
		UpdatedAt:            pgdb.FromTimestamp(updatedAt),
	}
}
