package message

type MsgType string

// log is a log message that can be propagated
const MsgTypeLog = MsgType("log")

// tic-state is some code with or without a run signal in the comment
const MsgTypeTicState = MsgType("tic-state")

// tic-snapshot is an instruction to take a copy of the current TIC code
const MsgTypeTicSnapshot = MsgType("tic-snapshot")

// A request from the kiosk to make a snapshot
const MsgTypeKioskMakeSnapshot = MsgType("kiosk-make-snapshot")

type MsgDataMakeSnapshot struct {
	PlayerName string `json:"player_name"`
	EffectName string `json:"effect_name"`
}

// A request from the kiosk to prep a new player
const MsgTypeKioskNewPlayer = MsgType("kiosk-new-player")

// Msg is the base struct for representing some information that is passed around
type MsgHeader struct {
	Type MsgType
}

type Msg struct {
	Type       MsgType        `json:"type"`
	Data       map[string]any `json:"data"`
	StringData string         `json:"string_data"`
}

// MsgReceiver is the interface that a type must implement to receive messages
type MsgReceiver interface {
	MsgHandler(msgType MsgType, data []byte) error
}

// MsgSender can be embedded in a struct to allow it to send messages to registered receivers
type MsgPropagator struct {
	messageReceivers []MsgReceiver
}

// AddReceiver registers a receiver to a recieve Msgs sent by a MsgSender
func (b *MsgPropagator) AddReceiver(r MsgReceiver) {
	b.messageReceivers = append(b.messageReceivers, r)
}

// RemoveReceiver unregisters a receiver to a MsgSender
func (b *MsgPropagator) RemoveReceiver(r MsgReceiver) {
	for i, receiver := range b.messageReceivers {
		if receiver == r {
			b.messageReceivers = append(b.messageReceivers[:i], b.messageReceivers[i+1:]...)
			break
		}
	}
}

// Send sends a message to all registered receivers
func (b *MsgPropagator) Propagate(msgType MsgType, msgData []byte) {
	for _, receiver := range b.messageReceivers {
		receiver.MsgHandler(msgType, msgData)
	}
}
