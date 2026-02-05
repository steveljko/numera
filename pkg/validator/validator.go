package validator

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

type Validation struct {
	validator *validator.Validate
}

func New() *Validation {
	return &Validation{
		validator: validator.New(),
	}
}

// Validate validates a struct using the underlying validator and returns
// a map of field names to human-readable error messages.
// If the struct passes validation, an empty map is returned.
func (v *Validation) Validate(s interface{}) map[string]string {
	errs := make(map[string]string)

	err := v.validator.Struct(s)
	if err == nil {
		return errs
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			field := fieldName(fe)
			errs[field] = Translate(fe)
		}
	}

	return errs
}

// AddError manually adds an error message to the error map.
func (v *Validation) AddError(errs map[string]string, field, message string) map[string]string {
	if errs == nil {
		errs = make(map[string]string)
	}
	errs[field] = message
	return errs
}
