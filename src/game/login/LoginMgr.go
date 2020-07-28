package login

import (
	"encoding/json"
	"github.com/liangdas/mqant/gate"
	"redisClient"
)

func NewLoginMgr(proxy *LoginProxy) *LoginMgr {
	mgr := new(LoginMgr)
	mgr.proxy = proxy
	return mgr
}

type LoginData struct {
	Uid     uint64
	Account string
	Psw     string
	session gate.Session
}

type LoginMgr struct {
	proxy *LoginProxy
}

func (this *LoginMgr) onUserLogin(account string, passWord string) (uint64, bool) {
	accountKey := redisClient.GetAccountKey(account)
	if result, _ := redisClient.Get(accountKey); result != nil {
		data := &LoginData{}
		if e := json.Unmarshal(result, data); e == nil {
			if data.Psw == passWord {
				return data.Uid, true
			}
		} else {
			return 0, false
		}
	}

	newUserID := this.getNewUID()
	if newUserID == 0 {
		return 0, false
	}

	loginData := &LoginData{
		Uid:     newUserID,
		Account: account,
		Psw:     passWord,
	}

	if saveData, e := json.Marshal(loginData); e == nil {
		redisClient.Set(accountKey, string(saveData))
	} else {
		return 0, false
	}
	return newUserID, true
}

func (this *LoginMgr) onUserLogout(uid string) {

}

func (this *LoginMgr) getNewUID() uint64 {
	key := "USER_ID_AUTO_INCR_KEY"
	userID, _ := redisClient.IncrKey(key)
	return uint64(userID)
}
