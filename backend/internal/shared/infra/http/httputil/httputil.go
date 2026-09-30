package httputil

import (
	"encoding/json"
	"net/http"
	"uuid"

	"github.com/go-chi/chi/v5"

	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httperror"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb"
)

const MaxBodyBytes = 1024 * 1024

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(dst)
}

func JSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(body)
}

func RequirePathUUID(w http.ResponseWriter, r *http.Request, name, message string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		httperror.BadRequest(w, r, message)

		return uuid.Nil(), false
	}

	return id, true
}

func LookupError(w http.ResponseWriter, r *http.Request, err error, notFoundMessage string) {
	if pgdb.IsNoRows(err) {
		httperror.NotFound(w, r, notFoundMessage)
		return
	}

	httperror.Generic(w, r, err)
}
