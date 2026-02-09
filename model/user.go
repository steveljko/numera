package model

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type Currency string

const (
	CurrencyEUR Currency = "EUR"
	CurrencyUSD Currency = "USD"
	CurrencyRSD Currency = "RSD"
	CurrencyGBP Currency = "GBP"
	CurrencyJPY Currency = "JPY"
	CurrencyCHF Currency = "CHF"
)

type User struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	Currency  Currency  `db:"currency"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type UserView struct {
	ID           int64    `db:"id"`
	Name         string   `db:"name"`
	Email        string   `db:"email"`
	Currency     Currency `db:"currency"`
	TotalBalance decimal.Decimal
}

func (u *User) ToView() UserView {
	return UserView{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		Currency: u.Currency,
	}
}

func (u *User) ToViewWithTotalBalance(total decimal.Decimal) UserView {
	return UserView{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		Currency:     u.Currency,
		TotalBalance: total,
	}
}

func (uv *UserView) GetTotalBalanceWithCurrency() string {
	return FormatBalance(uv.TotalBalance, uv.Currency)
}

// GetUserByID gets a user using id
func GetUserByID(db *sql.DB, id int64) (*User, error) {
	query := `
		SELECT
			id, name, email, currency, created_at, updated_at 
    FROM users WHERE id = ? LIMIT 1
	`

	var user User
	err := db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Currency,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail gets a user using id
func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	query := `
		SELECT
			id, name, email, password, currency, created_at, updated_at 
    FROM users WHERE email = ? LIMIT 1
	`

	var user User
	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Currency,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

type CreateUserInput struct {
	Name            string `form:"name" validate:"required,min=3,max=100"`
	Email           string `form:"email" validate:"required,email,max=100"`
	Password        string `form:"password" validate:"required,min=8"`
	PasswordConfirm string `form:"password_confirm" validate:"required,eqfield=Password"`
}

type LoginInput struct {
	Email    string `form:"email" validate:"required,email,max=100"`
	Password string `form:"password" validate:"required,min=8"`
}

// CreateUser hashes the user's password and persists the record to the db
func CreateUser(db *sql.DB, input CreateUserInput) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO
			users (name, email, password, currency)
		VALUES
			(?, ?, ?, ?)
	`
	_, err = db.Exec(query, input.Name, input.Email, hashedPassword, CurrencyUSD)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func CalculateTotalBalanceByUserID(db *sql.DB, userID int64) (decimal.Decimal, error) {
	var total decimal.Decimal
	sumQuery := `SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE user_id = ? AND is_active = 1`
	err := db.QueryRow(sumQuery, userID).Scan(&total)
	if err != nil {
		return decimal.Zero, err
	}

	return total, err
}

type ChangeCurrencyRequest struct {
	Currency Currency `validate:"required,oneof=EUR USD RSD GBP JPY CHF"`
}

func ChangeCurrencyByUserID(db *sql.DB, userID int64, currency Currency) error {
	query := `
		UPDATE users
		SET currency = ?
		WHERE id = ?
	`

	result, err := db.Exec(query, currency, userID)
	if err != nil {
		return fmt.Errorf("failed to change currency for user: %w", err)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	return nil
}
