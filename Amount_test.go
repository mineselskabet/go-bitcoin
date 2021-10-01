package bitcoin

import (
	"fmt"
	"math"
	"testing"
)

var (
	max = fmt.Sprintf("%d.0", math.MaxInt64/BTC)
	min = fmt.Sprintf("%d.0", math.MinInt64/BTC)
)

func TestParse(t *testing.T) {
	cases := []struct {
		in       string
		expected Amount
	}{
		{"", 0},
		{"0", 0},
		{"2", 2 * BTC},
		{"20", 20 * BTC},
		{"0.1", 100 * MilliBTC},
		{"0.100", 100 * MilliBTC},
		{"0.12345678", 12345678 * Satoshi},
		{"1.1", BTC + 100*MilliBTC},
		{"1.100", BTC + 100*MilliBTC},
		{"1.12345678", BTC + 12345678*Satoshi},
		{"20.1", 20*BTC + 100*MilliBTC},
		{"20.100", 20*BTC + 100*MilliBTC},
		{"20.12345678", 20*BTC + 12345678*Satoshi},
		{"0020.1", 20*BTC + 100*MilliBTC},
		{"020.100", 20*BTC + 100*MilliBTC},
		{"020.12345678", 20*BTC + 12345678*Satoshi},
		{"300.12345678", 300*BTC + 12345678*Satoshi},
		{"0020.0", 20 * BTC},
		{"0020.00000004", 20*BTC + 4*Satoshi},
		{"-20.1", -20*BTC + -100*MilliBTC},
		{"+20.1", 20*BTC + 100*MilliBTC},
		{"0020..0", 0},
		{"++0020.0", 0},
		{"--0020.0", 0},
		{"0+020.0", 0},
		{"0-020.0", 0},
		{"0.0 BTC", 0},
		{max, 92233720368 * BTC},
		{min, -92233720368 * BTC},
	}

	for _, c := range cases {
		result, _ := Parse(c.in)

		if result != c.expected {
			t.Errorf("'%s' parsed as %d sats (%s), %d (%s) expected ", c.in, result, result.String(), c.expected, c.expected.String())
		}
	}
}

func TestString(t *testing.T) {
	cases := []struct {
		in       string
		expected string
	}{
		{"0", "0 BTC"},
		{"2", "2 BTC"},
		{"2.01", "2.01 BTC"},
		{"0.002", "2 mBTC"},
		{"0.00023", "23000 sats"},
		{"-2", "-2 BTC"},
		{"-2.01", "-2.01 BTC"},
		{"-0.002", "-2 mBTC"},
		{"-0.00023", "-23000 sats"},
		{max, "92233720368 BTC"},
		{min, "-92233720368 BTC"},
	}

	for _, c := range cases {
		v, _ := Parse(c.in)
		result := v.String()

		if result != c.expected {
			t.Errorf("'%s'.String() -> '%s' expected '%s'", c.in, result, c.expected)
		}
	}
}

func TestSplit(t *testing.T) {
	cases := []struct {
		in    string
		pos   Amount
		part1 int
		part2 int
	}{
		{"0", 0, 0, 0},
		{"0.12345678", BTC, 0, 12345678},
		{"0.12345678", 100 * MilliBTC, 1, 2345678},
		{"0.12345678", 10 * MilliBTC, 12, 345678},
		{"0.12345678", MilliBTC, 123, 45678},
		{"0.12345678", 100 * MicroBTC, 1234, 5678},
		{"0.12345678", 10 * MicroBTC, 12345, 678},
		{"0.12345678", MicroBTC, 123456, 78},
		{"0.12345678", 10 * Satoshi, 1234567, 8},
		{"0.12345678", Satoshi, 12345678, 0},
		{"1.12345678", BTC, 1, 12345678},
		{"1.12345678", 100 * MilliBTC, 11, 2345678},
		{"1.12345678", 10 * MilliBTC, 112, 345678},
		{"1.12345678", MilliBTC, 1123, 45678},
		{"-1.12345678", 100 * MicroBTC, -11234, 5678},
		{"1.12345678", 10 * MicroBTC, 112345, 678},
		{"1.12345678", MicroBTC, 1123456, 78},
		{"1.12345678", 10 * Satoshi, 11234567, 8},
		{"1.12345678", Satoshi, 112345678, 0},
		{"10.12345678", BTC, 10, 12345678},
		{"10.12345678", 100 * MilliBTC, 101, 2345678},
		{"10.12345678", 10 * MilliBTC, 1012, 345678},
		{"10.12345678", MilliBTC, 10123, 45678},
		{"10.12345678", 100 * MicroBTC, 101234, 5678},
		{"10.12345678", 10 * MicroBTC, 1012345, 678},
		{"-10.12345678", MicroBTC, -10123456, 78},
		{"10.12345678", 10 * Satoshi, 101234567, 8},
		{"10.12345678", Satoshi, 1012345678, 0},
		{"10.12345678", 10 * BTC, 1, 12345678},
		{"10.12345678", 100 * BTC, 0, 1012345678},
		{"10.12345678", 1000 * BTC, 0, 1012345678},
		{"10.12345678", 10000 * BTC, 0, 1012345678},
		{max, BTC, 92233720368, 0},
		{min, BTC, -92233720368, 0},
	}

	for _, c := range cases {
		sats, _ := Parse(c.in)
		result1, result2 := sats.split(c.pos)

		if result1 != c.part1 || result2 != c.part2 {
			t.Errorf("%d (%s) splitted as %d:%d, %d:%d expected", sats, sats.String(), result1, result2, c.part1, c.part2)
		}
	}
}

func TestSplitString(t *testing.T) {
	cases := []struct {
		in    string
		pos   Amount
		part1 string
		part2 string
	}{
		{"0", BTC, "0", "0"},
		{"0.12345678", BTC, "0", "12345678"},
		{"0.12345678", 100 * MilliBTC, "1", "2345678"},
		{"0.12345678", 10 * MilliBTC, "12", "345678"},
		{"0.12345678", MilliBTC, "123", "45678"},
		{"0.12345678", 100 * MicroBTC, "1234", "5678"},
		{"0.12345678", 10 * MicroBTC, "12345", "678"},
		{"0.12345678", MicroBTC, "123456", "78"},
		{"0.12345678", 10 * Satoshi, "1234567", "8"},
		{"0.12345678", Satoshi, "12345678", "0"},
		{"1.12345678", BTC, "1", "12345678"},
		{"1.12345678", 100 * MilliBTC, "11", "2345678"},
		{"1.12345678", 10 * MilliBTC, "112", "345678"},
		{"1.12345678", MilliBTC, "1123", "45678"},
		{"-1.12345678", 100 * MicroBTC, "-11234", "5678"},
		{"1.12345678", 10 * MicroBTC, "112345", "678"},
		{"1.12345678", MicroBTC, "1123456", "78"},
		{"1.12345678", 10 * Satoshi, "11234567", "8"},
		{"1.12345678", Satoshi, "112345678", "0"},
		{"10.0012", BTC, "10", "0012"},
		{"10.12345678", BTC, "10", "12345678"},
		{"10.12345678", 100 * MilliBTC, "101", "2345678"},
		{"10.12345678", 10 * MilliBTC, "1012", "345678"},
		{"10.12345678", MilliBTC, "10123", "45678"},
		{"10.12345678", 100 * MicroBTC, "101234", "5678"},
		{"10.12345678", 10 * MicroBTC, "1012345", "678"},
		{"-10.12345678", MicroBTC, "-10123456", "78"},
		{"10.12345678", 10 * Satoshi, "101234567", "8"},
		{"10.12345678", Satoshi, "1012345678", "0"},
		{"10.12345678", 10 * BTC, "1", "012345678"},
		{"10.12345678", 100 * BTC, "0", "1012345678"},
		{"10.12345678", 1000 * BTC, "0", "01012345678"},
		{"10.12345678", 10000 * BTC, "0", "001012345678"},
		{max, BTC, "92233720368", "0"},
		{min, BTC, "-92233720368", "0"},
	}

	for _, c := range cases {
		sats, _ := Parse(c.in)
		result1, result2 := sats.SplitString(c.pos)

		if result1 != c.part1 || result2 != c.part2 {
			t.Errorf("%d (%s) splitted at %d as %s:%s, %s:%s expected", sats, sats.String(), c.pos, result1, result2, c.part1, c.part2)
		}
	}
}
