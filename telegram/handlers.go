package telegram

import (
	"github.com/aryahadii/miyanbor"
	"github.com/aryahadii/sarioself/db"
	"github.com/aryahadii/sarioself/model"
	"github.com/aryahadii/sarioself/ui/text"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	telegramAPI "gopkg.in/telegram-bot-api.v4"
)

func sessionStartHandler(userSession *miyanbor.UserSession, input interface{}) {
	var userInfo model.User
	err := db.UsersCollection.Find(bson.M{"user-id": userSession.GetUserID()}).One(&userInfo)
	if err != nil {
		Bot.AskStringQuestion(text.MsgEnterStudentID, userSession.GetUserID(),
			userSession.GetChatID(), enterStudentIDCallback)
	}
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
