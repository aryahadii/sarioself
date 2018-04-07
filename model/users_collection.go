package model

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	UserID             int
	ReservationService string
	StudentID          string
	Password           string
}
