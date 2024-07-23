package main

type Message struct {
	Type string
	Data interface{}
}

func NewMessageTicState(state TicState) Message {
	return Message{Type: "tic-state", Data: state}
}

type MessageReceiver interface {
	messageHandler(message *Message)
}

type MessageBroadcaster struct {
	messageReceivers []MessageReceiver
}

func (b *MessageBroadcaster) addReceiver(r MessageReceiver) {
	b.messageReceivers = append(b.messageReceivers, r)
}

func (b *MessageBroadcaster) removeReceiver(r MessageReceiver) {
	for i, receiver := range b.messageReceivers {
		if receiver == r {
			b.messageReceivers = append(b.messageReceivers[:i], b.messageReceivers[i+1:]...)
			break
		}
	}
}

func (b *MessageBroadcaster) broadcast(message *Message) {
	for _, receiver := range b.messageReceivers {
		receiver.messageHandler(message)
	}
}
