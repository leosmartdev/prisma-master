package validate

// Maritime Mobile Service Identities
// https://www.navcen.uscg.gov/?pageName=mtmmsi#format
//
// Maritime Identification Digits (MID) are three digit identifiers ranging from 201 to 775
// denoting the administration (country) or geographical area of the administration responsible for the ship station so identified.

import (
	"fmt"
	"prisma/tms/rest"
	"regexp"
	"strconv"
)

const mmsiRule = "MMSI"

var patternNumber = regexp.MustCompile("[0-9]")

// MMSI Format
func MMSI(value string) (errs []rest.ErrorValidation) {
	if len(value) != 9 {
		errs = append(errs, rest.ErrorValidation{
			Property: value,
			Rule:     mmsiRule,
			Message:  fmt.Sprintf("The length of the property should be 9, got(%d)", len(value))})
		return errs
	}

	return parseMMSI(value)
}

func mid(arr []rune, start int) (errs []rest.ErrorValidation) {
	s := string(arr[start : start+3])
	number, err := strconv.Atoi(s)
	if err != nil {
		errs = append(errs, rest.ErrorValidation{
			Property: string(arr),
			Rule:     mmsiRule,
			Message:  fmt.Sprintf("expected MID number [201-775], got(%s)", s)})
	}
	if number < 201 || number > 775 {
		errs = append(errs, rest.ErrorValidation{
			Property: string(arr),
			Rule:     mmsiRule,
			Message:  fmt.Sprintf("expected MID number [201-775], got(%s)", s)})
	}
	return errs
}

func xnumbers(arr []rune, start, count int) (errs []rest.ErrorValidation) {
	arrsl := arr[start : start+count]
	for _, v := range arrsl {
		if patternNumber.MatchString(string(v)) {
			continue
		}

		errs = append(errs, rest.ErrorValidation{
			Property: string(arr),
			Rule:     mmsiRule,
			Message:  fmt.Sprintf("expected([0-9]), got(%s)", string(v))})
	}
	return errs
}

func repeated(arr []rune, start, count int, r rune) (errs []rest.ErrorValidation) {
	arrsl := arr[start : start+count]
	for _, v := range arrsl {
		if v == r {
			continue
		}

		errs = append(errs, rest.ErrorValidation{
			Property: string(arr),
			Rule:     mmsiRule,
			Message:  fmt.Sprintf("expected(%s), got(%s)", string(r), string(v))})
	}
	return errs
}

func parseMMSI(value string) (errs []rest.ErrorValidation) {
	arr := []rune(value)

	switch arr[0] {
	case '8': // MID(1), X(4)=5
		errs1 := mid(arr, 1)
		errs2 := xnumbers(arr, 4, 5)
		errs = append(errs, errs1...)
		errs = append(errs, errs2...)
	case '0':
		switch arr[1] {
		case '0':
			switch arr[2] {
			case '9': // 0(1)=2, 9(2)=3, X(5)=4
				errs1 := repeated(arr, 2, 3, '9')
				errs2 := xnumbers(arr, 5, 4)
				errs = append(errs, errs1...)
				errs = append(errs, errs2...)
			default: // 0(1)=1, MID(2), X(5)=4
				errs1 := mid(arr, 2)
				errs2 := xnumbers(arr, 5, 4)
				errs = append(errs, errs1...)
				errs = append(errs, errs2...)
			}
		default: // MID(1), X(4)=5
			errs1 := mid(arr, 1)
			errs2 := xnumbers(arr, 4, 5)
			errs = append(errs, errs1...)
			errs = append(errs, errs2...)
		}
	case '1': // 1(1)=2, MID(3), X(6)=3
		errs1 := repeated(arr, 1, 2, '1')
		errs2 := mid(arr, 3)
		errs3 := xnumbers(arr, 6, 3)
		errs = append(errs, errs1...)
		errs = append(errs, errs2...)
		errs = append(errs, errs3...)
	case '9':
		switch arr[1] {
		case '9': // 9(1)=1, MID(2), X(5)=4
			errs1 := mid(arr, 2)
			errs2 := xnumbers(arr, 5, 4)
			errs = append(errs, errs1...)
			errs = append(errs, errs2...)
		case '8': // 8(1)=1, MID(2), X(5)=4
			errs1 := mid(arr, 2)
			errs2 := xnumbers(arr, 5, 4)
			errs = append(errs, errs1...)
			errs = append(errs, errs2...)
		case '7':
			switch arr[2] {
			// 7(1)=1, 0(2)=1, X(3)=6
			// 7(1)=1, 2(2)=1, X(3)=6
			// 7(1)=1, 4(2)=1, X(3)=6
			case '0', '2', '4':
				errs = xnumbers(arr, 3, 6)
			default:
				errs = append(errs, rest.ErrorValidation{
					Property: string(arr),
					Rule:     mmsiRule,
					Message:  fmt.Sprintf("expected([0,2,4]), got(%s)", string(arr[2]))})
			}
		default:
			errs = append(errs, rest.ErrorValidation{
				Property: string(arr),
				Rule:     mmsiRule,
				Message:  fmt.Sprintf("expected([7-9]), got(%s)", string(arr[1]))})
		}
	default: // MID(0), X(3)=6
		errs1 := mid(arr, 0)
		errs2 := xnumbers(arr, 3, 6)
		errs = append(errs, errs1...)
		errs = append(errs, errs2...)
	}

	return errs
}
