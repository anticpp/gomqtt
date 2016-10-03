package gomqtt

import (
	"bytes"
	"testing"
)

func TestDecodeVariableInt4(t *testing.T) {
	for _, c := range []struct {
		in        []byte
		want      int
		length    int  // Decode length.
		completed bool // If completed bytes.
	}{
		// Normal bytes.
		{[]byte{0x00}, 0, 1, true},
		{[]byte{0x01}, 1, 1, true},
		{[]byte{0x80, 0x01}, 128, 2, true},
		{[]byte{0x80, 0x80, 0x01}, 16384, 3, true},
		{[]byte{0x80, 0x80, 0x80, 0x01}, 2097152, 4, true},

		// More arbitary bytes at tail.
		{[]byte{0x00, 0x01}, 0, 1, true},
		{[]byte{0x01, 0x01}, 1, 1, true},
		{[]byte{0x80, 0x01, 0x01}, 128, 2, true},
		{[]byte{0x80, 0x80, 0x01, 0x01}, 16384, 3, true},
		{[]byte{0x80, 0x80, 0x80, 0x01, 0x01}, 2097152, 4, true},

		// Most 4 bytes.
		// The fifth byte should be ignored, although 'More-Byte' indecates by the 4'st byte.
		{[]byte{0x80, 0x80, 0x80, 0x81, 0x01}, 2097152, 4, true},

		// Incomplete bytes.
		{[]byte{0x80}, -1, -1, false},
		{[]byte{0x80, 0x80}, -1, -1, false},
		{[]byte{0x80, 0x80, 0x80}, -1, -1, false},

		// Randam.
		{[]byte{0x3A}, 58, 1, true},
		{[]byte{0x8F, 0x23}, 4495, 2, true},
		{[]byte{0x93, 0xA5, 0x78}, 1970835, 3, true},
		{[]byte{0xA6, 0xBF, 0x89, 0x04}, 8544166, 4, true},
	} {

		v, n, err := decodeVariableInt4(c.in)
		if c.completed == false {
			if err == nil {
				t.Errorf("In %v. Incompleted but decode success.\n", c.in)
			} else if _, ok := err.(ErrorDecodeMore); !ok {
				t.Errorf("In %v. Incompleted but decode with error %v. Should be ErrorDecodeMore.\n", err)
			}
		} else {
			if err != nil {
				t.Errorf("In %v. Decode error %v\n", c.in, err)
			} else if c.length != n {
				t.Errorf("In %v, decode length unexpected. (want)%v!=(decode)%v.", c.in, c.length, n)
			} else if c.want != v {
				t.Errorf("In %v, (want)%v!=(decode)%v.", c.in, c.want, v)
			}
		}
	}
}

func TestEncodeVariableInt4(t *testing.T) {
	for _, c := range []struct {
		in   int
		want []byte
	}{
		{0, []byte{0x00}},
		{1, []byte{0x01}},
		{128, []byte{0x80, 0x01}},
		{16384, []byte{0x80, 0x80, 0x01}},
		{2097152, []byte{0x80, 0x80, 0x80, 0x01}},

		// Randam.
		{58, []byte{0x3A}},
		{4495, []byte{0x8F, 0x23}},
		{1970835, []byte{0x93, 0xA5, 0x78}},
		{8544166, []byte{0xA6, 0xBF, 0x89, 0x04}},

		// Beyond maximum. Max 4 bytes should be encoded.
		// 268435455 is maximum.
		{268435455 + 1000, []byte{0xE7, 0x87, 0x80, 0x00}},
	} {
		var err error
		buf := make([]byte, 0, 32)
		buf, _, err = encodeVariableInt4(c.in, buf)
		if err != nil {
			t.Errorf("Encode %v. Error occurs %v", c.in, err)
		} else if bytes.Compare(c.want, buf) != 0 {
			t.Errorf("Encode %v. (want)%v!=(encode)%v.", c.in, c.want, buf)
		}
	}
}
