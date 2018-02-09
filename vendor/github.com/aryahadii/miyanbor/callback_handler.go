package miyanbor

import (
	"regexp"

	"github.com/pkg/errors"
)

func (b *Bot) SetSessionStartCallbackHandler(function CallbackFunction) {
	sessionStartCallbackFunction = function
}

func (b *Bot) SetFallbackCallbackHandler(function CallbackFunction) {
	fallbackCallbackFunction = function
}

func (b *Bot) AddCallbackHandler(callbackQueryPattern string, function CallbackFunction) error {
	// Compile pattern
	regexPattern, err := regexp.Compile(callbackQueryPattern)
	if err != nil {
		return errors.Wrap(err, "can't compile pattern")
	}

	// Add to callbacks list
	callbacks = append(callbacks, callback{
		Pattern:  regexPattern,
		Function: function,
	})
	return nil
}

func (b *Bot) AddCommandHandler(commandPattern string, function CallbackFunction) error {
	// Compile pattern
	regexPattern, err := regexp.Compile(commandPattern)
	if err != nil {
		return errors.Wrap(err, "can't compile pattern")
	}

	// Add to callbacks list
	commandsCallbacks = append(commandsCallbacks, callback{
		Pattern:  regexPattern,
		Function: function,
	})
	return nil
}
