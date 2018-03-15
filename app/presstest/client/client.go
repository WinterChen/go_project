package main

import(
	"log"
	"flag"
	"time"
	"sync"
	"go_project/tcpclient"
	"go_project/proto"
	"encoding/json"
)
var exitChan chan bool


//msgCnt:每次连接上server后，发送消息个数。发送完后关闭连接。
//reconnectCnt:关闭连接后重新连接次数。
//connectionId:连接的ID
func startAClient(serverAddr string, reconnectCnt int, msgCnt int, connectionId uint32, ch chan *StatInfo){
	bodyBuf := "hello world"
	head := proto.NewProtoHead(uint16(len(bodyBuf)), 1, connectionId)
	msg := proto.NewMessage(head)
	msg.WriteBody([]byte(bodyBuf))
	log.Printf("client:%d starting...\n", connectionId)
	st := &StatInfo{

	}
	for i := 0; i < reconnectCnt; i++ {
		cli := tcpclient.NewTcpClient(serverAddr, 1024)
		err := cli.Start()
		if err != nil {
			log.Println("tcpClient start fail, err:", err.Error())
			return
		} 
		start := time.Now()
		for j := 0; j < msgCnt; j++{
			//发送一个消息给server
			//time.Sleep(5*time.Minute)
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
			//log.Printf("head:magic[%d],seq[%d], body:%s\n", msg.Head.Magic, msg.Head.Seq, string(msg.BodyBuf.Bytes()))
		}
		end := time.Now()
		timeUse := end.Sub(start).Nanoseconds()
		log.Printf("id:%d, 第 %d 次连接每次发送了%d个消息,耗时：%d/%d\n", connectionId, i, msgCnt, timeUse, (timeUse)/int64(msgCnt))
		//统计信息
		st.Conns = 1
		st.Msgs = msgCnt
		st.TimeUse = (float64(timeUse)/float64(1000000))
		ch <- st
		log.Printf("st->ch, st:%v\n", st)
		cli.Disconnect()//关闭连接
	}
	log.Printf("[%d]本次连接了%d次\n", connectionId, reconnectCnt)
	time.Sleep(2*time.Minute)
	exitChan <- true
	return 
}
type StatInfo struct{
	Id int `json:"id"`
	Conns int `json:"conns"`
	Msgs int `json:"msgs"`
	TimeUse float64 `json:"timeUse"`
	lock sync.Mutex 
}

func NewStatInfo()(*StatInfo){
	stat := &StatInfo{
		Id : 0,
		Conns : 0,
		Msgs : 0,
		TimeUse : 0.0,
	}
	return stat

}

func main() {
	serverAddr := flag.String("server", "127.0.0.1:33333", "服务端的地址")
	statAddr := flag.String("stat", "127.0.0.1:44444", "统计模块的地址")
	cliCnt := flag.Int("client", 1, "客户端个数")
	msgCnt := flag.Int("msg", 10, "每次发送消息个数")
	reconnectCnt := flag.Int("conn", 1, "每个客户端关闭后重连次数")
	id := flag.Int("id", 0, "id")
	flag.Parse()
	flag.Usage()
	exitChan = make(chan bool)
	start := time.Now()
	//统计
	statChan := make (chan *StatInfo, (*cliCnt)*2)
	go SendStatInfo(statChan, *statAddr, *id)

	for i := 0;i < *cliCnt; i++ {
		go startAClient(*serverAddr, *reconnectCnt, *msgCnt, uint32(i), statChan)
	}
	
	for j := 0; j < *cliCnt; j++ {
		<- exitChan
	}
	end := time.Now()
	log.Printf("timeuse:%v\n", end.Sub(start))
}

func SendStatInfo(ch chan *StatInfo, statAddr string, id int){
	cli := tcpclient.NewTcpClient(statAddr, 1024)
	err := cli.Start()
	if err != nil {
		log.Println("tcpClient start fail, err:", err.Error())
		return
	} 
	
	
	var seq uint32 = 0
	for {
		st := <- ch

		log.Printf("recv st: %v\n", st)
		head := proto.NewProtoHead(uint16(0), 2, seq)
		msg := proto.NewMessage(head)
		st.Id = id
		seq++
		buf, err := json.Marshal(st)
		if err != nil {
			log.Printf("json marshal err: %s\n", err.Error())
			continue
		}
		head.BodyLen = uint16 (len(buf))
		msg.WriteBody(buf)
		err = cli.Write(msg)
		if err != nil {
			log.Printf("write err: %s\n", err.Error())
		}
		//log.Printf("write to stat server succ, body:%s\n", string(buf))

	}
}