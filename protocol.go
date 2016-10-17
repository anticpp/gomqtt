package gomqtt

import (
//"fmt"
)

const (
	maxVariableIntLength = 4
)

// Return: Val, N, Error
//			Val         - Decode value.
//			N		    - Decode length.
//			Error       - !nil             Success
//						  nil              Error
func decodeVariableInt32(in []byte) (int, int, error) {

	val := 0
	length := 0

	i := 0
	mul := 1
	more := true
	for i = 0; i < len(in) && more && length < maxVariableIntLength; i++ {
		digit := int(in[i])
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
// If the input value is greater than maximum, the result will be no guarantee.
//
// Return: buf, N, Error
//			buf         - Encode buffer.
//			N		    - Encode length.
//			Error       - !nil             Success
//						  nil              Error
func encodeVariableInt32(val int, out []byte) ([]byte, int, error) {

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

	return out, n, nil
}

// Return: Val, N, Error
//			Val         - Decode value.
//			N		    - Decode length.
//			Error       - !nil             Success
//						  nil              Error
func decodeInt16(in []byte) (int, int, error) {

	if len(in) < 2 {
		return 0, 0, ErrorDecodeMore{}
	}

	b0 := (int(in[0]) << 8)
	b1 := int(in[1])
	return b0 + b1, 2, nil
}

// Return: buf, N, Error
//			buf         - Encode buffer.
//			N		    - Encode length.
//			Error       - !nil             Success
//						  nil              Error
func encodeInt16(val int, out []byte) ([]byte, int, error) {
	b0 := (val >> 8) & 0x00FF
	b1 := val & 0x00FF
	out = append(out, byte(b0))
	out = append(out, byte(b1))
	return out, 2, nil
}

// Return: S, N, Error
//			S         - Decode string.
//			N		    - Decode length.
//			Error       - !nil             Success
//						  nil              Error
func decodeString(in []byte) (string, int, error) {
	l, n, err := decodeInt16(in)
	if err != nil {
		return "", 0, err
	}
	if l > len(in)-n {
		l = len(in) - n
	}
	out := string(in[n : l+n])

	return out, l + n, nil
}

// Return: buf, N, Error
//			buf         - Encode buffer.
//			N		    - Encode length.
//			Error       - !nil             Success
//						  nil              Error
func encodeString(s string, out []byte) ([]byte, int, error) {
	var n int
	var err error

	l := len(s)
	out, n, err = encodeInt16(l, out)
	if err != nil {
		return nil, 0, err
	}

	out = append(out, []byte(s)...)
	return out, l + n, nil
}

func decodeRawData(in []byte) ([]byte, int, error) {
	out := make([]byte, 0)
	l, n, err := decodeInt16(in)
	if err != nil {
		return nil, 0, err
	}
	if l > len(in)-n {
		l = len(in) - n
	}
	out = append(out, in[n:n+l]...)
	return out, l + n, nil
}

func encodeRawData(in []byte, out []byte) ([]byte, int, error) {
	var n int
	var err error

	l := len(in)
	out, n, err = encodeInt16(l, out)
	if err != nil {
		return nil, 0, err
	}

	out = append(out, in...)
	return out, l + n, nil
}
