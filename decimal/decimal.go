package decimal

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

type Decimal int64

const Precision = 4
const Divisor = 10000
const Max int64 = math.MaxInt64
const Min int64 = math.MinInt64

var Overflow = errors.New("Decimal overflow")
var Invalid = errors.New("Invalid literal")

func Parse(s string) (Decimal, error) {

	var (
		ret        uint64
		negative   uint64
		fractional bool
		precision  int = Precision
	)

	s = strings.TrimLeft(s, " \t\r\n")
	slen := len(s)
	if slen == 0 {
		// empty string
	} else if s[0] == '-' {
		negative = math.MaxUint64
		s = s[1:]
		slen--
	} else if s[0] == '+' {
		s = s[1:]
		slen--
	}

	for i := 0; i < slen && (!fractional || precision > 0); i++ {
		if !fractional && s[i] == '.' {
			fractional = true
		} else if s[i] < '0' || s[i] > '9' {
			break
		} else if ret > math.MaxUint64/10 {
			return Decimal(0), Overflow
		} else {
			ret = ret * 10
			digit := uint64(s[i] - '0')
			if ret > math.MaxUint64-digit {
				return Decimal(0), Overflow
			}

			ret += digit
			if fractional {
				precision--
			}
		}
	}

	if ret == 0 && precision == Precision {
		return Decimal(0), Invalid
	}

	for ; precision > 0; precision-- {
		if ret > math.MaxUint64/10 {
			return Decimal(0), Overflow
		}
		ret = ret * 10
	}

	if negative == 0 {
		if ret > math.MaxInt64 {
			return Decimal(math.MaxInt64), Overflow
		}
	} else {
		if ret > math.MaxInt64+1 {
			return Decimal(math.MinInt64), Overflow
		}
	}

	return Decimal((ret ^ negative) - negative), nil
}

func (d Decimal) String() string {
	var integ, fract int64

	value := int64(d)
	integ = value / Divisor
	if d >= 0 {
		fract = value % Divisor
	} else {
		fract = -(value % Divisor)
	}
	return fmt.Sprintf("%d.%.04d", integ, fract)
}
