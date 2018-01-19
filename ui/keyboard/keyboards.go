package keyboard

import (
	telegramBotAPI "gopkg.in/telegram-bot-api.v4"
)

func NewMainKeyboard() telegramBotAPI.ReplyKeyboardMarkup {
	foodSchedule := telegramBotAPI.NewKeyboardButton(ui.text.BtnTextFoodSchedule)
	row1 := telegramBotAPI.NewKeyboardButtonRow(foodSchedule)

	return telegramBotAPI.NewReplyKeyboard(row1)
}
