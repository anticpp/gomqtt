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

	h.Length, n, err = decodeVariableInt4(in[1:])
	if err != nil {
		return 0, err
	}
	return n + 1, nil
}

type messageConnect struct {
}

type messageConnectAck struct {
}
