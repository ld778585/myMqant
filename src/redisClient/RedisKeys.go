package redisClient

import (
	"fmt"
	"strconv"
)

func GetSessionKey(userID string) string {
	key := fmt.Sprintf("session:%s", userID)
	return key
}

func GetAccountKey(account string) string {
	key := fmt.Sprintf("Account:%s", account)
	return key
}

func GetUserDataKey(name string,uid int64) string{
	key := fmt.Sprintf("%s:%s", name,strconv.FormatInt(int64(uid), 10))
	return key
}