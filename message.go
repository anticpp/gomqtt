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
	getHeader() fixHeader

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
	return fmt.Sprintf("FixHeader { Type: %v, Dup: %v, Qos: %v, Retain: %v, Length:%v }",
		MessageTypeString[h.Type],
		h.Dup,
		QosString[h.Qos],
		h.Retain,
		h.Length)
}

func (h fixHeader) getType() int {
	return h.Type
}
func (h fixHeader) getQos() int {
	return h.Qos
}
func (h fixHeader) getDup() int {
	return h.Dup
}
func (h fixHeader) getRetain() int {
	return h.Retain
}
func (h fixHeader) typeName() string {
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
	header fixHeader

	// Variable Header
	name             string
	level            int
	flagUserName     int
	flagPassword     int
	flagWillRetain   int
	flagWillQos      int
	flagWillFlag     int
	flagCleanSession int
	keepAlive        int

	// Payload
	clientId    string
	willTopic   string
	willMessage []byte
	userName    string
	password    string
}

func newMessageConnect() *messageConnect {

	m := new(messageConnect)
	m.header.Type = MessageTypeConnect

	return m
}

func (m *messageConnect) setHeader(h fixHeader) {
	m.header = h
}
func (m *messageConnect) getHeader() fixHeader {
	return m.header
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
	header fixHeader

	// Variable Header
	sessionPresent int
	returnCode     int

	// Payload
	// None
}

func newMessageConnectAck() *messageConnectAck {

	m := new(messageConnectAck)
	m.header.Type = MessageTypeConnectAck

	m.returnCode = ConnectCodeOk

	return m
}

func (m *messageConnectAck) setHeader(h fixHeader) {
	m.header = h
}

func (m *messageConnectAck) getHeader() fixHeader {
	return m.header
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

// Publish
type messagePub struct {
	header fixHeader

	// Variable Header
	topicName string
	packetId  int

	// Payload
	data []byte
}

func newMessagePub() *messagePub {
	m := new(messagePub)
	m.header.Type = MessageTypePub

	return m
}

func (m *messagePub) String() string {
	s := "Publish { "
	s += m.header.String()
	s += ", Variable Header: { "
	s += fmt.Sprintf(" topic_name: %v, packet_id: %v", m.topicName, m.packetId)
	s += " }"
	s += fmt.Sprintf(", Payload(%v): [...]", len(m.data))
	s += " }"
	return s
}

func (m *messagePub) setHeader(h fixHeader) {
	m.header = h
}
func (m *messagePub) getHeader() fixHeader {
	return m.header
}

func (m *messagePub) decodePayload(in []byte) (int, error) {
	var n int
	var err error
	var decodeLen = 0

	// Topic Name
	m.topicName, n, err = decodeString(in)
	if err != nil {
		return 0, err
	}
	in = in[n:]
	decodeLen += n

	// Packet Identifier
	if m.header.Qos == QoS1 || m.header.Qos == QoS2 {
		m.packetId, n, err = decodeInt16(in)
		if err != nil {
			return 0, err
		}
		in = in[n:]
		decodeLen += n
	}

	// Data
	// Remaining data is payload
	m.data = in
	decodeLen += len(m.data)

	return decodeLen, nil
}

func (m *messagePub) encode(out []byte) ([]byte, error) {
	panic("Don't used.")
	return nil, nil
}

// PublishAck
type messagePubAck struct {
	header fixHeader

	// Variable Header
	packetId int

	// Payload
	// None
}

func newMessagePubAck() *messagePubAck {
	m := new(messagePubAck)
	m.header.Type = MessageTypePubAck

	return m
}

func (m *messagePubAck) String() string {

	s := "PublishAck { "
	s += m.header.String()
	s += fmt.Sprintf(", Variable Header { PacketId: %v }", m.packetId)
	s += ", Payload: None"
	s += " }"
	return s
}

func (m *messagePubAck) setHeader(h fixHeader) {
	m.header = h
}
func (m *messagePubAck) getHeader() fixHeader {
	return m.header
}

func (m *messagePubAck) decodePayload(in []byte) (int, error) {
	panic("Don't used.")
	return 0, nil
}

func (m *messagePubAck) encode(out []byte) ([]byte, error) {

	if out == nil {
		out = make([]byte, 0, 1024)
	}

	var err error

	payload := make([]byte, 0)
	payload, err = encodeInt16(m.packetId, payload)
	if err != nil {
		return nil, err
	}

	m.header.Length = len(payload)
	out, err = m.header.encode(out)
	if err != nil {
		return nil, err
	}
	out = append(out, payload...)

	return out, nil
}

// Subscribe
type topicFilter struct {
	topic string
	qos   int
}

type messageSub struct {
	header fixHeader

	// Variable Header
	packetId int

	// Payload
	filters []topicFilter
}

func newMessageSub() *messageSub {
	m := new(messageSub)
	m.header.Type = MessageTypeSub

	m.filters = []topicFilter{}

	return m
}

func (m *messageSub) String() string {
	s := "Subscribe { "
	s += m.header.String()
	s += fmt.Sprintf(", Variable Header { PacketId: %v }", m.packetId)
	s += ", Payload { Filters: ["
	for _, f := range m.filters {
		s += fmt.Sprintf(" { topic: %v, qos: %v }, ", f.topic, f.qos)
	}
	s += "] }"
	s += " }"
	return s
}
func (m *messageSub) setHeader(h fixHeader) {
	m.header = h
}
func (m *messageSub) getHeader() fixHeader {
	return m.header
}
func (m *messageSub) decodePayload(in []byte) (int, error) {

	var n int
	var err error
	var decodeLen = 0

	// Packet Identifier
	m.packetId, n, err = decodeInt16(in)
	if err != nil {
		return 0, err
	}
	in = in[n:]
	decodeLen += n

	// Filters
	for {
		var filter topicFilter

		if len(in) == 0 {
			break
		}

		filter.topic, n, err = decodeString(in)
		if err != nil {
			return 0, err
		}
		in = in[n:]
		decodeLen += n

		filter.qos = int(in[0])
		in = in[1:]
		decodeLen += 1

		m.filters = append(m.filters, filter)
	}

	return decodeLen, nil
}
func (m *messageSub) encode(out []byte) ([]byte, error) {
	panic("Don't use")
	return nil, nil
}

// SubscribeAck
type messageSubAck struct {
	header fixHeader

	// Variable Header
	packetId int

	// Payload
	returnCodes []int
}

func newMessageSubAck() *messageSubAck {
	m := new(messageSubAck)
	m.header.Type = MessageTypeSubAck

	m.returnCodes = []int{}

	return m
}
func (m *messageSubAck) String() string {
	s := "SubscribeAck { "
	s += m.header.String()
	s += fmt.Sprintf(", Variable Header { PacketId: %v }", m.packetId)
	s += ", Payload: { ReturnCodes: ["
	for _, code := range m.returnCodes {
		s += fmt.Sprintf("%v, ", code)
	}
	s += "] }"
	s += "}"
	return s
}

func (m *messageSubAck) setHeader(h fixHeader) {
	m.header = h
}
func (m *messageSubAck) getHeader() fixHeader {
	return m.header
}
func (m *messageSubAck) decodePayload(in []byte) (int, error) {
	panic("Don't use")
	return 0, nil
}
func (m *messageSubAck) encode(out []byte) ([]byte, error) {

	if out == nil {
		out = make([]byte, 0, 1024)
	}

	var err error

	payload := make([]byte, 0)
	payload, err = encodeInt16(m.packetId, payload)
	if err != nil {
		return nil, err
	}

	for _, code := range m.returnCodes {
		payload = append(payload, byte(code))
	}

	m.header.Length = len(payload)
	out, err = m.header.encode(out)
	if err != nil {
		return nil, err
	}
	out = append(out, payload...)

	return out, nil
}

// PublishRec

// PublishComp

// Disconnect
type messageDisconnect struct {
	header fixHeader

	// Variable Header
	// None

	// Payload
	// None
}

func newMessageDisconnect() *messageDisconnect {
	m := new(messageDisconnect)
	m.header.Type = MessageTypeDisconnect

	return m
}

func (m *messageDisconnect) String() string {
	s := "Disconnect { "
	s += m.header.String()
	s += ", Variable Header: None, Payload: None }"
	return s
}

func (m *messageDisconnect) setHeader(h fixHeader) {
	m.header = h
}
func (m *messageDisconnect) getHeader() fixHeader {
	return m.header
}
func (m *messageDisconnect) decodePayload(in []byte) (int, error) {
	return 0, nil
}
func (m *messageDisconnect) encode(out []byte) ([]byte, error) {
	panic("Don't use")
	return nil, nil
}
