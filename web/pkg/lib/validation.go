package lib

import (
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
)

var (
	validate   *validator.Validate
	translator ut.Translator
)

type customValidator struct {
	validateFunc validator.Func
	translation  string
}

var customValidators = map[string]customValidator{
	"mixed_name": {
		mixedNameValidationFunc,
		"{0} must contain at least one alphanumeric character, hyphen, underscore, or space",
	},
	"optional_mixed_name": {optionalMixedNameValidationFunc, ""},

	"username": {
		usernameValidationFunc,
		"{0} must contain only alphanumeric characters and underscores",
	},

	"slug": {
		slugValidationFunc,
		"{0} must contain only lowercase alphanumeric characters and hyphens",
	},
	"optional_slug": {optionalSlugValidationFunc, ""},

	"workspace_name": {
		workspaceNameValidationFunc,
		"{0} must contain only alphanumeric characters, hyphens, underscores, or spaces",
	},
	"optional_workspace_name": {optionalWorkspaceNameValidationFunc, ""},
}

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}

		return name
	})

	en := en.New()
	uni := ut.New(en, en)

	translator, _ = uni.GetTranslator("en")
	if err := en_translations.RegisterDefaultTranslations(validate, translator); err != nil {
		panic("failed to register default translations: " + err.Error())
	}

	for tag, v := range customValidators {
		if err := validate.RegisterValidation(tag, v.validateFunc); err != nil {
			panic("failed to register custom validator: " + tag + ": " + err.Error())
		}

		if strings.TrimSpace(v.translation) != "" {
			if err := validate.RegisterTranslation("username", translator, func(ut ut.Translator) error {
				return ut.Add(tag, v.translation, true)
			}, func(ut ut.Translator, fe validator.FieldError) string {
				t, _ := ut.T("username", fe.Field())
				return t
			}); err != nil {
				panic("failed to register custom translation: " + tag + ": " + err.Error())
			}
		}
	}
}

var (
	// mixedNameValidationFunc is a custom validation function for mixed name fields matching the regex ^[a-zA-Z0-9-_ ]+$
	mixedNameRegex     = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-_ ]+)?$`)
	usernameRegex      = regexp.MustCompile(`[a-zA-Z0-9_]+`)
	slugRegex          = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	workspaceNameRegex = regexp.MustCompile(`^[a-zA-Z0-9-_ ]+$`)
)

func workspaceNameValidationFunc(fl validator.FieldLevel) bool {
	return workspaceNameRegex.MatchString(strings.TrimSpace(fl.Field().String()))
}

func optionalWorkspaceNameValidationFunc(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return true
	}

	return workspaceNameValidationFunc(fl)
}

func mixedNameValidationFunc(fl validator.FieldLevel) bool {
	return mixedNameRegex.MatchString(strings.TrimSpace(fl.Field().String()))
}

func optionalMixedNameValidationFunc(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return true
	}

	return mixedNameValidationFunc(fl)
}

func usernameValidationFunc(fl validator.FieldLevel) bool {
	return usernameRegex.MatchString(strings.TrimSpace(fl.Field().String()))
}

func slugValidationFunc(fl validator.FieldLevel) bool {
	return slugRegex.MatchString(strings.TrimSpace(fl.Field().String()))
}

func optionalSlugValidationFunc(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return true
	}

	return slugValidationFunc(fl)
}

func ValidateStruct(data any) error {
	err := validate.Struct(data)

	switch e := err.(type) {
	case *validator.InvalidValidationError:
		return apperrors.New("unable to validate data", http.StatusInternalServerError)

	case validator.ValidationErrors:
		validationError := apperrors.NewValidationError()

		for _, v := range e {
			validationError.Add(v.Field(), strings.TrimSpace(strings.Replace(v.Translate(translator), v.Field(), "", 1)))
		}

		return validationError

	default:
		return e
	}
}
