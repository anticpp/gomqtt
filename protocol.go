package gomqtt

import (
//	"fmt"
)

const (
	maxVariableIntLength = 4
)

// Return: Val, N, Error
//			Val         - Decode value.
//			N		    - Byte length when success.
//			Error       - !nil             Success
//						  nil              Error
func decodeVariableInt4(buf []byte) (int, int, error) {

	val := 0
	length := 0

	i := 0
	mul := 1
	more := true
	for i = 0; i < len(buf) && more && length < maxVariableIntLength; i++ {
		digit := int(buf[i])
		length++

		val += ((digit & 0x7F) * mul)
		mul *= 128

		if digit&0x80 == 0 { //If more byte.
			more = false
		} else {
			more = true
		}
	}

	if more && length < maxVariableIntLength { // Most 'maxVariableIntLength' bytes expecting.
		return 0, 0, ErrorDecodeMore{}
	}

	return val, length, nil
}

// Maximum val 268435455.
// If the input value is greater than maximum, there will be no guarantee to the result.
// TODO:
//	Maybe we should return an error when encoding length beyonds 'maxVariableIntLength'.
//
// Return: buf, N, Error
//			buf         - Encode buffer.
//			N		    - Byte length when success.
//			Error       - !nil             Success
//						  nil              Error
func encodeVariableInt4(val int, out []byte) ([]byte, int, error) {

	l := len(out)

	digit := 0
	n := 0
	for {
		n++

		digit = val % 128
		val = val / 128

		if val > 0 && n < maxVariableIntLength { // Most 'maxVariableIntLength' should be encoded.
			digit |= 0x80
		}

		out = append(out, byte(digit))

		if val <= 0 || n >= maxVariableIntLength {
			break
		}
	}

	return out[:l+n], n, nil
}
