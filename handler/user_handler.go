package handler

import (
	"database/sql"
	"net/http"
	"numera/middleware"
	"numera/model"
	"numera/pkg/session"
	"numera/pkg/validator"
	"numera/views/pages"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	db      *sql.DB
	logger  *logrus.Logger
	session *session.Session
}

func NewUserHandler(db *sql.DB, logger *logrus.Logger, session *session.Session) *UserHandler {
	return &UserHandler{
		db:      db,
		logger:  logger,
		session: session,
	}
}

func (h *UserHandler) RegisterRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireGuest(h.session))
		r.Use(middleware.WithLogger(h.logger))

		r.Get("/register", h.handleShowRegister)
		r.Post("/register", h.handleRegister)
	})
}

// handleShowRegister renders register page
func (h *UserHandler) handleShowRegister(w http.ResponseWriter, r *http.Request) {
	view(w, r, pages.Register())
}

// handleRegister cretes new user
func (h *UserHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context())

	input := model.CreateUserInput{
		Name:            r.PostFormValue("name"),
		Email:           r.PostFormValue("email"),
		Password:        r.PostFormValue("password"),
		PasswordConfirm: r.PostFormValue("passwordconfirm"),
	}

	v := validator.New()
	errors := v.Validate(input)

	user, _ := model.GetUserByEmail(h.db, input.Email)
	if user != nil {
		logger.WithField("email", input.Email).Error("email_already_taken")

		errors = validator.New().AddError(errors, "email", "This email address is already in use")
	}

	if len(errors) > 0 {
		TriggerErrorToast(w, "Please check the form for errors")
		view(w, r, pages.FormErrors(errors))
		return
	}

	err := model.CreateUser(h.db, input)
	if err != nil {
		logger.WithField("email", input.Email).Error("user_failed_to_create")

		TriggerErrorToast(w, "Oops! We ran into an issue creating your account. Let’s try that again.")
		return
	}

	logger.WithField("email", input.Email).Info("user_successfully_created")
	RedirectUsingHtmx(w, "/login")
}
