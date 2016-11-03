package gomqtt

type sessionPoolType []*sessionType

type eventTreeType struct {
	// Topic - Session
	tree map[string]sessionPoolType

	in chan *messagePub
}

func newEventTree() *eventTreeType {

	return &eventTreeType{tree: make(map[string]sessionPoolType),
		in: make(chan *messagePub, 100)}
}

func (t *eventTreeType) sendEvent(message *messagePub) {
	t.in <- message
}

func (t *eventTreeType) start() {

}
