// Package validate provides functions to validate different resources.
package validate

import (
	"fmt"
	"prisma/tms/rest"
	"strings"
)

const emailRule = "ValidEmail"
const emailAt = "@"

func Email(value string) (errs []rest.ErrorValidation) {

	value = strings.TrimSpace(value)

	if value == "" {
		errs = append(errs, rest.ErrorValidation{
			Property: value,
			Rule:     "Required",
			Message:  "Required non-empty property"})
	}

	if strings.Count(value, emailAt) != 1 {
		errs = append(errs, rest.ErrorValidation{
			Property: value,
			Rule:     emailRule,
			Message:  fmt.Sprintf("Expected one @, got %v", value)})
	}

	return errs
}
