package messages

import "strconv"

type CornerItemChangeMessage struct {
	Message
	Item      int    `json:"item"`
	SenderID  int64  `json:"senderId"`
	Direction string `json:"direction,omitempty"` // Optional field for direction, if needed
}

func NewCornerItemChangeMessage(item int, senderID int64, direction string) CornerItemChangeMessage {
	return CornerItemChangeMessage{
		Message: Message{
			MessageType: CornerItemChange,
		},
		Item:      item,
		SenderID:  senderID,
		Direction: direction,
	}
}

func (m CornerItemChangeMessage) String() string {
	return "CornerItemChangeMessage{" +
		"Item: " + strconv.Itoa(m.Item) +
		", SenderID: " + strconv.FormatInt(m.SenderID, 10) +
		", Direction: " + m.Direction +
		"}"
}

func (m CornerItemChangeMessage) Type() MessageType {
	return m.MessageType
}
