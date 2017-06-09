package model

import (
	"github.com/inkyblackness/shocked-model"
)

// ElectronicMessageAdapter is the entry point for an electronic message.
type ElectronicMessageAdapter struct {
	context archiveContext
	store   model.DataStore

	messageType model.ElectronicMessageType
	id          int
	data        *observable
}

func newElectronicMessageAdapter(context archiveContext, store model.DataStore) *ElectronicMessageAdapter {
	adapter := &ElectronicMessageAdapter{
		context: context,
		store:   store,

		data: newObservable()}

	adapter.clear()

	return adapter
}

func (adapter *ElectronicMessageAdapter) clear() {
	adapter.id = -1
	var message model.ElectronicMessage
	adapter.data.set(&message)
}

func (adapter *ElectronicMessageAdapter) messageData() *model.ElectronicMessage {
	return adapter.data.get().(*model.ElectronicMessage)
}

// OnMessageDataChanged registers a callback for data changes.
func (adapter *ElectronicMessageAdapter) OnMessageDataChanged(callback func()) {
	adapter.data.addObserver(callback)
}

// ID returns the ID of the electronic message.
func (adapter *ElectronicMessageAdapter) ID() int {
	return adapter.id
}

// RequestMessage requests to load the message data of specified ID
func (adapter *ElectronicMessageAdapter) RequestMessage(messageType model.ElectronicMessageType, id int) {
	adapter.clear()
	adapter.id = id
	adapter.messageType = messageType
	adapter.store.ElectronicMessage(adapter.context.ActiveProjectID(), messageType, id,
		func(message model.ElectronicMessage) { adapter.onMessageData(messageType, id, message) },
		adapter.context.simpleStoreFailure("ElectronicMessage"))
}

func (adapter *ElectronicMessageAdapter) onMessageData(messageType model.ElectronicMessageType, id int, message model.ElectronicMessage) {
	if (adapter.messageType == messageType) && (adapter.id == id) {
		adapter.data.set(&message)
	}
}

// Title returns the title of the message.
func (adapter *ElectronicMessageAdapter) Title(language int) string {
	return safeString(adapter.messageData().Title[language])
}

// Sender returns the sender of the message.
func (adapter *ElectronicMessageAdapter) Sender(language int) string {
	return safeString(adapter.messageData().Sender[language])
}

// Subject returns the subject of the message.
func (adapter *ElectronicMessageAdapter) Subject(language int) string {
	return safeString(adapter.messageData().Subject[language])
}

// VerboseText returns the text in long form of the message.
func (adapter *ElectronicMessageAdapter) VerboseText(language int) string {
	return safeString(adapter.messageData().VerboseText[language])
}

// TerseText returns the text in short form of the message.
func (adapter *ElectronicMessageAdapter) TerseText(language int) string {
	return safeString(adapter.messageData().TerseText[language])
}
