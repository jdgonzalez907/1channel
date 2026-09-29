package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httperror"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httputil"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb/sqlc"
)

const pathParamIdentificationNumber = "identification_number"

type PersonalInformationReadHandler struct {
	queries *sqlc.Queries
}

func NewPersonalInformationReadHandler(queries *sqlc.Queries) *PersonalInformationReadHandler {
	return &PersonalInformationReadHandler{queries: queries}
}

func (h *PersonalInformationReadHandler) Register(r chi.Router) {
	r.Get("/personal-information/{identification_number}", h.handleFindByIdentificationNumber)
}

func (h *PersonalInformationReadHandler) handleFindByIdentificationNumber(w http.ResponseWriter, r *http.Request) {
	identificationNumber := chi.URLParam(r, pathParamIdentificationNumber)

	if err := domain.ValidateIdentificationNumber(identificationNumber); err != nil {
		httperror.BadRequest(w, r, "invalid identification number")
		return
	}

	row, err := h.queries.FindPersonalInformationByIdentificationNumber(r.Context(), identificationNumber)
	if err != nil {
		httputil.LookupError(w, r, err, "personal information not found")
		return
	}

	httputil.JSON(w, http.StatusOK, personalInformationFromRow(row))
}

func personalInformationFromRow(row sqlc.PersonalInformation) PersonalInformationResponse {
	return PersonalInformationResponse{
		ID:                   pgdb.UUIDString(row.ID),
		IdentificationNumber: row.IdentificationNumber,
		FirstName:            row.FirstName,
		LastName:             row.LastName,
		PhoneNumber:          row.PhoneNumber,
		Email:                row.Email,
		Address:              row.Address,
		CreatedAt:            pgdb.FromTimestamp(row.CreatedAt),
		UpdatedAt:            pgdb.FromTimestamp(row.UpdatedAt),
	}
}
