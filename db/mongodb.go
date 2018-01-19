package db

import (
	"github.com/aryahadii/sarioself/configuration"
	"github.com/pkg/errors"
	mgo "gopkg.in/mgo.v2"
)

var (
	UsersCollection *mgo.Collection
	session         *mgo.Session
)

const (
	dbName              = "sarioself"
	usersCollectionName = "users"
)

func InitMongoDB() error {
	var err error
	session, err = mgo.Dial(configuration.SarioselfConfig.GetString("mongodb.address"))
	if err != nil {
		errors.Wrap(err, "MongoDB session can't be created")
	}
	UsersCollection = session.DB(dbName).C(usersCollectionName)
	return nil
}

func Close() {
	session.Close()
}
