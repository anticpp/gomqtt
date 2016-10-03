package gomqtt

import (
	"testing"
)

func TestDecodeVariableInt(t *testing.T) {
	for _, c := range []struct {
		in        []byte
		want      int
		expecting int
	}{
		// Normal bytes.
		{[]byte{0x01}, 1, 1},
		{[]byte{0x80, 0x01}, 128, 2},
		{[]byte{0x80, 0x80, 0x01}, 16384, 3},
		{[]byte{0x80, 0x80, 0x80, 0x01}, 2097152, 4},

		// More arbitary bytes at tail.
		{[]byte{0x01, 0x01}, 1, 1},
		{[]byte{0x80, 0x01, 0x01}, 128, 2},
		{[]byte{0x80, 0x80, 0x01, 0x01}, 16384, 3},
		{[]byte{0x80, 0x80, 0x80, 0x01, 0x01}, 2097152, 4},

		// Most 4 bytes.
		// The fifth byte should be ignored, although 'More-Byte' indecates by the 4'st byte.
		{[]byte{0x80, 0x80, 0x80, 0x81, 0x01}, 2097152, 4},

		// Incomplete bytes. Should expecting more.
		{[]byte{0x80}, -1, 2},
		{[]byte{0x80, 0x80}, -1, 3},
		{[]byte{0x80, 0x80, 0x80}, -1, 4},

		// Randam.
		{[]byte{0x3A}, 58, 1},
		{[]byte{0x8F, 0x23}, 4495, 2},
		{[]byte{0x93, 0xA5, 0x78}, 1970835, 3},
		{[]byte{0xA6, 0xBF, 0x89, 0x04}, 8544166, 4},
	} {
		v, expecting := decodeVariableInt4(c.in)
		if expecting != c.expecting {
			t.Errorf("Expecting of %v, (want)%v!=(decode)%v.", c.in, c.expecting, expecting)
		}
		if expecting <= len(c.in) && v != c.want {
			t.Errorf("Digit of %v, (want)%v!=(decode)%v.", c.in, c.want, v)
		}
	}
}

func TestEncodeRemainingLength(t *testing.T) {

}

func TestEncodeAndDecodeRemainingLength(t *testing.T) {

}
