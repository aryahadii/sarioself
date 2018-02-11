package telegram

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aryahadii/miyanbor"
	"github.com/aryahadii/sarioself/db"
	"github.com/aryahadii/sarioself/model"
	"github.com/aryahadii/sarioself/selfservice"
	"github.com/aryahadii/sarioself/ui/text"
	"github.com/sirupsen/logrus"
	telegramAPI "gopkg.in/telegram-bot-api.v4"
)

func sessionStartHandler(userSession *miyanbor.UserSession, input interface{}) {
}

func startCommandHandler(userSession *miyanbor.UserSession, matches interface{}) {
	welcomeMessage := telegramAPI.NewMessage(userSession.GetChatID(), text.MsgWelcome)
	welcomeMessage.ReplyMarkup = generateMainKeyboard()
	Bot.Send(welcomeMessage)
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
	sortedFoods := sortFoodsByTime(foods)

	// Send menu
	keyboard := generateMenuKeyboard(sortedFoods)
	menuMsgText := generateMenuMessage(sortedFoods)
	msg := telegramAPI.NewMessage(userSession.GetChatID(), menuMsgText)
	msg.ReplyMarkup = keyboard
	Bot.Send(msg)
}

func creditCommandHandler(userSession *miyanbor.UserSession, matches interface{}) {
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

	// Get credit
	credit, err := samadClient.GetCredit()
	if err != nil {
		logrus.Errorf("can't GetCredit, %v", err)
		msg := telegramAPI.NewMessage(userSession.GetChatID(), text.MsgAnErrorOccured)
		Bot.Send(msg)
		return
	}

	// Send message
	formattedCredit := fmt.Sprintf("%v تومان", credit/10)
	msg := telegramAPI.NewMessage(userSession.GetChatID(), formattedCredit)
	Bot.Send(msg)
}

func foodReserveMessageHandler(userSession *miyanbor.UserSession, matches interface{}) {
	matchedGroups, ok := matches.([]string)
	if !ok {
		sendErrorMsg(userSession.GetChatID())
		return
	}

	userInfo, err := getUserInfo(userSession)
	if err != nil {
		return
	}

	// Create client
	samadClient, err := selfservice.NewSamadAUTClient(userInfo.StudentID, userInfo.Password)
	if err != nil {
		logrus.Errorf("can't create new Samad client, %v", err)
		sendErrorMsg(userSession.GetChatID())
		return
	}

	// Reserve
	unixTime, err := strconv.ParseInt(matchedGroups[2], 10, 64)
	if err != nil {
		msg := telegramAPI.NewMessage(userSession.GetChatID(), text.MsgAnErrorOccured)
		Bot.Send(msg)
		return
	}
	mealTime := time.Unix(unixTime, 0)
	toggled, err := samadClient.ToggleFoodReservation(&mealTime, matchedGroups[1])
	if err != nil || !toggled {
		if samadError, ok := err.(selfservice.SamadError); ok {
			sendCustomErrorMsg(userSession.GetChatID(), samadError.What)
		} else {
			sendErrorMsg(userSession.GetChatID())
		}
		return
	}

	// Success message
	successMessage := telegramAPI.NewMessage(userSession.GetChatID(), text.MsgReservationToggleSuccess)
	Bot.Send(successMessage)
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
