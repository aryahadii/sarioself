package miyanbor

type UserSession struct {
	userID          int
	chatID          int64
	payload         map[string]interface{}
	messageCallback CallbackFunction
}

func (us *UserSession) GetUserID() int {
	return us.userID
}

func (us *UserSession) GetChatID() int64 {
	return us.chatID
}

func (us *UserSession) GetPayload() map[string]interface{} {
	return us.payload
}
