package main
import (
	"log"
	"flag"	
	"go_project/tcpclient"
)


func main(){
	serverAddr := flag.String("server", "127.0.0.1:33333", "服务器的ip:port")
	flag.Parse()
	tcpClient := tcpclient.NewTcpClient(*serverAddr, 1024*1024)
	err := tcpClient.Start()
	if err != nil {
		log.Panicln("tcpClient start fail, err:", err.Error())
		return
	}
	//发送一个消息给server
	bodyBuf := "hello world"
	head := &tcpclient.ProtoHead{
		bodyLen : uint16(len(bodyBuf)),
		magic : 1,
		seq : 1,
	}
	msg := &tcpclient.Message{
		head : head,
		bodyBuf : []byte(bodyBuf),
	}
	err = tcpClient.Write(msg)
	if err != nil {
		log.Printf("write error, err:%s\n", err.Error())
		return
	}
	//等待server响应
	log.Println("waiting for response...")
	msg = tcpClient.GetMessage()
	if msg == nil {
		log.Println("get message err")
		return
	}
	log.Printf("head:magic[%d],seq[%d], body:%s\n", msg.head.magic, msg.head.seq, string(msg.bodyBuf))


}