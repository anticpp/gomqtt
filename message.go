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

type fixHeader struct {
	Type   int32
	Dup    int32
	Qos    int32
	Retain int32
	Length int32
}

func (h fixHeader) String() string {
	return fmt.Sprintf("{Type: %v, Dup: %v, Qos: %v, Retain: %v}",
		MessageTypeString[h.Type],
		h.Dup,
		QosString[h.Qos],
		h.Retain)
}

type messageConnect struct {
}

type messageConnectAck struct {
}
