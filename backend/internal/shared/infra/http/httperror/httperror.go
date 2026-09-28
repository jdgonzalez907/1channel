package httperror

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

const ContentType = "application/problem+json"

type Problem struct {
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

func Write(w http.ResponseWriter, r *http.Request, p Problem) {
	if p.Instance == "" {
		p.Instance = chimiddleware.GetReqID(r.Context())
	}

	w.Header().Set("Content-Type", ContentType)
	w.WriteHeader(p.Status)

	_ = json.NewEncoder(w).Encode(p)
}

func BadRequest(w http.ResponseWriter, r *http.Request, detail string) {
	Write(w, r, Problem{
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: detail,
	})
}

func Unauthorized(w http.ResponseWriter, r *http.Request, detail string) {
	Write(w, r, Problem{
		Title:  "Unauthorized",
		Status: http.StatusUnauthorized,
		Detail: detail,
	})
}

func Forbidden(w http.ResponseWriter, r *http.Request, detail string) {
	Write(w, r, Problem{
		Title:  "Forbidden",
		Status: http.StatusForbidden,
		Detail: detail,
	})
}

func NotFound(w http.ResponseWriter, r *http.Request, detail string) {
	Write(w, r, Problem{
		Title:  "Not Found",
		Status: http.StatusNotFound,
		Detail: detail,
	})
}

func MethodNotAllowed(w http.ResponseWriter, r *http.Request, detail string) {
	Write(w, r, Problem{
		Title:  "Method Not Allowed",
		Status: http.StatusMethodNotAllowed,
		Detail: detail,
	})
}

func Conflict(w http.ResponseWriter, r *http.Request, detail string) {
	Write(w, r, Problem{
		Title:  "Conflict",
		Status: http.StatusConflict,
		Detail: detail,
	})
}

func Unprocessable(w http.ResponseWriter, r *http.Request, detail string) {
	Write(w, r, Problem{
		Title:  "Unprocessable Entity",
		Status: http.StatusUnprocessableEntity,
		Detail: detail,
	})
}

func Generic(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		Write(w, r, Problem{
			Title:  "Gateway Timeout",
			Status: http.StatusGatewayTimeout,
			Detail: "request timed out",
		})
	default:
		Write(w, r, Problem{
			Title:  "Internal Server Error",
			Status: http.StatusInternalServerError,
			Detail: "internal server error",
		})
	}
}
