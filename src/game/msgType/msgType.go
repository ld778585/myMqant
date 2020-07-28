package msgType

import "fmt"

type MessageType struct {
	//模块名字
	ModuleName string
	//消息名字
	MessageName string
	//发送到客户端的消息名字
	SendMsgName string
	//提示信息
	Tip string
}

func NewMessageType(moduleName string, msgName string, tip string) *MessageType {
	mt := &MessageType{}
	mt.ModuleName = moduleName
	mt.MessageName = msgName
	mt.Tip = tip
	mt.SendMsgName = fmt.Sprintf("%s/%s", moduleName, msgName)
	return mt
}

//登录相关
var (
	CS_USER_LOGIN              *MessageType //客户端请求用户登录
	RPC_USER_LOGOUT            *MessageType //用户下线
	RPC_LOAD_USER_INFO_FROM_DB *MessageType //从db加载用户数据
	RPC_USER_LOGIN_SUCCESS     *MessageType //用户登录成功
)

func init() {
	CS_USER_LOGIN = NewMessageType("Login", "userLogin", "客户端请求用户登录")
	RPC_USER_LOGOUT = NewMessageType("", "userLogout", "用户下线")
	RPC_LOAD_USER_INFO_FROM_DB = NewMessageType("DBSvr", "loadUserInfo", "从db加载用户数据")
	RPC_USER_LOGIN_SUCCESS = NewMessageType("Hall", "userLoginSuccess", "用户登录成功")
}
