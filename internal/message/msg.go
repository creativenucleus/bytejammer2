package message

type MsgType string

const MsgTypeLog = MsgType("log")
const MsgTypeTicState = MsgType("tic-state")

// Msg is the base struct for representing some information that is passed around
type Msg struct {
	Type MsgType
	Data interface{}
}

// MsgReceiver is the interface that a type must implement to receive messages
type MsgReceiver interface {
	MsgHandler(message Msg) error
}

// MsgSender can be embedded in a struct to allow it to send messages to registered receivers
type MsgSender struct {
	messageReceivers []MsgReceiver
}

// AddReceiver registers a receiver to a recieve Msgs sent by a MsgSender
func (b *MsgSender) AddReceiver(r MsgReceiver) {
	b.messageReceivers = append(b.messageReceivers, r)
}

// RemoveReceiver unregisters a receiver to a MsgSender
func (b *MsgSender) RemoveReceiver(r MsgReceiver) {
	for i, receiver := range b.messageReceivers {
		if receiver == r {
			b.messageReceivers = append(b.messageReceivers[:i], b.messageReceivers[i+1:]...)
			break
		}
	}
}

// Send sends a message to all registered receivers
func (b *MsgSender) Send(message Msg) {
	for _, receiver := range b.messageReceivers {
		receiver.MsgHandler(message)
	}
}
