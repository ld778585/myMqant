package redisClient

import (
	"fmt"
)

func GetSessionKey(userID string) string {
	key := fmt.Sprintf("session:%s", userID)
	return key
}