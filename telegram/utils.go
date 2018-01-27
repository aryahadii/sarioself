package telegram

import (
	"fmt"
	"time"

	"github.com/aryahadii/miyanbor"
	"github.com/aryahadii/sarioself/db"
	"github.com/aryahadii/sarioself/model"
	"github.com/aryahadii/sarioself/ui/text"
	"github.com/yaa110/go-persian-calendar/ptime"
	"gopkg.in/mgo.v2/bson"
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

func getFormattedTime(time time.Time) string {
	jalaliDate := ptime.New(time)
	return fmt.Sprintf("%s %dام", weekdays[int(jalaliDate.Weekday())], jalaliDate.Day())
}
