package http

import (
	"errors"
	"net/http"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/app"
	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httperror"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httputil"
)

type PersonalInformationWriteHandler struct {
	registerContactPersonalInformation app.RegisterContactPersonalInformation
}

func NewPersonalInformationWriteHandler(
	registerContactPersonalInformation app.RegisterContactPersonalInformation,
) *PersonalInformationWriteHandler {
	return &PersonalInformationWriteHandler{
		registerContactPersonalInformation: registerContactPersonalInformation,
	}
}

func (h *PersonalInformationWriteHandler) Register(r chi.Router) {
	r.Put("/contacts/{id}/personal-information", h.handleSave)
}

func (h *PersonalInformationWriteHandler) handleSave(w http.ResponseWriter, r *http.Request) {
	contactID, ok := httputil.RequirePathUUID(w, r, "id", "invalid contact id")
	if !ok {
		return
	}

	var req SavePersonalInformationRequest
	if err := httputil.DecodeJSON(w, r, &req); err != nil {
		httperror.BadRequest(w, r, "invalid request body")
		return
	}

	personalInformation, err := h.registerContactPersonalInformation.Execute(r.Context(), app.RegisterContactPersonalInformationInput{
		ContactID:             contactID,
		PersonalInformationID: uuid.NewV7(),
		IdentificationNumber:  req.IdentificationNumber,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		PhoneNumber:           req.PhoneNumber,
		Email:                 req.Email,
		Address:               req.Address,
		At:                    time.Now().UTC(),
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	httputil.JSON(w, http.StatusOK, personalInformationFromDomain(personalInformation))
}

func (h *PersonalInformationWriteHandler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrContactNotFound):
		httperror.NotFound(w, r, err.Error())
	case domain.IsPersonalInformationValidationError(err):
		httperror.Unprocessable(w, r, err.Error())
	default:
		httperror.Generic(w, r, err)
	}
}

func personalInformationFromDomain(personalInformation *domain.PersonalInformation) PersonalInformationResponse {
	return PersonalInformationResponse{
		ID:                   personalInformation.ID().String(),
		IdentificationNumber: personalInformation.IdentificationNumber(),
		FirstName:            personalInformation.FirstName(),
		LastName:             personalInformation.LastName(),
		PhoneNumber:          personalInformation.PhoneNumber(),
		Email:                personalInformation.Email(),
		Address:              personalInformation.Address(),
		CreatedAt:            personalInformation.CreatedAt(),
		UpdatedAt:            personalInformation.UpdatedAt(),
	}
}
