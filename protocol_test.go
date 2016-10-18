package gomqtt

import (
	"bytes"
	"testing"
)

func TestDecodeVariableInt32(t *testing.T) {
	for _, c := range []struct {
		in        []byte
		want      int
		length    int  // Decode length.
		completed bool // If completed bytes.
	}{
		// Empty
		{[]byte{}, 0, 0, false},

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

		v, n, err := decodeVariableInt32(c.in)
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

func TestEncodeVariableInt32(t *testing.T) {
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

		{268435455, []byte{0xFF, 0xFF, 0xFF, 0x7F}},
	} {
		var err error
		var buf []byte
		buf, err = encodeVariableInt32(c.in, nil)
		if err != nil {
			t.Errorf("Encode %v. Error occurs %v", c.in, err)
		} else if bytes.Compare(c.want, buf) != 0 {
			t.Errorf("Encode %v. (want)%v!=(encode)%v.", c.in, c.want, buf)
		}
	}
}

func TestDecodeInt16(t *testing.T) {
	for _, c := range []struct {
		in   []byte
		want int
	}{
		{[]byte{0x00, 0x00}, 0x0000},
		{[]byte{0x00, 0x01}, 0x0001},
		{[]byte{0x01, 0x01}, 0x0101},
		{[]byte{0xA0, 0x00}, 0xA000},
		{[]byte{0x0A, 0xB9}, 0x0AB9},
		{[]byte{0xFF, 0xFF}, 0xFFFF},
	} {
		v, n, err := decodeInt16(c.in)
		if err != nil {
			t.Errorf("In %v. Decode error %v\n", c.in, err)
		} else if n != 2 {
			t.Errorf("In %v, decode length unexpected. (want)%v!=(decode)%v.", c.in, 2, n)
		} else if c.want != v {
			t.Errorf("In %v, (want)%v!=(decode)%v.", c.in, c.want, v)
		}

	}
}

func TestEncodeInt16(t *testing.T) {
	for _, c := range []struct {
		in   int
		want []byte
	}{
		{0x0000, []byte{0x00, 0x00}},
		{0x0001, []byte{0x00, 0x01}},
		{0x0101, []byte{0x01, 0x01}},
		{0xA000, []byte{0xA0, 0x00}},
		{0x0AB9, []byte{0x0A, 0xB9}},
		{0xFFFF, []byte{0xFF, 0xFF}},
	} {
		var err error
		var out []byte

		out, err = encodeInt16(c.in, nil)
		if err != nil {
			t.Errorf("Encode 0x%X. Error occurs %v", c.in, err)
		} else if bytes.Compare(c.want, out) != 0 {
			t.Errorf("Encode 0x%X. (want)%v!=(encode)%v.", c.in, c.want, out)
		}
	}
}

func TestDecodeString(t *testing.T) {
	for _, c := range []struct {
		in   []byte
		want string
	}{
		{[]byte{0x00, 0x00}, ""},
		{[]byte{0x00, 0x05, 'h', 'e', 'l', 'l', 'o', 'x', 'x'}, "hello"},
		{[]byte{0x01, 0x01, '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
			'a'},
			"0123456789012345" +
				"0123456789012345" +
				"0123456789012345" +
				"0123456789012345" +
				"0123456789012345" +
				"0123456789012345" +
				"0123456789012345" +
				"0123456789012345" +
				"0123456789012345" +
				"0123456789012345" +
				"0123456789012345" +
				"0123456789012345" +
				"0123456789012345" +
				"0123456789012345" +
				"0123456789012345" +
				"0123456789012345" +
				"a"},
	} {
		v, n, err := decodeString(c.in)
		if err != nil {
			t.Errorf("In %v. Decode error %v\n", c.in, err)
		} else if n != 2+len(c.want) {
			t.Errorf("In %v, decode length unexpected. (want)%v!=(decode)%v.", c.in, 2+len(c.want), n)
		} else if v != c.want {
			t.Errorf("In %v, (want)%v!=(decode)%v.", c.in, c.want, v)
		}
	}
}

func TestEncodeString(t *testing.T) {
	for _, c := range []struct {
		in   string
		want []byte
	}{
		{"", []byte{0x00, 0x00}},
		{"hello", []byte{0x00, 0x05, 'h', 'e', 'l', 'l', 'o'}},
		{"0123456789012345" +
			"0123456789012345" +
			"0123456789012345" +
			"0123456789012345" +
			"0123456789012345" +
			"0123456789012345" +
			"0123456789012345" +
			"0123456789012345" +
			"0123456789012345" +
			"0123456789012345" +
			"0123456789012345" +
			"0123456789012345" +
			"0123456789012345" +
			"0123456789012345" +
			"0123456789012345" +
			"0123456789012345" +
			"a",
			[]byte{0x01, 0x01, '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5',
				'a'}},
	} {
		var err error
		var out []byte
		out, err = encodeString(c.in, nil)
		if err != nil {
			t.Errorf("Encode '%v'. Error occurs %v", c.in, err)
		} else if bytes.Compare(c.want, out) != 0 {
			t.Errorf("Encode '%v'. (want)%v!=(encode)%v.", c.in, c.want, out)
		}

	}
}
func TestDecodeRawData(t *testing.T) {

	for _, c := range []struct {
		in   []byte
		want []byte
	}{
		{[]byte{0x00, 0x00}, []byte{}},
		{[]byte{0x00, 0x05, 'h', 'e', 'l', 'l', 'o'}, []byte{'h', 'e', 'l', 'l', 'o'}},
		{[]byte{0x00, 0x05, 'h', 'e', 'l', 'l', 'o', 'x', 'x'}, []byte{'h', 'e', 'l', 'l', 'o'}},
	} {
		out, _, err := decodeRawData(c.in)
		if err != nil {
			t.Fatalf("Decode fail. %v", c.in)
		}

		if bytes.Compare(c.want, out) != 0 {
			t.Fatalf("Decode fail. (want)%v!=(decode)%v", c.want, out)
		}

	}
}
func TestEncodeRawData(t *testing.T) {

	for _, c := range []struct {
		in   []byte
		want []byte
	}{
		{[]byte{}, []byte{0x00, 0x00}},
		{[]byte{'h', 'e', 'l', 'l', 'o'}, []byte{0x00, 0x05, 'h', 'e', 'l', 'l', 'o'}},
	} {
		var err error
		var out []byte

		out, err = encodeRawData(c.in, nil)
		if err != nil {
			t.Fatalf("Encode fail. %v", c.in)
		}

		if bytes.Compare(c.want, out) != 0 {
			t.Fatalf("Encode fail. (want)%v!=(decode)%v", c.want, out)
		}

	}
}

func TestFixHeaderEncodeDecode(t *testing.T) {
	for _, c := range []fixHeader{
		fixHeader{Type: 0, Dup: 0, Qos: 0, Retain: 0},
		fixHeader{Type: 4, Dup: 0, Qos: 0, Retain: 0},
		fixHeader{Type: 0, Dup: 1, Qos: 0, Retain: 0},
		fixHeader{Type: 0, Dup: 0, Qos: 1, Retain: 0},
		fixHeader{Type: 0, Dup: 0, Qos: 0, Retain: 1},
		fixHeader{Type: 4, Dup: 1, Qos: 2, Retain: 1},
	} {
		var err error
		var out []byte

		out, err = c.encode(nil)
		if err != nil {
			t.Fatalf("Encode fail. %v", c)
		}

		n := len(out)

		c1 := fixHeader{}
		var n1 int
		n1, err = c1.decode(out)
		if err != nil {
			t.Fatalf("Decode fail. %v", c)
		}

		if n != n1 {
			t.Fatalf("Encode size %v!=Decode size %v. %v", n, n1, c)
		}

		if c.Type != c1.Type {
			t.Fatalf("Encode type %v!=Decode type %v. %v", c.Type, c1.Type, c)
		}

		if c.Dup != c1.Dup {
			t.Fatalf("Encode Dup %v!=Decode Dup %v. %v", c.Dup, c1.Dup, c)
		}

		if c.Qos != c1.Qos {
			t.Fatalf("Encode Qos %v!=Decode Qos %v. %v", c.Qos, c1.Qos, c)
		}

		if c.Retain != c1.Retain {
			t.Fatalf("Encode Retain %v!=Decode Retain %v. %v", c.Retain, c1.Retain, c)
		}

	}
}
