package validator

import (
	"strings"
	"unicode"

	v "github.com/go-playground/validator/v10"
)

// Translate converts validator.FieldError into a
// user-friendly error message based on the validation tag.
func Translate(fe v.FieldError) string {
	field := capitalizeFirst(humanizeField(fe.Field()))
	param := fe.Param()

	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return "Invalid email format"
	case "min":
		return field + " must be at least " + param + " characters"
	case "max":
		return field + " must be at most " + param + " characters"
	case "eqfield":
		return field + " must match " + humanizeField(param)
	case "len":
		return field + " must be exactly " + param + " characters"
	case "gte":
		return field + " must be greater than or equal to " + param
	case "lte":
		return field + " must be less than or equal to " + param
	case "numeric":
		return field + " must be a number"
	default:
		return field + " is invalid"
	}
}

// fieldName returns a lowercase version of the struct field name
func fieldName(fe v.FieldError) string {
	return strings.ToLower(fe.Field())
}

// humanizeField converts a camelCase or PascalCase field name into a human-readable string.
// Example: "FirstName" -> "first name"
func humanizeField(field string) string {
	if field == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(field) + 4)

	for i, r := range field {
		if i > 0 && unicode.IsUpper(r) {
			b.WriteRune(' ')
		}
		b.WriteRune(unicode.ToLower(r))
	}

	return b.String()
}

// capitalizeFirst converts the first character of the string to uppercase.
func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}

	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}

	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
