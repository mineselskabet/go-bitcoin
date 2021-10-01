package bitcoin

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

// Amount is an integer precision type representing an amount in Satoshis.
// The underlying value is always in satoshis. 0.1 bitcoin is represented
// as 10000000.
// Amount can do arithmetic like an integer.
// Will typically be used like "amount := 250 * bitcoin.Satoshi" or
// "amount := 1400 * bitcoin.MilliBTC".
type Amount int64

const (
	Satoshi  Amount = 1
	MicroBTC Amount = 100 * Satoshi
	MilliBTC Amount = 1000 * MicroBTC
	BTC      Amount = 1000 * MilliBTC

	// AllBTC is all the minable bitcoin.
	AllBTC Amount = 20999999*BTC + 97690000*Satoshi
)

// Float64 returns the amount as floats of unit. For example
// calling Float64(MilliBTC) on 1.5BTC will return 1500.0.
// Note: This must not be used for calculations as precision
// will be lost.
func (a Amount) Float64(unit Amount) float64 {
	if unit == 0 {
		return math.Inf(1)
	}

	return float64(a) / float64(unit)
}

// Abs returns the absolute value.
func (a Amount) Abs() Amount {
	if a < 0 {
		return -a
	}

	return a
}

// MarshalText implements encoding.TextMarshaler.
func (a Amount) MarshalText() (text []byte, err error) {
	left, right := a.SplitString(BTC)

	return []byte(left + "." + right), nil
}

// UnmarshalText imeplemts encoding.TextUnmarshaler.
func (a *Amount) UnmarshalText(text []byte) error {
	decoded, err := Parse(string(text))
	if err != nil {
		return err
	}

	*a = decoded

	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Amount) UnmarshalJSON(in []byte) error {
	if len(in) > 2 && in[len(in)-1] == '"' && in[0] == '"' {
		in = in[1 : len(in)-2]
	}

	err := a.UnmarshalText(in)

	if err != nil {
		var f float64
		err2 := json.Unmarshal(in, &f)
		if err2 == nil {
			asFloat, err2 := Parse(fmt.Sprintf("%.08f", f))
			if err2 != nil {
				return err
			}

			*a = asFloat

			return nil
		}

		return err
	}

	return nil
}

func (a Amount) split(unit Amount) (int, int) {
	if a == 0 {
		return 0, 0
	}

	if unit == 0 {
		return 0, -1
	}

	right := a % unit
	left := (a - right) / unit

	return int(left), int(right.Abs())
}

// SplitString splits the value as two strings at pos. Pos 0
// is the decimal point in BTC. SplitString(0) for 1.055000 BTC
// will result in the values "1" and "055".
// If amount is not equal to 10^x for some integer value of x,
// the result is undefined.
func (a Amount) SplitString(unit Amount) (string, string) {
	left, right := a.split(unit)

	log10 := int(math.Log10(float64(unit)))
	rightFormat := fmt.Sprintf("%%0%dd", log10)

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

// Format will return a string representing the amount in units
// of unit. For example Format(MilliBTC) returns "1500" for an
// amount of 1.5 BTC.
// If amount is not equal to 10^x for some integer value of x,
// the result is undefined.
func (a Amount) Format(unit Amount) string {
	left, right := a.SplitString(unit)

	if right == "0" {
		return left
	}

	return left + "." + right
}

// String implements fmt.Stringer.
func (a Amount) String() string {
	switch {
	case a.Abs() > BTC, a == 0:
		return a.Format(BTC) + " BTC"

	case a.Abs() > MilliBTC:
		return a.Format(MilliBTC) + " mBTC"

	default:
		return a.Format(Satoshi) + " sats"
	}
}

// Parse parses a numeric string representing a value in
// bitcoin. Parse assumes the value is a decimal or
// integer. "1.4" will be parsed as 1.4 BTC. "1" will
// be parsed as 1.0 BTC.
func Parse(in string) (Amount, error) {
	value := 0 * Satoshi
	decimals := false
	negative := false

	mul := 100 * MilliBTC

	add := func(add Amount) {
		if !decimals {
			add *= BTC
			value *= 10
		} else {
			add *= mul
			mul /= 10
		}

		value += add
	}

	runeVal := func(r rune) Amount {
		v := Amount(r) - '0'

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
