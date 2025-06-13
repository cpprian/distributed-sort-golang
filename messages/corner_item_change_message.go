package messages

type CornerItemChangeMessage struct {
	BaseMessage
	Item     int64 `json:"item"`
	SenderID int64 `json:"senderId"`
}

func NewCornerItemChangeMessage(item int64, senderID int64) CornerItemChangeMessage {
	return CornerItemChangeMessage{
		BaseMessage: NewBaseMessage(CornerItemChange),
		Item:        item,
		SenderID:    senderID,
	}
}

func (m CornerItemChangeMessage) GetItem() int64 {
	return m.Item
}
func (m CornerItemChangeMessage) GetSenderID() int64 {
	return m.SenderID
}
