package miyanbor

import "strconv"

func getUserSessionKey(userID int, chatID int64) string {
	return strconv.Itoa(userID) + "#" + strconv.FormatInt(chatID, 10)
}
