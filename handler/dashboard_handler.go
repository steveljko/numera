package handler

import (
	"database/sql"
	"net/http"
	"numera/middleware"
	"numera/model"
	"numera/pkg/session"
	"numera/services"
	"numera/views/pages"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type DashboardHandler struct {
	db              *sql.DB
	logger          *logrus.Logger
	session         *session.Session
	exchangeService *services.ExchangeService
}

func NewDashboardHandler(
	db *sql.DB,
	logger *logrus.Logger,
	session *session.Session,
	exchangeService *services.ExchangeService,
) *DashboardHandler {
	return &DashboardHandler{
		db:              db,
		logger:          logger,
		session:         session,
		exchangeService: exchangeService,
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

	balancesByCurrency, err := model.CalculateBalanceByCurrencies(h.db, user.ID)
	if err != nil {
		logger.WithError(err).WithField("user_id", user.ID).Warn("failed_to_calculate_total_balance")
		http.Error(w, "Failed, please refresh!", http.StatusInternalServerError)
		return
	}

	// converts all balances to user's preferred currency and sum them
	total := decimal.Zero
	for currency, balance := range balancesByCurrency {
		convertedAmount, err := h.exchangeService.ConvertAmount(r.Context(), balance, currency, user.Currency)
		if err != nil {
			logger.WithError(err).
				WithField("user_id", user.ID).
				WithField("from_currency", currency).
				WithField("to_currency", user.Currency).
				Warn("failed_to_convert_currency")
			http.Error(w, "Failed, please refresh!", http.StatusInternalServerError)
			return
		}
		total = total.Add(convertedAmount)
	}

	view(w, r, pages.Dashboard(user.ToViewWithTotalBalance(total)))
}
