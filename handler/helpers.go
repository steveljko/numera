package handler

import (
	"context"
	"encoding/json"
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

func GetUserID(ctx context.Context) int64 {
	userID, ok := ctx.Value("USER_ID").(int64)
	if !ok {
		return 0
	}
	return userID
}

func IsHtmx(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func RedirectUsingHtmx(w http.ResponseWriter, url string) {
	w.Header().Add("HX-Redirect", url)
}

func TriggerHtmx(w http.ResponseWriter, event string) {
	w.Header().Add("HX-Trigger", event)
}

// ToastType represents the type of toast notification
type ToastType string

const (
	ToastSuccess ToastType = "success"
	ToastError   ToastType = "error"
	ToastWarning ToastType = "warning"
	ToastInfo    ToastType = "info"
)

// ToastData contains the data for the toast event
type ToastData struct {
	Type       ToastType `json:"type"`
	Text       string    `json:"text"`
	Descripton string    `json:"description,omitempty"`
}

// TriggerToast sends an HX-Trigger header to show a toast notification
func TriggerToast(w http.ResponseWriter, toastType ToastType, text string, description string) {
	data := ToastData{
		Type:       toastType,
		Text:       text,
		Descripton: description,
	}

	triggerData := map[string]ToastData{
		"toast": data,
	}

	jsonData, err := json.Marshal(triggerData)
	if err != nil {
		w.Header().Set("HX-Trigger", "toast")
		return
	}

	w.Header().Set("HX-Trigger", string(jsonData))
}

// TriggerSuccessToast sends a success-themed toast notification to the client
func TriggerSuccessToast(w http.ResponseWriter, text string) {
	TriggerToast(w, ToastSuccess, text, "")
}

// TriggerErrorToast sends an error-themed toast notification to the client
func TriggerErrorToast(w http.ResponseWriter, text string) {
	TriggerToast(w, ToastError, text, "")
}

// TriggerWarningToast sends a warning-themed toast notification to the client
func TriggerWarningToast(w http.ResponseWriter, text string) {
	TriggerToast(w, ToastWarning, text, "")
}

// TriggerInfoToast sends a information-themed toast notification to the client
func TriggerInfoToast(w http.ResponseWriter, text string) {
	TriggerToast(w, ToastInfo, text, "")
}
