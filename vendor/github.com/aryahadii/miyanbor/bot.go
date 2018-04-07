package miyanbor

import (
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	telegramAPI "gopkg.in/telegram-bot-api.v4"
)

// Bot contains API functions and params to contact Telegram servers.
type Bot struct {
	*telegramAPI.BotAPI
	updateConfig  telegramAPI.UpdateConfig
	updateChannel telegramAPI.UpdatesChannel
	Timeout       int
}

// NewBot creates new instance of Bot struct, also creates BotAPI using token.
func NewBot(token string, verboseLogging bool, sessionTimeout int) (*Bot, error) {
	if verboseLogging {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.ErrorLevel)
	}

	var err error
	telegramBot := &Bot{}
	telegramBot.BotAPI, err = telegramAPI.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	usersSessionCache = cache.New(time.Duration(sessionTimeout)*time.Minute,
		time.Duration(sessionTimeout)*time.Minute)

	return telegramBot, nil
}

// StartUpdater initializes updater channel.
func (b *Bot) StartUpdater(offset, timeout int) error {
	// Create update config
	b.updateConfig = telegramAPI.NewUpdate(offset)
	b.updateConfig.Timeout = timeout

	// Init update channel
	var err error
	b.updateChannel, err = b.GetUpdatesChan(b.updateConfig)
	if err != nil {
		return err
	}

	// Get updates
	for update := range b.updateChannel {
		logrus.Infof("new update")
		go func(update telegramAPI.Update) {
			startTime := time.Now()
			b.handleNewUpdate(&update)
			logrus.WithField("took", time.Since(startTime)).
				Infof("update handled!")
		}(update)
	}

	return nil
}
