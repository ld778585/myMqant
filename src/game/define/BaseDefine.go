package define

//版本号
const VERSION = "1.0.0.1"

//服务器类型
const (
	SERVER_TYPE_GATE = "Gate"
	SERVER_TYPE_LOGIN = "Login"
)

var SERVER_NAMES = []string{
	SERVER_TYPE_GATE,
	SERVER_TYPE_LOGIN,
}