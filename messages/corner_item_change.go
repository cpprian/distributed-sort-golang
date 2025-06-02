package messages

type CornerItemChangeMessage[T any] struct {
	Message
	Item     T     `json:"item"`
	SenderID int64 `json:"senderId"`
}

func NewCornerItemChangeMessage[T any](item T, senderID int64) CornerItemChangeMessage[T] {
	return CornerItemChangeMessage[T]{
		Message: Message{
			MessageType: CornerItemChange,
		},
		Item:     item,
		SenderID: senderID,
	}
}
