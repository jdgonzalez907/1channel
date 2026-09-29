package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/app"
	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
)

func strPtr(value string) *string {
	return &value
}

func newPersonalInformationWriteRouter(register app.RegisterContactPersonalInformation) chi.Router {
	router := chi.NewRouter()
	NewPersonalInformationWriteHandler(register).Register(router)

	return router
}

func doSavePersonalInformation(router chi.Router, contactID, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, "/contacts/"+contactID+"/personal-information", strings.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func TestPersonalInformationWriteHandler(t *testing.T) {
	contactID := uuid.NewV7()
	personalInformationID := uuid.NewV7()
	now := time.Now().UTC()
	repoErr := errors.New("repository failure")

	saved := domain.RehydratePersonalInformation(
		personalInformationID, "12345678", strPtr("Juan"), strPtr("Perez"), nil, nil, nil, now, now,
	)

	tests := []struct {
		name       string
		contactID  string
		body       string
		setup      func(t *testing.T, register *app.MockRegisterContactPersonalInformation)
		wantStatus int
	}{
		{
			name:      "saves and associates the personal information",
			contactID: contactID.String(),
			body:      `{"identification_number":"12345678","first_name":"Juan","last_name":"Perez","email":"juan@example.com"}`,
			setup: func(t *testing.T, register *app.MockRegisterContactPersonalInformation) {
				t.Helper()
				register.On("Execute", mock.Anything, mock.MatchedBy(func(in app.RegisterContactPersonalInformationInput) bool {
					return in.ContactID == contactID &&
						in.PersonalInformationID != uuid.Nil() &&
						in.IdentificationNumber == "12345678" &&
						in.FirstName != nil && *in.FirstName == "Juan" &&
						in.Email != nil && *in.Email == "juan@example.com" &&
						in.At.Location() == time.UTC
				})).Return(saved, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "only the document is required",
			contactID: contactID.String(),
			body:      `{"identification_number":"12345678"}`,
			setup: func(t *testing.T, register *app.MockRegisterContactPersonalInformation) {
				t.Helper()
				register.On("Execute", mock.Anything, mock.MatchedBy(func(in app.RegisterContactPersonalInformationInput) bool {
					return in.IdentificationNumber == "12345678" &&
						in.FirstName == nil && in.LastName == nil && in.PhoneNumber == nil && in.Email == nil && in.Address == nil
				})).Return(domain.RehydratePersonalInformation(personalInformationID, "12345678", nil, nil, nil, nil, nil, now, now), nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "invalid document maps to 422",
			contactID: contactID.String(),
			body:      `{"identification_number":"12.345"}`,
			setup: func(t *testing.T, register *app.MockRegisterContactPersonalInformation) {
				t.Helper()
				register.On("Execute", mock.Anything, mock.Anything).
					Return(nil, errors.Join(app.ErrRegisteringContactPersonalInformation, domain.ErrIdentificationNumberInvalid)).Once()
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:      "too long optional field maps to 422",
			contactID: contactID.String(),
			body:      `{"identification_number":"12345678","email":"x"}`,
			setup: func(t *testing.T, register *app.MockRegisterContactPersonalInformation) {
				t.Helper()
				register.On("Execute", mock.Anything, mock.Anything).
					Return(nil, errors.Join(app.ErrRegisteringContactPersonalInformation, domain.ErrEmailTooLong)).Once()
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:      "contact not found maps to 404",
			contactID: contactID.String(),
			body:      `{"identification_number":"12345678"}`,
			setup: func(t *testing.T, register *app.MockRegisterContactPersonalInformation) {
				t.Helper()
				register.On("Execute", mock.Anything, mock.Anything).
					Return(nil, errors.Join(app.ErrRegisteringContactPersonalInformation, domain.ErrContactNotFound)).Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid contact id maps to 400",
			contactID:  "not-a-uuid",
			body:       `{"identification_number":"12345678"}`,
			setup:      func(t *testing.T, register *app.MockRegisterContactPersonalInformation) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed body maps to 400",
			contactID:  contactID.String(),
			body:       `{"identification_number":`,
			setup:      func(t *testing.T, register *app.MockRegisterContactPersonalInformation) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "generic error maps to 500",
			contactID: contactID.String(),
			body:      `{"identification_number":"12345678"}`,
			setup: func(t *testing.T, register *app.MockRegisterContactPersonalInformation) {
				t.Helper()
				register.On("Execute", mock.Anything, mock.Anything).Return(nil, repoErr).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			register := &app.MockRegisterContactPersonalInformation{}
			tt.setup(t, register)
			router := newPersonalInformationWriteRouter(register)

			// Act
			rec := doSavePersonalInformation(router, tt.contactID, tt.body)

			// Assert
			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantStatus == http.StatusOK {
				var body PersonalInformationResponse
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, personalInformationID.String(), body.ID)
			}

			register.AssertExpectations(t)
		})
	}
}
