package handler

import (
	"database/sql"
	"net/http"
	"numera/model"
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
	r.Get("/user/{id}", h.handleGetUserByIDTest)
}

func (h *UserHandler) handleGetUserByIDTest(w http.ResponseWriter, r *http.Request) {
	userID, err := routeParamAsInt64(r, "id")
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"path":  r.URL.Path,
		}).Warn("invalid user ID parameter")
		http.Error(w, "Invalid ID provided", http.StatusBadRequest)
		return
	}

	h.logger.WithField("user_id", userID).Debug("fetching user")

	user, err := model.GetUserByID(h.db, userID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("failed to fetch user from database")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"route_id": userID,
		"id":       user.ID,
		"email":    user.Email,
	}).Info("user fetched successfully")

	view(w, r, pages.Hello(user))
}
