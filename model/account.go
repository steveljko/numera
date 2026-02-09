package model

import (
	"database/sql"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrAccountNotFound = errors.New("account not found")
)

type AccountType string

const (
	AccountTypeChecking AccountType = "checking"
	AccountTypeSavings  AccountType = "savings"
	AccountTypeCash     AccountType = "cash"
)

type Account struct {
	ID                    int64           `db:"id"`
	Name                  string          `db:"name"`
	AccountType           AccountType     `db:"account_type"`
	Balance               decimal.Decimal `db:"balance"`
	Color                 string          `db:"color"`
	Currency              Currency        `db:"currency"`
	AllowsNegativeBalance bool            `db:"allows_negative_balance"`
	IsActive              int             `db:"is_active"`
	UserID                int64           `db:"user_id"`
	CreatedAt             time.Time       `db:"created_at"`
	UpdatedAt             time.Time       `db:"updated_at"`
}

type AccountView struct {
	ID                    int64           `db:"id"`
	Name                  string          `db:"name"`
	AccountType           AccountType     `db:"account_type"`
	Balance               decimal.Decimal `db:"balance"`
	Color                 string          `db:"color"`
	Currency              Currency        `db:"currency"`
	AllowsNegativeBalance bool            `db:"allows_negative_balance"`
	IsActive              int             `db:"is_active"`
}

func (av *AccountView) GetColorClass() string {
	switch av.Color {
	case "blue":
		return "bg-blue-500"
	case "green":
		return "bg-green-500"
	case "red":
		return "bg-red-500"
	case "purple":
		return "bg-purple-500"
	case "orange":
		return "bg-orange-500"
	case "gray":
		return "bg-gray-500"
	case "yellow":
		return "bg-yellow-500"
	case "pink":
		return "bg-pink-500"
	case "indigo":
		return "bg-indigo-500"
	case "teal":
		return "bg-teal-500"
	default:
		return "bg-blue-500"
	}
}

func (av *AccountView) GetBalanceWithCurrency() string {
	return FormatBalance(av.Balance, av.Currency)
}

func (a *Account) ToView() AccountView {
	return AccountView{
		ID:                    a.ID,
		Name:                  a.Name,
		AccountType:           a.AccountType,
		Balance:               a.Balance,
		Color:                 a.Color,
		Currency:              a.Currency,
		AllowsNegativeBalance: a.AllowsNegativeBalance,
		IsActive:              a.IsActive,
	}
}

func (a *Account) IsOwnedByUserID(userID int64) bool {
	return a.UserID == userID
}

// GetAccountByID gets an account using id
func GetAccountByID(db *sql.DB, id int64) (*Account, error) {
	query := `
		SELECT
			id, name, account_type, balance, color, currency,
			allows_negative_balance, is_active, user_id,
			created_at, updated_at
		FROM accounts WHERE id = ? LIMIT 1
	`
	var account Account
	err := db.QueryRow(query, id).Scan(
		&account.ID,
		&account.Name,
		&account.AccountType,
		&account.Balance,
		&account.Color,
		&account.Currency,
		&account.AllowsNegativeBalance,
		&account.IsActive,
		&account.UserID,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	if account.IsActive == 0 {
		return nil, errors.New("account is inactive")
	}

	return &account, nil
}

// GetAccountsByUserID gets all active accounts for a user
func GetAccounstByID(db *sql.DB, userID int64) ([]Account, error) {
	query := `
		SELECT
			id, name, account_type, balance, color, currency,
			allows_negative_balance, is_active, user_id,
			created_at, updated_at
		FROM accounts
		WHERE user_id = ? AND is_active = 1
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var account Account
		err = rows.Scan(
			&account.ID,
			&account.Name,
			&account.AccountType,
			&account.Balance,
			&account.Color,
			&account.Currency,
			&account.AllowsNegativeBalance,
			&account.IsActive,
			&account.UserID,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

type CreateAccountInput struct {
	Name                  string          `form:"name" validate:"required,min=1,max=100"`
	AccountType           AccountType     `form:"account_type" validate:"required,oneof=checking savings cash"`
	Balance               decimal.Decimal `form:"balance"`
	Color                 string          `form:"color" validate:"required"`
	Currency              Currency        `form:"currency" validate:"required,oneof=EUR USD RSD GBP JPY CHF"`
	AllowsNegativeBalance bool            `form:"allows_negative_balance"`
}

type UpdateAccountInput struct {
	Name                  string      `form:"name" validate:"required,min=1,max=100"`
	AccountType           AccountType `form:"account_type" validate:"required,oneof=checking savings cash"`
	Color                 string      `form:"color" validate:"required"`
	Currency              Currency    `form:"currency" validate:"required,oneof=EUR USD RSD GBP JPY CHF"`
	AllowsNegativeBalance bool        `form:"allows_negative_balance"`
	IsActive              int         `form:"is_active" validate:"oneof=0 1"`
}

// CreateAccount creates a new account for a user
func CreateAccount(db *sql.DB, userID int64, input CreateAccountInput) (int64, error) {
	query := `
		INSERT INTO accounts(
			name, account_type, balance, color, currency, allows_negative_balance, user_id
		) VALUES
			(?, ?, ?, ?, ?, ?, ?)
	`
	result, err := db.Exec(
		query,
		input.Name,
		input.AccountType,
		input.Balance,
		input.Color,
		input.Currency,
		input.AllowsNegativeBalance,
		userID,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateAccount updates an existing account
func UpdateAccount(db *sql.DB, id int64, input UpdateAccountInput) error {
	query := `
		UPDATE accounts
		SET
			name = ?,
			account_type = ?,
			color = ?,
			currency = ?,
			allows_negative_balance = ?,
			is_active = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	result, err := db.Exec(
		query,
		input.Name,
		input.AccountType,
		input.Color,
		input.Currency,
		input.AllowsNegativeBalance,
		input.IsActive,
		id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrAccountNotFound
	}

	return nil
}

// DeleteAccount soft deletes an account
func DeleteAccount(db *sql.DB, id int64) error {
	query := `
		UPDATE accounts
		SET is_active = 0, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrAccountNotFound
	}

	return nil
}
