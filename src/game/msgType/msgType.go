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
	RPC_USER_LOGOUT            *MessageType //用户下线
)

func init() {
	RPC_USER_LOGOUT = NewMessageType("", "logOut", "用户下线")
}