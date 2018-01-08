package yaml

import (
	"fmt"
	"strconv"
)

// Octal is an int64 that represents an octal number.
type Octal int64

// String converts the given octal number to a string
// with a leading "0".
func (o Octal) String() string {
	return fmt.Sprintf("%#o", int64(o))
}

// Int64 returns the Octal number as an int64
func (number Octal) Int64() int64 {
	return int64(number)
}

// New takes an octal number as a string and returns
// an int64 as the Octal type, or 0. Use New2() for better
// error handling.
func New(octal string) Octal {
	number, err := strconv.ParseInt(octal, 8, 64)
	if err != nil {
		return 0
	}
	return Octal(number)
}

// New2 takes an octal number as a string and returns
// an int64 as the Octal type, along with an error value.
func New2(octal string) (Octal, error) {
	number, err := strconv.ParseInt(octal, 8, 64)
	if err != nil {
		return 0, err
	}
	return Octal(number), nil
}
