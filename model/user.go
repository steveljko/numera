package model

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

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
	Currency  Currency  `db:"currency"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
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
			id, name, email, currency, created_at, updated_at 
    FROM users WHERE email = ? LIMIT 1
	`

	var user User
	err := db.QueryRow(query, email).Scan(
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

type CreateUserInput struct {
	Name            string `form:"name" validate:"required,min=3,max=100"`
	Email           string `form:"email" validate:"required,email,max=100"`
	Password        string `form:"password" validate:"required,min=8"`
	PasswordConfirm string `form:"password_confirm" validate:"required,eqfield=Password"`
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
