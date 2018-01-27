package telegram

import (
	"github.com/aryahadii/miyanbor"
	"github.com/aryahadii/sarioself/db"
	"github.com/aryahadii/sarioself/model"
	"github.com/aryahadii/sarioself/ui/text"
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
