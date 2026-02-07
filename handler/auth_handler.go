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
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db      *sql.DB
	logger  *logrus.Logger
	session *session.Session
}

func NewAuthHandler(db *sql.DB, logger *logrus.Logger, session *session.Session) *AuthHandler {
	return &AuthHandler{
		db:      db,
		logger:  logger,
		session: session,
	}
}

func (h *AuthHandler) RegisterRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireGuest(h.session))
		r.Use(middleware.WithLogger(h.logger))

		r.Get("/login", h.handleShowLogin)
		r.Post("/login", h.handleLogin)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(h.session))
		r.Use(middleware.WithLogger(h.logger))

		r.Delete("/logout", h.handleLogout)
	})
}

// handleShowLogin renders login page
func (h *AuthHandler) handleShowLogin(w http.ResponseWriter, r *http.Request) {
	view(w, r, pages.Login())
}

// handleRegister cretes new user
func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context())

	input := model.LoginInput{
		Email:    r.PostFormValue("email"),
		Password: r.PostFormValue("password"),
	}

	v := validator.New()
	errors := v.Validate(input)

	if len(errors) > 0 {
		TriggerErrorToast(w, "Please check the form for errors")
		view(w, r, pages.FormErrors(errors))
		return
	}

	user, err := model.GetUserByEmail(h.db, input.Email)
	if err != nil {
		logger.WithError(err).WithField("email", input.Email).Error("db_lookup_failed")

		TriggerErrorToast(w, "Something went wrong, please try again")
		return
	}

	// used for preventing timing attacks
	dummyHash, _ := bcrypt.GenerateFromPassword([]byte("dummy"), bcrypt.DefaultCost)

	var passwordMatch bool
	if user != nil {
		err := bcrypt.CompareHashAndPassword(
			[]byte(user.Password),
			[]byte(input.Password),
		)
		passwordMatch = (err == nil)
	} else {
		bcrypt.CompareHashAndPassword(
			[]byte(string(dummyHash)),
			[]byte(input.Password),
		)
	}

	if user == nil || !passwordMatch {
		logger.WithField("email", input.Email).Warn("login_attempt_denied")

		errors = v.AddError(errors, "email", "Invalid email or password")
		TriggerErrorToast(w, "Please check the form for errors")
		view(w, r, pages.FormErrors(errors))
		return
	}

	if err := h.session.RenewToken(r.Context()); err != nil {
		logger.WithError(err).Error("session_token_renewable_failed")

		TriggerErrorToast(w, "Something went wrong, please try again")
		return
	}

	logger.Infof("user_logged_in: %d", user.ID)
	h.session.SetUserID(r, user.ID)
	RedirectUsingHtmx(w, "/dashboard")
}

func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if err := h.session.Destroy(r.Context()); err != nil {
		middleware.GetLogger(r.Context()).
			WithError(err).
			Error("session_token_destroy_failed")

		TriggerErrorToast(w, "Failed to logout, please try again!")
		return
	}

	TriggerSuccessToast(w, "Successfully logged out!")
	RedirectUsingHtmx(w, "/login")
}
