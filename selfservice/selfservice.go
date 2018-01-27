package selfservice

import (
	"net/http/cookiejar"

	"github.com/aryahadii/sarioself/model"
)

type userSessionData struct {
	username string
	password string
	csrf     string
	jar      *cookiejar.Jar
}

// Client is interface for clients of restaurants
type Client interface {
	GetAvailableFoods() [][]*model.Food
	ReserveFood(food *model.Food) error
}
