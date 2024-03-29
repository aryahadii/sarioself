package telegram

import (
	"github.com/aryahadii/miyanbor"
	"github.com/aryahadii/sarioself/configuration"
	"github.com/sirupsen/logrus"
)

var (
	Bot *miyanbor.Bot
)

const (
	foodReservePattern = `RES#(?P<weekday>[\w.]+)#(?P<foodid>\d+)`
)

// StartBot makes telegram bot ready and starts it's updater
func StartBot() {
	logrus.Infof("Telegram bot is going to start")

	token := configuration.SarioselfConfig.GetString("bots.telegram.token")
	debug := (configuration.SarioselfConfig.GetBool("bots.telegram.debug") &&
		configuration.SarioselfConfig.GetBool("bots.telegram.debug"))
	sessionTimeout := configuration.SarioselfConfig.GetInt("bots.telegram.session-timeout")
	updaterTimeout := configuration.SarioselfConfig.GetInt("bots.telegram.updater-timeout")

	var err error
	Bot, err = miyanbor.NewBot(token, debug, sessionTimeout)
	if err != nil {
		logrus.Fatalln(err)
	}
	setCallbacks(Bot)
	Bot.StartUpdater(0, updaterTimeout)
}

func setCallbacks(bot *miyanbor.Bot) {
	bot.SetSessionStartCallbackHandler(sessionStartHandler)
	bot.SetFallbackCallbackHandler(unknownMessageHandler)

	bot.AddCommandHandler("start", startCommandHandler)
	bot.AddCommandHandler("credit", creditCommandHandler)
	bot.AddCommandHandler("menu", menuCommandHandler)

	bot.AddMessageHandler("اعتبار", creditCommandHandler)
	bot.AddMessageHandler("منو", menuCommandHandler)

	bot.AddCallbackHandler(foodReservePattern, foodReserveMessageHandler)
}
