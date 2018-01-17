package main

import(
	"log"
	"flag"
	"time"
	"go_project/tcpclient"
)
var exitChan chan bool


//msgCnt:每次连接上server后，发送消息个数。发送完后关闭连接。
//reconnectCnt:关闭连接后重新连接次数。
//connectionId:连接的ID
func startAClient(serverAddr string, msgCnt int, reconnectCnt int, connectionId uint32){
	bodyBuf := "hello world"
	head := &tcpclient.ProtoHead{
		BodyLen : uint16(len(bodyBuf)),
		Magic : 1,
		Seq : connectionId,
	}
	msg := &tcpclient.Message{
		Head : head,
		BodyBuf : []byte(bodyBuf),
	}
	log.Panicf("client:%d starting...\n", connectionId)
	for i := 0; i < reconnectCnt; i++ {
		cli := tcpclient.NewTcpClient(serverAddr, 1024)
		err := cli.Start()
		if err != nil {
			log.Println("tcpClient start fail, err:", err.Error())
			return
		} 
		for j := 0; j < msgCnt; j++{
			//发送一个消息给server
			err = cli.Write(msg)
			if err != nil {
				log.Printf("write error, err:%s\n", err.Error())
				return
			}
			//等待server响应
			//log.Println("waiting for response...")
			msg = cli.GetMessage()
			if msg == nil {
				log.Println("get message err")
				return
			}
			//log.Printf("head:magic[%d],seq[%d], body:%s\n", msg.Head.Magic, msg.Head.Seq, string(msg.BodyBuf))
		}
		log.Printf("id:%d, 第 %d 次连接每次发送了%d个消息\n", connectionId, i, msgCnt)
		
		cli.Disconnect()//关闭连接
	}
	log.Printf("[%d]本次连接了%d次\n", connectionId, reconnectCnt)
	exitChan <- true
	return 
}
func main() {
	serverAddr := flag.String("server", "127.0.0.1:33333", "服务端的地址")
	cliCnt := flag.Int("client", 1, "客户端个数")
	msgCnt := flag.Int("msg", 10, "每次发送消息个数")
	reconnectCnt := flag.Int("conn", 1, "每个客户端关闭后重连次数")
	flag.Parse()
	flag.Usage()
	exitChan = make(chan bool)
	start := time.Now()
	for i := 0;i < *cliCnt; i++ {
		go startAClient(*serverAddr, *reconnectCnt, *msgCnt, uint32(i))
	}
	for j := 0; j < *cliCnt; j++ {
		<- exitChan
	}
	end := time.Now()
	log.Printf("timeuse:%v\n", end.Sub(start))
}