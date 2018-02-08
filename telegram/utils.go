package telegram

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/aryahadii/miyanbor"
	"github.com/aryahadii/sarioself/db"
	"github.com/aryahadii/sarioself/model"
	"github.com/aryahadii/sarioself/ui/text"
	"github.com/yaa110/go-persian-calendar/ptime"
	"gopkg.in/mgo.v2/bson"
	telegramAPI "gopkg.in/telegram-bot-api.v4"
)

func getUserInfo(userSession *miyanbor.UserSession) (*model.User, error) {
	var userInfo model.User
	err := db.UsersCollection.Find(bson.M{"user-id": userSession.GetUserID()}).One(&userInfo)
	if err != nil {
		Bot.AskStringQuestion(text.MsgEnterStudentID, userSession.GetUserID(),
			userSession.GetChatID(), enterStudentIDCallback)
		return nil, err
	}
	return &userInfo, nil
}

var (
	weekdays = map[int]string{
		0: "شنبه",
		1: "یک‌شنبه",
		2: "دوشنبه",
		3: "سه‌شنبه",
		4: "چهارشنبه",
		5: "پنج‌شنبه",
		6: "جمعه",
	}
)

func getFormattedDayWeekday(time time.Time) string {
	jalaliDate := ptime.New(time)
	return fmt.Sprintf("%s %dام", weekdays[int(jalaliDate.Weekday())], jalaliDate.Day())
}

func getFormattedWeekday(time time.Time) string {
	jalaliDate := ptime.New(time)
	return fmt.Sprintf("%s", weekdays[int(jalaliDate.Weekday())])
}

func generateMenuKeyboard(foods []*model.Food) *telegramAPI.InlineKeyboardMarkup {
	rows := [][]telegramAPI.InlineKeyboardButton{}
	for _, food := range foods {
		formattedTime := getFormattedWeekday(*food.Date)
		caption := fmt.Sprintf(text.MsgKeyboardFoodItem, formattedTime, food.Name)
		data := fmt.Sprintf(text.FoodInlineButtonData, food.ID, strconv.FormatInt(food.Date.Unix(), 10))
		button := telegramAPI.NewInlineKeyboardButtonData(caption, data)

		row := telegramAPI.NewInlineKeyboardRow(button)
		rows = append(rows, row)
	}
	markup := telegramAPI.NewInlineKeyboardMarkup(rows...)
	return &markup
}

func generateMenuMessage(foods []*model.Food) string {
	menuMsgText := ""
	for _, food := range foods {
		formattedTime := getFormattedDayWeekday(*food.Date)
		if food.Status == model.FoodStatusUnavailable {
			menuMsgText += fmt.Sprintf(text.MsgNotSelectableFoodMenuItem,
				formattedTime, food.Name, food.SideDish, strconv.Itoa(food.PriceTooman))
		} else if food.Status == model.FoodStatusReserved {
			menuMsgText += fmt.Sprintf(text.MsgSelectedFoodMenuItem,
				formattedTime, food.Name, food.SideDish, strconv.Itoa(food.PriceTooman))
		} else {
			menuMsgText += fmt.Sprintf(text.MsgNotSelectedFoodMenuItem,
				formattedTime, food.Name, food.SideDish, strconv.Itoa(food.PriceTooman))
		}
	}
	return menuMsgText
}

func sortFoodsByTime(foods map[time.Time][]*model.Food) []*model.Food {
	keys := []time.Time{}
	for key := range foods {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i int, j int) bool {
		return keys[i].Before(keys[j])
	})

	sortedFoods := []*model.Food{}
	for _, key := range keys {
		sortedFoods = append(sortedFoods, foods[key]...)
	}
	return sortedFoods
}

func sendErrorMsg(chatID int64) {
	msg := telegramAPI.NewMessage(chatID, text.MsgAnErrorOccured)
	Bot.Send(msg)
}

func sendCustomErrorMsg(chatID int64, errorMessage string) {
	msg := telegramAPI.NewMessage(chatID, errorMessage)
	Bot.Send(msg)
}
