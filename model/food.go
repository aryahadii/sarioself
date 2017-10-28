package model

import "time"

type FoodStatus int

const (
	// FoodStatusReservable means food is available for reserve
	FoodStatusReservable FoodStatus = 0 + iota
	// FoodStatusReserved means food have been reserved by user
	FoodStatusReserved
	// FoodStatusSecondOption means another food have been reserved for the day
	FoodStatusSecondOption
	// FoodStatusUnavailable means food cannot be reserved for some reason
	FoodStatusUnavailable
)

type MealTime int

const (
	// MealTimeBreakfast indicates breakfast
	MealTimeBreakfast MealTime = 0 + iota
	// MealTimeLunch indicates lunch
	MealTimeLunch
	// MealTimeDinner indicates dinner
	MealTimeDinner
)

// Food contains information of a food
type Food struct {
	Name        string
	SideDish    string
	PriceTooman string
	MealTime    MealTime
	Status      FoodStatus
	Date        *time.Time
}
