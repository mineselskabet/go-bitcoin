package bitcoin

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Bitcoin is an integer precision type representing a value in Bitcoin.
// The underlying value is always in satoshis. 0.1 bitcoin is represented
// as 10000000.
// Bitcoin can do arithmetic like an integer.
type Bitcoin int64

const (
	Satoshi  Bitcoin = 1
	MicroBTC Bitcoin = 100 * Satoshi
	MilliBTC Bitcoin = 100000 * Satoshi
	BTC      Bitcoin = 100000000 * Satoshi

	// AllBTC is all the minable bitcoin.
	AllBTC Bitcoin = 20999999*BTC + 97690000*Satoshi
)

// Abs returns the absolute value.
func (v Bitcoin) Abs() Bitcoin {
	if v < 0 {
		return -v
	}

	return v
}

// MarshalText implements encoding.TextMarshaler.
func (v Bitcoin) MarshalText() (text []byte, err error) {
	left, right := v.SplitString(0)

	return []byte(left + "." + right), nil
}

// UnmarshalText imeplemts encoding.TextUnmarshaler.
func (v *Bitcoin) UnmarshalText(text []byte) error {
	decoded, err := Parse(string(text))
	if err != nil {
		return err
	}

	*v = decoded

	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (v *Bitcoin) UnmarshalJSON(in []byte) error {
	if len(in) > 2 && in[len(in)-1] == '"' && in[0] == '"' {
		in = in[1 : len(in)-2]
	}

	err := v.UnmarshalText(in)

	if err != nil {
		var f float64
		err2 := json.Unmarshal(in, &f)
		if err2 == nil {
			asFloat, err2 := Parse(fmt.Sprintf("%.08f", f))
			if err2 != nil {
				return err
			}

			*v = asFloat

			return nil
		}

		return err
	}

	return nil
}

func (v Bitcoin) split(pos int) (int, int) {
	posValue := 1 * Satoshi
	for i := 8; i > pos; i-- {
		posValue *= 10
	}

	right := v % posValue
	left := (v - right) / posValue

	return int(left), int(right.Abs())
}

// SplitString splits the value as two strings at pos. Pos 0
// is the decimal point in BTC. SplitString(0) for 1.055000 BTC
// will result in the values "1" and "055".
func (v Bitcoin) SplitString(pos int) (string, string) {
	left, right := v.split(pos)

	rightFormat := fmt.Sprintf("%%0%dd", 8-pos)

	leftStr := fmt.Sprintf("%d", left)
	rightStr := fmt.Sprintf(rightFormat, right)

	for i := len(rightStr) - 1; i > 0; i-- {
		if rightStr[i] != '0' {
			break
		}

		rightStr = rightStr[0:i]
	}

	return leftStr, rightStr
}

// BTC formats the value as a string expressed in BTC including a unit suffix.
func (v Bitcoin) BTC() string {
	left, right := v.SplitString(0)
	unit := "BTC"

	if right == "0" {
		return fmt.Sprintf("%s %s", left, unit)
	}

	return fmt.Sprintf("%s.%s %s", left, right, unit)
}

// MilliBTC formats the value as a string representing a
// number of MilliBTC including a unit suffix.
func (v Bitcoin) MilliBTC() string {
	left, right := v.SplitString(3)
	unit := "mBTC"

	if right == "0" {
		return fmt.Sprintf("%s %s", left, unit)
	}

	return fmt.Sprintf("%s.%s %s", left, right, unit)
}

// Satoshi formats the value as a string representing a number
// of satoshis including suffix.
func (v Bitcoin) Satoshi() string {
	return fmt.Sprintf("%d sats", v)
}

// String implements fmt.Stringer.
func (v Bitcoin) String() string {
	left := fmt.Sprintf("%d", v)
	right := "0"
	unit := "sats"

	switch {
	case v.Abs() > BTC, v == 0:
		left, right = v.SplitString(0)
		unit = "BTC"

	case v.Abs() > MilliBTC:
		left, right = v.SplitString(3)
		unit = "mBTC"
	}

	if right == "0" {
		return fmt.Sprintf("%s %s", left, unit)
	}

	return fmt.Sprintf("%s.%s %s", left, right, unit)
}

// Parse parses a numeric string representing a value in
// bitcoin. Parse assumes the value is a decimal or
// integer. "1.4" will be parsed as 1.4 BTC. "1" will
// be parsed as 1.0 BTC.
func Parse(in string) (Bitcoin, error) {
	value := 0 * Satoshi
	decimals := false
	negative := false

	mul := 100 * MilliBTC

	add := func(add Bitcoin) {
		if !decimals {
			add *= BTC
			value *= 10
		} else {
			add *= mul
			mul /= 10
		}

		value += add
	}

	runeVal := func(r rune) Bitcoin {
		v := Bitcoin(r) - '0'

		return v
	}

	// Collect whole BTC before the comma.
	for pos, r := range in {

		switch r {
		case '+':
			if pos == 0 {
				break
			} else {
				return 0, errors.New("parse error, stray +")
			}

		case '-':
			if pos == 0 {
				negative = true
			} else {
				return 0, errors.New("parse error, stray -")
			}

		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			add(runeVal(r))

		case '.', ',':
			if decimals {
				return 0, errors.New("parse error, too many decimal points")
			}

			decimals = true

		default:
			return 0, errors.New("parse error, unknown character: " + string(r) + " of '" + in + "'")
		}
	}

	if negative {
		return -value, nil
	}

	return value, nil
}
