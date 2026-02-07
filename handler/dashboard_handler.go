package handler

import (
	"database/sql"
	"net/http"
	"numera/middleware"
	"numera/model"
	"numera/pkg/session"
	"numera/views/pages"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type DashboardHandler struct {
	db      *sql.DB
	logger  *logrus.Logger
	session *session.Session
}

func NewDashboardHandler(db *sql.DB, logger *logrus.Logger, session *session.Session) *DashboardHandler {
	return &DashboardHandler{
		db:      db,
		logger:  logger,
		session: session,
	}
}

func (h *DashboardHandler) RegisterRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(h.session))
		r.Use(middleware.WithLogger(h.logger))

		r.Get("/dashboard", h.handleIndex)
	})
}

func (h *DashboardHandler) handleIndex(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context())

	user, err := model.GetUserByID(h.db, GetUserID(r.Context()))
	if err != nil {
		logger.WithError(err).Error("user_fetch_failed")
		http.Error(w, "Failed, please refresh!", http.StatusInternalServerError)
		return
	}

	view(w, r, pages.Dashboard(user.ToView()))
}
