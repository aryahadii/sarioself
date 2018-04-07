package miyanbor

import (
	"fmt"

	cache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	telegramAPI "gopkg.in/telegram-bot-api.v4"
)

var (
	callbackQueryCallbacks       []callback
	messagesCallbacks            []callback
	commandsCallbacks            []callback
	sessionStartCallbackFunction CallbackFunction
	fallbackCallbackFunction     CallbackFunction

	usersSessionCache *cache.Cache
)

func (b *Bot) handleNewUpdate(update *telegramAPI.Update) {
	// Try to get user session from cache
	userSession, err := getUserSession(update)
	if err != nil {
		logrus.WithError(err).Errorf("can't get user session")
		return
	}

	// Create user session if it doesn't exits
	if userSession == nil {
		userSession = createUserSession(update)
		if sessionStartCallbackFunction != nil {
			sessionStartCallbackFunction(userSession, nil, update)
		}
	}

	var updateHandled bool

	// Find and call callback function
	if update.CallbackQuery != nil {
		for _, callback := range callbackQueryCallbacks {
			if matches := callback.Pattern.FindStringSubmatch(update.CallbackQuery.Data); matches != nil {
				callback.Function(userSession, matches, update)
				updateHandled = true
				break
			}
		}
	} else if update.Message != nil {
		if update.Message.IsCommand() {
			for _, callback := range commandsCallbacks {
				if matches := callback.Pattern.FindStringSubmatch(update.Message.Command()); matches != nil {
					callback.Function(userSession, matches, update)
					updateHandled = true
					break
				}
			}
		} else {
			if userSession.messageCallback != nil {
				userSession.messageCallback(userSession, nil, update)
				updateHandled = true
			} else {
				for _, callback := range messagesCallbacks {
					if matches := callback.Pattern.FindStringSubmatch(update.Message.Text); matches != nil {
						callback.Function(userSession, matches, update)
						updateHandled = true
						break
					}
				}
			}
		}
	} else {
		logrus.Errorf("unknown update")
	}

	// Call fallback callback function
	if !updateHandled {
		fallbackCallbackFunction(userSession, nil, update)
	}
}

func getSenderChatID(update *telegramAPI.Update) (int64, error) {
	if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.Chat.ID, nil
	}
	if update.Message != nil {
		return update.Message.Chat.ID, nil
	}
	return 0, fmt.Errorf("can't get ChatID from update")
}

func getSenderUser(update *telegramAPI.Update) (*telegramAPI.User, error) {
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From, nil
	}
	if update.Message != nil {
		return update.Message.From, nil
	}
	return nil, fmt.Errorf("can't get User from update")
}

func getUserSession(update *telegramAPI.Update) (*UserSession, error) {
	chatID, err := getSenderChatID(update)
	if err != nil {
		return nil, err
	}

	user, err := getSenderUser(update)
	if err != nil {
		return nil, err
	}

	// Find UserSession
	var userSession *UserSession
	userSessionInterface, found := usersSessionCache.Get(getUserSessionKey(user.ID, chatID))
	if found {
		userSession = userSessionInterface.(*UserSession)
		userSession.User = user
	}
	return userSession, nil
}

func createUserSession(update *telegramAPI.Update) *UserSession {
	chatID, _ := getSenderChatID(update)
	user, _ := getSenderUser(update)

	userSession := &UserSession{
		User:    user,
		UserID:  user.ID,
		ChatID:  chatID,
		Payload: make(map[string]interface{}),
	}
	usersSessionCache.Add(getUserSessionKey(user.ID, chatID), userSession, cache.DefaultExpiration)
	return userSession
}
