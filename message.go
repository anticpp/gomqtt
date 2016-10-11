package gomqtt

import (
	"fmt"
)

const (
	QoS0 = 0
	QoS1 = 1
	QoS2 = 2
)

var QosString = []string{
	"Qos0",
	"Qos1",
	"Qos2",
}

const (
	MessageTypeConnect    = 1
	MessageTypeConnectAck = 2
	MessageTypePub        = 3
	MessageTypePubAck     = 4
	MessageTypePubRec     = 5
	MessageTypePubRel     = 6
	MessageTypePubComp    = 7
	MessageTypeSub        = 8
	MessageTypeSubAck     = 9
	MessageTypeUnsub      = 10
	MessageTypeUnsubAck   = 11
	MessageTypePingReq    = 12
	MessageTypePingResp   = 13
	MessageTypeDisconnect = 14
)

var MessageTypeString = []string{
	"Reserved",
	"Connect",
	"ConnectAck",
	"Publish",
	"PublishAck",
	"PublishRecord",
	"PublishRelease",
	"PublishComplete",
	"Subscribe",
	"SubscribeAck",
	"UnSubscribe",
	"UnSubscribeAck",
	"PingReq",
	"PingResp",
	"Disconnect",
	"Reserved",
}

type messageType interface {
	setHeader(h fixHeader)
	decode(in []byte) (int, error)
	encode(out []byte) ([]byte, int, error)
	String() string
}

type fixHeader struct {
	Type   int
	Dup    int
	Qos    int
	Retain int
	Length int
}

func (h fixHeader) String() string {
	return fmt.Sprintf("{Type: %v, Dup: %v, Qos: %v, Retain: %v, Length:%v}",
		MessageTypeString[h.Type],
		h.Dup,
		QosString[h.Qos],
		h.Retain,
		h.Length)
}

func (h fixHeader) TypeName() string {
	return MessageTypeString[h.Type]
}

// Return:
//		 	Bytes - Byte length when success.
//		    Error - nil  Success,
//					!nil error.
func (h *fixHeader) decode(in []byte) (int, error) {

	var err error
	var n int

	if len(in) == 0 {
		return 0, ErrorDecodeMore{}
	}

	b0 := in[0]
	h.Type = int(0x0F & (b0 >> 4))
	h.Dup = int(0x01 & (b0 >> 3))
	h.Qos = int(0x03 & (b0 >> 1))
	h.Retain = int(0x01 & b0)

	h.Length, n, err = decodeVariableInt32(in[1:])
	if err != nil {
		return 0, err
	}
	return n + 1, nil
}
func (h *fixHeader) encode(out []byte) ([]byte, int, error) {
	var n int
	var err error

	b0 := byte(0)
	b0 = byte((h.Type&0x0F)<<4 |
		(h.Dup&0x01)<<3 |
		(h.Qos&0x03)<<1 |
		(h.Retain & 0x01))

	out = append(out, b0)

	out, n, err = encodeVariableInt32(h.Length, out)
	if err != nil {
		return nil, 0, err
	}
	return out, n + 1, nil
}

// Message with raw payload.
type messageRaw struct {
	header  fixHeader
	payload []byte
}

// Connect
type messageConnect struct {
	header           fixHeader
	name             string
	level            int
	flagUserName     int
	flagPassword     int
	flagWillRetain   int
	flagWillQos      int
	flagWillFlag     int
	flagCleanSession int
	keepAlive        int
}

func (m *messageConnect) setHeader(h fixHeader) {
	m.header = h
}
func (m *messageConnect) String() string {

	s := "FixHeader: "
	s += m.header.String()
	s += ", Variable Header: "
	s += fmt.Sprintf("{ Name: %v, Level: %v, Flags: (UserName %v, Password %v, WillRetail %v, WillQos: %v, WillFlag: %v, CleanSession: %v ), KeepAlive: %v }",
		m.name, m.level, m.flagUserName, m.flagPassword, m.flagWillRetain, m.flagWillQos, m.flagWillFlag, m.flagCleanSession, m.keepAlive)
	return s
}
func (m *messageConnect) decode(in []byte) (int, error) {
	var n int
	var err error
	var decodeLen = 0

	m.name, n, err = decodeString(in)
	if err != nil {
		return 0, err
	}
	in = in[n:]
	decodeLen += n

	if len(in) < 2 {
		return 0, ErrorDecodeMore{}
	}
	m.level = int(in[0])
	flags := int(in[1])

	in = in[2:]
	decodeLen += 2

	m.flagUserName = ((flags >> 7) & 0x01)
	m.flagPassword = ((flags >> 6) & 0x01)
	m.flagWillRetain = ((flags >> 5) & 0x01)
	m.flagWillQos = ((flags >> 3) & 0x03)
	m.flagWillFlag = ((flags >> 2) & 0x01)
	m.flagCleanSession = ((flags >> 1) & 0x01)

	m.keepAlive, n, err = decodeInt16(in)
	if err != nil {
		return 0, nil
	}
	decodeLen += n

	return decodeLen, nil
}
func (m *messageConnect) encode(out []byte) ([]byte, int, error) {
	panic("Don't used.")
	return nil, 0, nil
}

// ConnectAck
type messageConnectAck struct {
	header         fixHeader
	sessionPresent int
	returnCode     int
}

func (m *messageConnectAck) decoe(in []byte) (int, error) {
	panic("Don't used.")
	return 0, nil
}
func (m *messageConnectAck) encode(out []byte) ([]byte, int, error) {
	b0 := byte(m.sessionPresent & 0x01)
	b1 := byte(m.returnCode & 0xFF)
	out = append(out, b0)
	out = append(out, b1)
	return out, 2, nil
}
