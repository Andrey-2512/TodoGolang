package validation

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Rule[T any] func(T) error

type Errors map[string]string

func (e Errors) Error() string {
	return "validation error"
}

func Check[T any](errs Errors, fieldName string, val T, rules ...Rule[T]) {
	if _, ok := errs[fieldName]; ok {
		return
	}
	for _, rule := range rules {
		if err := rule(val); err != nil {
			errs[fieldName] = err.Error()
			return
		}
	}
}

func Match(regExp *regexp.Regexp, customError string) Rule[*string] {
	return func(value *string) error {
		if value == nil || *value == "" {
			return nil
		}
		if ok := regExp.MatchString(*value); !ok {
			return errors.New(customError)
		}
		return nil
	}
}

func RequiredString(customError string) Rule[*string] {
	return func(value *string) error {
		if value == nil || len(strings.TrimSpace(*value)) == 0 {
			return errors.New(customError)
		}
		return nil
	}
}

func RuneLength(min, max int, customError string) Rule[*string] {
	return func(value *string) error {
		if value == nil || *value == "" {
			return nil
		}

		if count := utf8.RuneCountInString(*value); count > max || count < min {
			return errors.New(customError)
		}

		return nil
	}
}
