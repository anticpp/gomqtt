package gomqtt

import (
	"bytes"
	"testing"
)

func TestDecodeVariableInt4(t *testing.T) {
	for _, c := range []struct {
		in        []byte
		want      int
		completed bool
	}{
		// Normal bytes.
		{[]byte{0x00}, 0, true},
		{[]byte{0x01}, 1, true},
		{[]byte{0x80, 0x01}, 128, true},
		{[]byte{0x80, 0x80, 0x01}, 16384, true},
		{[]byte{0x80, 0x80, 0x80, 0x01}, 2097152, true},

		// More arbitary bytes at tail.
		{[]byte{0x00, 0x01}, 0, true},
		{[]byte{0x01, 0x01}, 1, true},
		{[]byte{0x80, 0x01, 0x01}, 128, true},
		{[]byte{0x80, 0x80, 0x01, 0x01}, 16384, true},
		{[]byte{0x80, 0x80, 0x80, 0x01, 0x01}, 2097152, true},

		// Most 4 bytes.
		// The fifth byte should be ignored, although 'More-Byte' indecates by the 4'st byte.
		{[]byte{0x80, 0x80, 0x80, 0x81, 0x01}, 2097152, true},

		// Incomplete bytes.
		{[]byte{0x80}, -1, false},
		{[]byte{0x80, 0x80}, -1, false},
		{[]byte{0x80, 0x80, 0x80}, -1, false},

		// Randam.
		{[]byte{0x3A}, 58, true},
		{[]byte{0x8F, 0x23}, 4495, true},
		{[]byte{0x93, 0xA5, 0x78}, 1970835, true},
		{[]byte{0xA6, 0xBF, 0x89, 0x04}, 8544166, true},
	} {

		v, err := decodeVariableInt4(c.in)
		if c.completed == false {
			if err == nil {
				t.Errorf("In %v. Incompleted but decode success.\n", c.in)
			} else if _, ok := err.(ErrorDecodeMore); !ok {
				t.Errorf("In %v. Incompleted but decode with error %v. Should be ErrorDecodeMore.\n", err)
			}
		} else {
			if err != nil {
				t.Errorf("In %v. Decode error %v\n", c.in, err)
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
	} {

		buf := make([]byte, 0, 32)
		buf, _ = encodeVariableInt4(c.in, buf)
		if bytes.Compare(c.want, buf) != 0 {
			t.Errorf("Encode %v. (want)%v!=(encode)%v.", c.in, c.want, buf)
		}
	}
}
