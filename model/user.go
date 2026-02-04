package model

import (
	"database/sql"
	"errors"
	"time"
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
