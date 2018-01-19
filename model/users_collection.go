package model

import "gopkg.in/mgo.v2/bson"

type User struct {
	ID                 bson.ObjectId `bson:"_id,omitempty"`
	UserID             int           `bson:"user-id"`
	ReservationService string        `bson:"reservation-service"`
	StudentID          string        `bson:"student-id"`
	Password           string        `bson:"password"`
}
