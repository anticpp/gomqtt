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

const (
	ConnectCodeOk                    = 0x00
	ConnectCodeRefuseProtocolVersion = 0x01
	ConnectCodeRefuseClientId        = 0x02
	ConnectCodeRefuseUnavailable     = 0x03
	ConnectCodeBadUsernameOrPassword = 0x04
	ConnectCodeNotAuthorized         = 0x05
)

type messageType interface {
	setHeader(h fixHeader)
	decodePayload(in []byte) (int, error) // Decode payload
	encode(out []byte) ([]byte, error)    // Encode header & payload
}

type fixHeader struct {
	Type   int
	Dup    int
	Qos    int
	Retain int
	Length int
}

func (h fixHeader) String() string {
	return fmt.Sprintf("FixHeader {Type: %v, Dup: %v, Qos: %v, Retain: %v, Length:%v}",
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
func (h *fixHeader) encode(out []byte) ([]byte, error) {

	if out == nil {
		out = make([]byte, 0)
	}

	var err error

	b0 := byte((h.Type&0x0F)<<4 |
		(h.Dup&0x01)<<3 |
		(h.Qos&0x03)<<1 |
		(h.Retain & 0x01))
	out = append(out, b0)

	out, err = encodeVariableInt32(h.Length, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Message with raw payload.
type messageRaw struct {
	header  fixHeader
	payload []byte
}

func newMessageRaw() *messageRaw {
	return new(messageRaw)
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
	clientId         string
	willTopic        string
	willMessage      []byte
	userName         string
	password         string
}

func newMessageConnect() *messageConnect {

	m := new(messageConnect)
	m.header.Type = MessageTypeConnect

	return m
}

func (m *messageConnect) setHeader(h fixHeader) {
	m.header = h
}
func (m *messageConnect) String() string {

	s := "Connect { "
	s += m.header.String()
	s += ", Variable Header "
	s += fmt.Sprintf("{ Name: %v, Level: %v, Flags: (UserName %v, Password %v, WillRetail %v, WillQos: %v, WillFlag: %v, CleanSession: %v ), KeepAlive: %v }",
		m.name,
		m.level,
		m.flagUserName,
		m.flagPassword,
		m.flagWillRetain,
		m.flagWillQos,
		m.flagWillFlag,
		m.flagCleanSession,
		m.keepAlive)
	s += fmt.Sprintf(", Payload { ClientId: %v, willTopic; %v, willMessage: len(%v), Username: %v, Password: %v}",
		m.clientId,
		m.willTopic,
		len(m.willMessage),
		m.userName,
		m.password)

	s += " }"
	return s
}
func (m *messageConnect) decodePayload(in []byte) (int, error) {
	var n int
	var err error
	var decodeLen = 0

	// Name
	m.name, n, err = decodeString(in)
	if err != nil {
		return 0, err
	}

	in = in[n:]
	decodeLen += n

	// Level, Flags
	if len(in) < 2 {
		return 0, ErrorDecodeMore{}
	}
	m.level = int(in[0])
	flags := int(in[1])
	m.flagUserName = ((flags >> 7) & 0x01)
	m.flagPassword = ((flags >> 6) & 0x01)
	m.flagWillRetain = ((flags >> 5) & 0x01)
	m.flagWillQos = ((flags >> 3) & 0x03)
	m.flagWillFlag = ((flags >> 2) & 0x01)
	m.flagCleanSession = ((flags >> 1) & 0x01)

	in = in[2:]
	decodeLen += 2

	// KeepAlive
	m.keepAlive, n, err = decodeInt16(in)
	if err != nil {
		return 0, nil
	}
	in = in[n:]
	decodeLen += n

	// Client Id
	m.clientId, n, err = decodeString(in)
	if err != nil {
		return 0, nil
	}
	in = in[n:]
	decodeLen += n

	// Will Topic, Will Message
	if m.flagWillFlag == 1 {

		m.willTopic, n, err = decodeString(in)
		if err != nil {
			return 0, nil
		}
		in = in[n:]
		decodeLen += n

		m.willMessage, n, err = decodeRawData(in)
		if err != nil {
			return 0, nil
		}
		in = in[n:]
		decodeLen += n
	}

	// Username
	if m.flagUserName == 1 {

		m.userName, n, err = decodeString(in)
		if err != nil {
			return 0, nil
		}
		in = in[n:]
		decodeLen += n
	}

	// Password
	if m.flagPassword == 1 {

		m.password, n, err = decodeString(in)
		if err != nil {
			return 0, nil
		}
		in = in[n:]
		decodeLen += n
	}

	return decodeLen, nil
}
func (m *messageConnect) encode(out []byte) ([]byte, error) {
	panic("Don't used.")
	return nil, nil
}

// ConnectAck
type messageConnectAck struct {
	header         fixHeader
	sessionPresent int
	returnCode     int
}

func newMessageConnectAck() *messageConnectAck {

	m := new(messageConnectAck)
	m.header.Type = MessageTypeConnectAck

	m.returnCode = ConnectCodeOk

	return m
}

func (m *messageConnectAck) decodePayload(in []byte) (int, error) {
	panic("Don't used.")
	return 0, nil
}
func (m *messageConnectAck) encode(out []byte) ([]byte, error) {

	if out == nil {
		out = make([]byte, 0, 1024)
	}

	var err error

	payload := make([]byte, 0)
	b0 := byte(m.sessionPresent & 0x01)
	b1 := byte(m.returnCode & 0xFF)
	payload = append(payload, b0)
	payload = append(payload, b1)

	m.header.Length = len(payload)
	out, err = m.header.encode(out)
	if err != nil {
		return nil, err
	}
	out = append(out, payload...)

	return out, nil
}
