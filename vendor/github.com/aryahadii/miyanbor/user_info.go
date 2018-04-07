package miyanbor

import telegramAPI "gopkg.in/telegram-bot-api.v4"

type UserInfo struct {
	telegramAPI.User
	UserID int
	ChatID int64
}
