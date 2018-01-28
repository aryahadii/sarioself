package telegram

import (
	"github.com/aryahadii/miyanbor"
	"github.com/aryahadii/sarioself/db"
	"github.com/aryahadii/sarioself/model"
	"github.com/aryahadii/sarioself/selfservice"
	"github.com/aryahadii/sarioself/ui/text"
	"github.com/sirupsen/logrus"
	telegramAPI "gopkg.in/telegram-bot-api.v4"
)

func sessionStartHandler(userSession *miyanbor.UserSession, input interface{}) {
	getUserInfo(userSession)
}

func menuCommandHandler(userSession *miyanbor.UserSession, matches interface{}) {
	userInfo, err := getUserInfo(userSession)
	if err != nil {
		return
	}

	// Create client
	samadClient, err := selfservice.NewSamadAUTClient(userInfo.StudentID, userInfo.Password)
	if err != nil {
		logrus.Errorf("can't create new Samad client, %v", err)
		msg := telegramAPI.NewMessage(userSession.GetChatID(), text.MsgAnErrorOccured)
		Bot.Send(msg)
		return
	}

	// Get foods list
	foods, err := samadClient.GetAvailableFoods()
	if err != nil {
		logrus.Errorf("can't GetAvailableFoods, %v", err)
		msg := telegramAPI.NewMessage(userSession.GetChatID(), text.MsgAnErrorOccured)
		Bot.Send(msg)
		return
	}

	// Send menu
	menuMsgText := generateMenuMessage(foods)
	msg := telegramAPI.NewMessage(userSession.GetChatID(), menuMsgText)
	Bot.Send(msg)
}

func reserveCommandHandler(userSession *miyanbor.UserSession, matches interface{}) {
	msg := telegramAPI.NewMessage(userSession.GetChatID(), "فعلا از این چیزها نداریم!")
	Bot.Send(msg)
}

func unknownMessageHandler(userSession *miyanbor.UserSession, input interface{}) {
	logrus.Errorln("Unknown Message", *userSession, input)
}

func enterStudentIDCallback(userSession *miyanbor.UserSession, input interface{}) {
	userSession.GetPayload()["student-id"] = input.(string)
	Bot.AskStringQuestion(text.MsgEnterPassword, userSession.GetUserID(),
		userSession.GetChatID(), enterPasswordCallback)
}

func enterPasswordCallback(userSession *miyanbor.UserSession, input interface{}) {
	userSession.GetPayload()["password"] = input.(string)

	// Add data to database
	userInfo := model.User{
		UserID:    userSession.GetUserID(),
		StudentID: userSession.GetPayload()["student-id"].(string),
		Password:  userSession.GetPayload()["password"].(string),
	}
	db.UsersCollection.Insert(userInfo)

	// Send success message
	successfulMsg := telegramAPI.NewMessage(userSession.GetChatID(), text.MsgProfileSuccess)
	Bot.Send(successfulMsg)
}
