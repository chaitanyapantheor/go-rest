package validate

import (
	"fmt"
	"reflect"
	"strings"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidators Custom Validations and error messages
func RegisterCustomValidators(validate *validator.Validate, translator ut.Translator) error {
	// Slugs
	if err := validate.RegisterValidation("slug", IsSlug); err != nil {
		return fmt.Errorf("RegisterValidation: %w", err)
	}
	if err := isSlugCustomError(translator); err != nil {
		return fmt.Errorf("isSlugCustomError: %w", err)
	}

	// Not blank
	if err := validate.RegisterValidation("notblank", NotBlank); err != nil {
		return fmt.Errorf("RegisterValidation: %w", err)
	}
	if err := notBlankCustomError(translator); err != nil {
		return fmt.Errorf("notBlank: %w", err)
	}

	// Headers required
	if err := validate.RegisterValidation("header", headersRequired); err != nil {
		return fmt.Errorf("RegisterValidation: %w", err)
	}
	if err := headersRequiredCustomError(translator); err != nil {
		return fmt.Errorf("header: %w", err)
	}

	return nil
}

// -----------------------------------------------------------------------
// Custom Validations
// -----------------------------------------------------------------------

// IsSlug checks if things are properly formed slugs
// Example: https://you.tools/slugify/
func IsSlug(fl validator.FieldLevel) bool {
	field := fl.Field()

	if err := CheckSlug(field.String()); err != nil {
		return false
	}
	return true
}
func isSlugCustomError(trans ut.Translator) error {
	return validate.RegisterTranslation("slug", trans, func(ut ut.Translator) error {
		return ut.Add("slug", "{0} is not in its proper form", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("slug", fe.Field())
		return t
	})
}

// NotBlank is the validation function for validating if the current field
// has a value or length greater than zero, or is not a space only string.
func NotBlank(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		return len(strings.TrimSpace(field.String())) > 0
	case reflect.Chan, reflect.Map, reflect.Slice, reflect.Array:
		return field.Len() > 0
	case reflect.Ptr, reflect.Interface, reflect.Func:
		return !field.IsNil()
	case reflect.Bool:
		return field.IsValid()
	default:
		return field.IsValid() && field.Interface() != reflect.Zero(field.Type()).Interface()
	}
}

func notBlankCustomError(trans ut.Translator) error {
	return validate.RegisterTranslation("notblank", trans, func(ut ut.Translator) error {
		return ut.Add("notblank", "{0} cannot be blank", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("notblank", fe.Field())
		return t
	})
}

// headersRequired checks if things are properly uuid headers
func headersRequired(fl validator.FieldLevel) bool {
	field := fl.Field()
	_ = field.String()

	return NotBlank(fl)
}
func headersRequiredCustomError(trans ut.Translator) error {
	return validate.RegisterTranslation("header", trans, func(ut ut.Translator) error {
		return ut.Add("header", "{0} is a required header", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("header", fe.Field())
		return t
	})
}
