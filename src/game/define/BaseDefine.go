package define

//版本号
const VERSION = "1.0.0.1"

//服务器类型
const (
	SERVER_TYPE_GATE = "Gate"
	SERVER_TYPE_LOGIN = "Login"
	SERVER_TYPE_DBSVR = "DBSvr"
	SERVER_TYPE_HALL = "Hall"
)

var SERVER_NAMES = []string{
	SERVER_TYPE_GATE,
	SERVER_TYPE_LOGIN,
	SERVER_TYPE_DBSVR,
	SERVER_TYPE_HALL,
}

//表名枚举
var USER_DATA_NAME = []string {
	"ROLE_DATA",	//角色数据
}
