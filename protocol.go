package gomqtt

import (
//	"fmt"
)

const (
	maxVariableIntLength = 4
)

// Variable int will be encoded from 1 to 4 bytes.
//
// Return: Val, Error
//			Val         - Decode value.
//			Error       - ErrorDecodeMore, expecting more.
func decodeVariableInt4(buf []byte) (int, error) {

	val := 0
	expecting := 0

	i := 0
	mul := 1
	more := true
	for i = 0; i < len(buf) && more && expecting < maxVariableIntLength; i++ {
		digit := int(buf[i])
		expecting++

		val += ((digit & 0x7F) * mul)
		mul *= 128

		if digit&0x80 == 0 { //If more byte.
			more = false
		} else {
			more = true
		}
	}

	if more && expecting < maxVariableIntLength { // Most 4 bytes expecting.
		return 0, ErrorDecodeMore{}
	}

	return val, nil
}

func encodeVariableInt4(val int, out []byte) ([]byte, int) {

	l := len(out)

	digit := 0
	n := 0
	for {
		digit = val % 128
		val = val / 128

		if val > 0 { // If more byte.
			digit |= 0x80
		}

		out = append(out, byte(digit))
		n++

		if val <= 0 || n >= maxVariableIntLength {
			break
		}
	}

	return out[:l+n], n
}
