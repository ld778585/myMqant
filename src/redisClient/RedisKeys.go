package redisClient

import (
	"fmt"
)

func GetSessionKey(userID string) string {
	key := fmt.Sprintf("session:%s", userID)
	return key
}

func GetAccountKey(account string) string {
	key := fmt.Sprintf("Account:%s", account)
	return key
}