package gomqtt

import (
//	"fmt"
)

const (
	maxVariableIntLength = 4
)

// Variable int will be encoded from 1 to 4 bytes.
//
// Return: Val, Expecting
//			Val         - Decode value. Only be meaningful when decoding is completed.
//						  Which is indecated by 'Expecting'.
//			Expecting   - Number of bytes expecting.
//				          Expecting<=len(buf), indicates decoding is completed.
//				          Expecting>len(buf), indecates decoding is incompleted, more bytes are expecting.
func decodeVariableInt4(buf []byte) (int, int) {

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
		expecting++
	}

	return val, expecting
}

func encodeVariableInt4(val int) []byte {

	digit := 0
	buf := make([]byte, maxVariableIntLength)

	i := 0
	for i = 0; i < len(buf) && val > 0; i++ {
		digit = val % 128
		val = val / 128

		if val > 0 { // If more byte.
			digit |= 0x80
		}

		buf[i] = byte(digit)
	}

	return buf[:i+1]
}
