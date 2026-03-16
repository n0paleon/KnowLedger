package utils

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

func ValidationErrorMessages(err error) []string {
	var valErrs validator.ValidationErrors
	if !errors.As(err, &valErrs) {
		return []string{err.Error()}
	}

	msgs := make([]string, 0, len(valErrs))
	for _, e := range valErrs {
		msgs = append(msgs, fmt.Errorf(
			"field '%s' failed validation '%s' (value: %v)",
			e.Field(), e.Tag(), e.Value(),
		).Error())
	}
	return msgs
}
