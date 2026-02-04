package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

// view renders a templ component and handles any rendering errors.
func view(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// routeParamAsInt64 extracts a URL parameter from the chi router context and
// attempts to parse it as a 64-bit integer.
//
// It returns 0 and an error if the parameter is missing or is not a valid
func routeParamAsInt64(r *http.Request, key string) (int64, error) {
	val := chi.URLParam(r, key)
	if val == "" {
		return 0, fmt.Errorf("param %s is missing", key)
	}

	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id format: %w", err)
	}

	return id, nil
}
