package selfservice

import (
	"net/http"
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
	GetAvailableFoods(sessionData *userSessionData, client *http.Client) [][]*model.Food
	ReserveFood(food *model.Food, sessionData *userSessionData, client *http.Client) error
}
