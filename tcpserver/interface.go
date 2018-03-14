package tcpserver

import(
	"go_project/proto"
)

type MessageProcessor interface {
	ProcessMsg(msg *proto.Message) (*proto.Message)
	
}