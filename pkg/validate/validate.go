// Package validate for validate functions
package validate

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/google/uuid"
)

// validate holds the settings and caches for validating request struct values.
var validate *validator.Validate

// translator is a cache of locale and translation information.
var translator ut.Translator

func init() {
	// Instantiate a validator.
	validate = validator.New()

	// Create a translator for english so the error messages are
	// more human-readable than technical.
	translator, _ = ut.New(en.New(), en.New()).GetTranslator("en")

	// Register the english error messages for use.
	if err := en_translations.RegisterDefaultTranslations(validate, translator); err != nil {
		return
	}

	// Use JSON tag names for errors instead of Go struct names.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// -------------------------------------------------------------------
	// Custom Validations and error messages
	// -------------------------------------------------------------------
	_ = RegisterCustomValidators(validate, translator)
}

// Check validates the provided model against it's declared tags.
func Check(val interface{}) error {

	if err := validate.Struct(val); err != nil {
		// Use a type assertion to get the real error value.
		verrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return fmt.Errorf("validator errors: %s", err)
		}

		var fields FieldErrors
		for _, verror := range verrors {
			field := FieldError{
				Field: verror.Field(),
				Error: verror.Translate(translator),
			}
			fields.FieldError = append(fields.FieldError, field)
		}

		return fields
	}

	return nil
}

// CheckUUID validates that the format of an id is valid.
func CheckUUID(uid string) error {
	if _, err := uuid.Parse(uid); err != nil {
		return errors.New("UUID is not in its proper form")
	}
	return nil
}

// CheckID validates that the format of an id is valid.
func CheckID(id string) error {
	return PositiveInt(id)
}

// CheckSlug validates that format meets strong alias formatting
func CheckSlug(str string) error {
	if err := CheckString(str); err != nil {
		return err
	}

	var re = regexp.MustCompile(`^[a-z0-9]+(?:[_-][a-z0-9]+)*$`)
	if len(re.FindStringIndex(str)) == 0 {
		return errors.New("invalid slug")
	}
	return nil
}

// CheckString validates that the format of an id is valid.
func CheckString(str string) error {
	if str := strings.TrimSpace(str); str == "" {
		return fmt.Errorf("string can not be blank")
	}
	return nil
}

// PositiveInt validates that the format of an id is valid.
func PositiveInt(num string) error {
	i, err := strconv.Atoi(num)
	if err != nil {
		return fmt.Errorf("%s is not a valid number", num)
	}
	if i == 0 {
		return fmt.Errorf("value cannot be zero")
	}
	if i < 1 {
		return fmt.Errorf("value cannot be negative")
	}
	maxInt := 1<<32 - 1
	if i > maxInt {
		return fmt.Errorf("value cannot be greater than %d", maxInt)
	}

	return nil
}
