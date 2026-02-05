package handler

import (
	"database/sql"
	"net/http"
	"numera/model"
	"numera/pkg/validator"
	"numera/views/pages"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewUserHandler(db *sql.DB, logger *logrus.Logger) *UserHandler {
	return &UserHandler{
		db:     db,
		logger: logger,
	}
}

func (h *UserHandler) RegisterRoutes(r *chi.Mux) {
	r.Get("/register", h.handleShowRegister)
	r.Post("/register", h.handleRegister)
}

// handleShowRegister renders register page
func (h *UserHandler) handleShowRegister(w http.ResponseWriter, r *http.Request) {
	view(w, r, pages.Register())
}

// handleRegister cretes new user
func (h *UserHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	input := model.CreateUserInput{
		Name:            r.PostFormValue("name"),
		Email:           r.PostFormValue("email"),
		Password:        r.PostFormValue("password"),
		PasswordConfirm: r.PostFormValue("passwordconfirm"),
	}

	val := validator.New()
	errors := val.Validate(input)

	user, _ := model.GetUserByEmail(h.db, input.Email)
	if user != nil {
		h.logger.WithFields(logrus.Fields{
			"path":  r.URL.Path,
			"email": input.Email,
		}).Error("email already taken")

		errors = validator.New().AddError(errors, "email", "This email address is already in use")
	}

	if len(errors) > 0 {
		TriggerErrorToast(w, "Please check the form for errors")
		view(w, r, pages.FormErrors(errors))
		return
	}

	err := model.CreateUser(h.db, input)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"path":  r.URL.Path,
			"email": input.Email,
		}).Error("failed to create user")
		TriggerErrorToast(w, "Oops! We ran into an issue creating your account. Let’s try that again.")
	}

	h.logger.WithFields(logrus.Fields{
		"path":  r.URL.Path,
		"email": input.Email,
	}).Info("user successfully created")
	RedirectUsingHtmx(w, "/login")
}
