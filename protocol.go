package gomqtt

// Return: val, n
//			val - Value of length.
//			n   - Number of bytes decode from buf.
func decodeRemainingLength(buf []byte) (int, int) {

	val := 0
	mul := 1

	i := 0
	for i = 0; i < len(buf); i++ {
		digit := int(buf[i])

		val += ((digit & 0x7F) * mul)
		mul *= 128

		if digit&0x80 == 0 { //If more byte.
			break
		}

		i++ // Next
	}
	return val, i + 1
}

func encodeRemainingLength(val int) []byte {

	digit := 0
	buf := make([]byte, 4)

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
