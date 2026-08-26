package validation

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

type Rule[T any] func(T) error

type Errors map[string]string

func (e Errors) Error() string {
	if len(e) == 0 {
		return ""
	}

	builder := strings.Builder{}
	keys := make([]string, len(e))

	totalLen := 0
	index := 0

	for k, v := range e {
		totalLen += len(k) + len(v) + len(": ")

		keys[index] = k
		index++

	}
	slices.Sort(keys)

	totalLen += len("; ") * (index - 1)

	builder.Grow(totalLen)

	for i, k := range keys {
		if i > 0 {
			_, _ = builder.WriteString("; ")
		}

		_, _ = builder.WriteString(k)
		_, _ = builder.WriteString(": ")
		_, _ = builder.WriteString(e[k])

	}

	return builder.String()
}

func Check[T any](errs Errors, fieldName string, val T, rules ...Rule[T]) {
	if errs == nil {
		errs = Errors{}
	}
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
