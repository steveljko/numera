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
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type AccountHandler struct {
	db      *sql.DB
	logger  *logrus.Logger
	session *session.Session
}

func NewAccountHandler(db *sql.DB, logger *logrus.Logger, session *session.Session) *AccountHandler {
	return &AccountHandler{
		db:      db,
		logger:  logger,
		session: session,
	}
}

func (h *AccountHandler) RegisterRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(h.session))
		r.Use(middleware.WithLogger(h.logger))

		r.Get("/accounts", h.handleShowIndex)
		r.Get("/accounts/create", h.handleShowCreate)
		r.Post("/accounts/create", h.handleCreate)

		r.Route("/accounts/{id}", func(r chi.Router) {
			r.Get("/edit", h.handleShowUpdate)
			r.Put("/update", h.handleUpdate)
			// TODO: add confirmation modal before destroy
			r.Delete("/destroy", h.handleDestroy)
		})
	})
}

func (h *AccountHandler) handleShowIndex(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context())
	userID := GetUserID(r.Context())

	logger.WithField("user_id", userID).Debug("fetching_accounts_for_user")

	accounts, err := model.GetAccounstByID(h.db, userID)
	if err != nil {
		logger.WithError(err).WithField("user_id", userID).Error("failed_to_fetch_accounts_by_user_id")
		http.Error(w, "Failed to fetch accounts", http.StatusInternalServerError)
		return
	}

	accountViews := make([]model.AccountView, len(accounts))
	for i, account := range accounts {
		accountViews[i] = account.ToView()
	}

	logger.WithFields(logrus.Fields{
		"user_id":        userID,
		"accounts_count": len(accounts),
	}).Debug("accounts_fetched_successfully")

	view(w, r, pages.AccountSlider(accountViews))
}

func (h *AccountHandler) handleShowCreate(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context())
	userID := GetUserID(r.Context())

	logger.WithField("user_id", userID).Debug("showing_create_account_modal")
	view(w, r, pages.CreateAccountModal())
}

func (h *AccountHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context())
	userID := GetUserID(r.Context())

	logger.WithField("user_id", userID).Debug("processing_account_creation")

	balance, err := decimal.NewFromString(r.FormValue("balance"))
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"user_id":       userID,
			"balance_input": r.FormValue("balance"),
		}).Warn("invalid_balance_format_using_zero")
		balance = decimal.Zero
	}

	input := model.CreateAccountInput{
		Name:                  r.FormValue("name"),
		AccountType:           model.AccountType(r.FormValue("account_type")),
		Balance:               balance,
		Color:                 r.FormValue("color"),
		Currency:              model.Currency(r.FormValue("currency")),
		AllowsNegativeBalance: r.FormValue("allows_negative_balance") == "true",
	}

	logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"account_name": input.Name,
		"account_type": input.AccountType,
		"currency":     input.Currency,
	}).Debug("validating_account_input")

	v := validator.New()
	errors := v.Validate(input)

	if len(errors) > 0 {
		logger.WithFields(logrus.Fields{
			"user_id":      userID,
			"error_count":  len(errors),
			"account_name": input.Name,
		}).Warn("account_creation_validation_failed")
		TriggerErrorToast(w, "Please check the form for errors")
		view(w, r, pages.CreateAccountFormErrors(errors))
		return
	}

	accountID, err := model.CreateAccount(h.db, userID, input)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"user_id":      userID,
			"account_name": input.Name,
			"account_type": input.AccountType,
		}).Error("failed_to_create_account")
		TriggerErrorToast(w, "Failed to create account")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.WithFields(logrus.Fields{
		"account_id":   accountID,
		"user_id":      userID,
		"name":         input.Name,
		"account_type": input.AccountType,
		"balance":      balance.String(),
		"currency":     input.Currency,
	}).Info("account_created_successfully")

	TriggerWithToast(w, "reloadAccounts", ToastSuccess, "Successfully created account!")
}

func (h *AccountHandler) handleShowUpdate(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context())
	userID := GetUserID(r.Context())

	id, err := routeParamAsInt64(r, "id")
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"id":      chi.URLParam(r, "id"),
		}).Error("invalid_account_id_parameter")
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"account_id": id,
	}).Debug("fetching_account_for_edit")

	account, err := model.GetAccountByID(h.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.WithFields(logrus.Fields{
				"user_id":    userID,
				"account_id": id,
			}).Warn("account_not_found")
			http.Error(w, "Account not found", http.StatusNotFound)
			return
		}
		logger.WithError(err).WithFields(logrus.Fields{
			"user_id":    userID,
			"account_id": id,
		}).Error("failed_to_fetch_account_for_edit")
		http.Error(w, "Failed to fetch account", http.StatusInternalServerError)
		return
	}

	if !account.IsOwnedByUserID(userID) {
		logger.WithFields(logrus.Fields{
			"account_id": id,
			"user_id":    userID,
		}).Warn("unauthorized_account_access_attempt")
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"account_id": id,
	}).Debug("showing_edit_account_modal")

	view(w, r, pages.EditAccountModal(account.ToView()))
}

func (h *AccountHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context())
	userID := GetUserID(r.Context())

	id, err := routeParamAsInt64(r, "id")
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"id":      chi.URLParam(r, "id"),
		}).Error("invalid_account_id_parameter")
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"account_id": id,
	}).Debug("processing_account_update")

	account, err := model.GetAccountByID(h.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.WithFields(logrus.Fields{
				"user_id":    userID,
				"account_id": id,
			}).Warn("account_not_found")
			http.Error(w, "Account not found", http.StatusNotFound)
			return
		}
		logger.WithError(err).WithFields(logrus.Fields{
			"user_id":    userID,
			"account_id": id,
		}).Error("failed_to_fetch_account_for_update")
		http.Error(w, "Failed to fetch account", http.StatusInternalServerError)
		return
	}

	if !account.IsOwnedByUserID(userID) {
		logger.WithFields(logrus.Fields{
			"account_id": id,
			"user_id":    userID,
		}).Warn("unauthorized_account_update_attempt")
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	input := model.UpdateAccountInput{
		Name:                  r.FormValue("name"),
		AccountType:           model.AccountType(r.FormValue("account_type")),
		Color:                 r.FormValue("color"),
		Currency:              model.Currency(r.FormValue("currency")),
		AllowsNegativeBalance: r.FormValue("allows_negative_balance") == "on",
		IsActive:              1,
	}

	v := validator.New()
	if validationErrors := v.Validate(input); len(validationErrors) > 0 {
		logger.WithFields(logrus.Fields{
			"user_id":     userID,
			"account_id":  id,
			"error_count": len(validationErrors),
		}).Warn("account_update_validation_failed")
		TriggerErrorToast(w, "Please check the form for errors")
		view(w, r, pages.CreateAccountFormErrors(validationErrors))
		return
	}

	err = model.UpdateAccount(h.db, id, input)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"user_id":      userID,
			"account_id":   id,
			"account_name": input.Name,
		}).Error("failed_to_update_account")
		TriggerErrorToast(w, "Failed to update account")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.WithFields(logrus.Fields{
		"account_id":   id,
		"user_id":      userID,
		"name":         input.Name,
		"account_type": input.AccountType,
	}).Info("account_updated_successfully")

	TriggerWithToast(w, "reloadAccounts", ToastSuccess, "Successfully updated account!")
}

func (h *AccountHandler) handleDestroy(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	logger := middleware.GetLogger(r.Context())

	accountID, err := routeParamAsInt64(r, "id")
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"id":      chi.URLParam(r, "id"),
		}).Error("invalid_account_id_parameter")
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"account_id": accountID,
	}).Debug("processing_account_deletion")

	account, err := model.GetAccountByID(h.db, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.WithFields(logrus.Fields{
				"user_id":    userID,
				"account_id": accountID,
			}).Warn("account_not_found")
			http.Error(w, "Account not found", http.StatusNotFound)
			return
		}
		logger.WithError(err).WithFields(logrus.Fields{
			"user_id":    userID,
			"account_id": accountID,
		}).Error("failed_to_fetch_account_for_deletion")
		http.Error(w, "Failed to fetch account", http.StatusInternalServerError)
		return
	}

	if !account.IsOwnedByUserID(userID) {
		logger.WithFields(logrus.Fields{
			"account_id": accountID,
			"user_id":    userID,
		}).Warn("unauthorized_account_deletion_attempt")
		TriggerErrorToast(w, "You are unauthorized for this action!")
		return
	}

	err = model.DeleteAccount(h.db, accountID)
	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"user_id":    userID,
			"account_id": accountID,
		}).Error("failed_to_delete_account")
		TriggerErrorToast(w, "Failed to delete account")
		return
	}

	logger.WithFields(logrus.Fields{
		"account_id": accountID,
		"user_id":    userID,
	}).Info("account_deleted_successfully")

	TriggerHtmx(w, "reloadAccounts")
}
