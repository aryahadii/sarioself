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

func sessionStartHandler(userSession *miyanbor.UserSession, matches []string, update interface{}) {
}

func startCommandHandler(userSession *miyanbor.UserSession, matches []string, update interface{}) {
	welcomeMessage := telegramAPI.NewMessage(userSession.ChatID, text.MsgWelcome)
	welcomeMessage.ReplyMarkup = generateMainKeyboard()
	Bot.Send(welcomeMessage)
}

func menuCommandHandler(userSession *miyanbor.UserSession, matches []string, update interface{}) {
	userInfo, err := getUserInfo(userSession)
	if err != nil {
		return
	}

	// Create client
	samadClient, err := selfservice.NewSamadAUTClient(userInfo.StudentID, userInfo.Password)
	if err != nil {
		logrus.Errorf("can't create new Samad client, %v", err)
		msg := telegramAPI.NewMessage(userSession.ChatID, text.MsgAnErrorOccured)
		Bot.Send(msg)
		return
	}

	// Get foods list
	foods, err := samadClient.GetAvailableFoods()
	if err != nil {
		logrus.Errorf("can't GetAvailableFoods, %v", err)
		msg := telegramAPI.NewMessage(userSession.ChatID, text.MsgAnErrorOccured)
		Bot.Send(msg)
		return
	}
	sortedFoods := sortFoodsByTime(foods)

	// Send menu
	keyboard := generateMenuKeyboard(sortedFoods)
	menuMsgText := generateMenuMessage(sortedFoods)
	msg := telegramAPI.NewMessage(userSession.ChatID, menuMsgText)
	msg.ReplyMarkup = keyboard
	Bot.Send(msg)
}

func creditCommandHandler(userSession *miyanbor.UserSession, matches []string, update interface{}) {
	userInfo, err := getUserInfo(userSession)
	if err != nil {
		return
	}

	// Create client
	samadClient, err := selfservice.NewSamadAUTClient(userInfo.StudentID, userInfo.Password)
	if err != nil {
		logrus.Errorf("can't create new Samad client, %v", err)
		msg := telegramAPI.NewMessage(userSession.ChatID, text.MsgAnErrorOccured)
		Bot.Send(msg)
		return
	}

	// Get credit
	credit, err := samadClient.GetCredit()
	if err != nil {
		logrus.Errorf("can't GetCredit, %v", err)
		msg := telegramAPI.NewMessage(userSession.ChatID, text.MsgAnErrorOccured)
		Bot.Send(msg)
		return
	}

	// Send message
	formattedCredit := fmt.Sprintf("%v تومان", credit/10)
	msg := telegramAPI.NewMessage(userSession.ChatID, formattedCredit)
	Bot.Send(msg)
}

func foodReserveMessageHandler(userSession *miyanbor.UserSession, matches []string, update interface{}) {
	if matches == nil {
		sendErrorMsg(userSession.ChatID)
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
		sendErrorMsg(userSession.ChatID)
		return
	}

	// Reserve
	unixTime, err := strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		msg := telegramAPI.NewMessage(userSession.ChatID, text.MsgAnErrorOccured)
		Bot.Send(msg)
		return
	}
	mealTime := time.Unix(unixTime, 0)
	toggled, err := samadClient.ToggleFoodReservation(&mealTime, matches[1])
	if err != nil || !toggled {
		if samadError, ok := err.(selfservice.SamadError); ok {
			sendCustomErrorMsg(userSession.ChatID, samadError.What)
		} else {
			sendErrorMsg(userSession.ChatID)
		}
		return
	}

	// Success message
	successMessage := telegramAPI.NewMessage(userSession.ChatID, text.MsgReservationToggleSuccess)
	Bot.Send(successMessage)
}

func unknownMessageHandler(userSession *miyanbor.UserSession, matches []string, update interface{}) {
	logrus.Errorln("Unknown Message", *userSession, update)
}

func enterStudentIDCallback(userSession *miyanbor.UserSession, matches []string, update interface{}) {
	userSession.Payload["student-id"] = matches[0]
	Bot.AskStringQuestion(text.MsgEnterPassword, userSession.UserID,
		userSession.ChatID, enterPasswordCallback)
}

func enterPasswordCallback(userSession *miyanbor.UserSession, matches []string, update interface{}) {
	userSession.Payload["password"] = matches[0]

	// Add data to database
	userInfo := model.User{
		UserID:    userSession.UserID,
		StudentID: userSession.Payload["student-id"].(string),
		Password:  userSession.Payload["password"].(string),
	}
	db.GetInstance().Create(&userInfo)

	// Send success message
	successfulMsg := telegramAPI.NewMessage(userSession.ChatID, text.MsgProfileSuccess)
	Bot.Send(successfulMsg)
}
