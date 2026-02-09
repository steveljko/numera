package model

import "github.com/shopspring/decimal"

func FormatBalance(amount decimal.Decimal, currency Currency) string {
	formatted := amount.StringFixed(2)
	switch currency {
	case CurrencyUSD:
		return "$" + formatted
	case CurrencyEUR:
		return "€" + formatted
	case CurrencyGBP:
		return "£" + formatted
	case CurrencyRSD:
		return formatted + " дин"
	case CurrencyJPY:
		return "¥" + amount.StringFixed(0)
	case CurrencyCHF:
		return "CHF " + formatted
	default:
		return formatted + " " + string(currency)
	}
}
