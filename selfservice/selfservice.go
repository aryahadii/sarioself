package selfservice

import (
	"net/http/cookiejar"
	"time"

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
	GetAvailableFoods() (map[time.Time][]*model.Food, error)
	ReserveFood(date *time.Time, foodID string) error
}
