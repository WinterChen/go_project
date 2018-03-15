package tcpserver

import (
	"net"
	"log"
	"encoding/binary"
	"bytes"
	"syscall"
	//"net/http"
	//"runtime/pprof"
	"container/list"
	"go_project/proto"
	//"unsafe"
)


type TcpServer struct {
	ExitCmd chan bool
	tcpAddr string
	handlers *list.List
	handlerCnt int
	reclaimer chan *MessageHandler
	//clientMap map[uint64] *MessageHandler//[ip+port]->MessageHandler
	//businessFuncMap map[int] func(args... interface{})
	msgProcessorMap map[int] MessageProcessor
}

type MessageHandler struct {
	
	conn net.Conn 
	ExitCmd chan bool
	writeChan chan []byte
	id int
	freeMessages chan *proto.Message //可用的message
	respMessages chan *proto.Message //响应的message
	reclaimer chan *MessageHandler //回收的MessageHandler
	owner *TcpServer

}



func NewTcpServer(tcpAddr string) (*TcpServer){
	return &TcpServer{
		ExitCmd : make(chan bool),
		tcpAddr : tcpAddr,
		handlers : list.New(),
		handlerCnt : 1000,
		reclaimer : make(chan *MessageHandler),
		//clientMap : make(map[uint64] *MessageHandler),
		//businessFuncMap : make(map[int] func(args... interface{})),
		msgProcessorMap : make(map[int] MessageProcessor),
	}
}

func (this *TcpServer)Start(){
	//提前生成对象池
	log.Printf("提前生成对象池, 个数:%d\n", this.handlerCnt)
	for i:=0;i<this.handlerCnt;i++ {
		h := NewMessageHandler(nil, i, this.reclaimer, this)
		this.handlers.PushBack(h)
		//log.Printf("i:%d, 每个大小：%d\n", i, unsafe.Sizeof(h))
	}
	//log.Println("MessageHandler cnt: ", this.handlerCnt)
	this.StartTcpServer(this.tcpAddr)
}

func (this *TcpServer) RegisterMessageProcessor(magic int, f MessageProcessor){
	this.msgProcessorMap[magic] = f
}

func (this *TcpServer) GetMessageProcessor(magic int) (MessageProcessor){
	f, ok := this.msgProcessorMap[magic]
	if !ok {
		return nil
	}
	return f
}


func (this *TcpServer) StartTcpServer(hostAndPort string) error {
	defer func(){
		this.ExitCmd <- true
	}()
	serverAddr, err := net.ResolveTCPAddr("tcp", hostAndPort)
	if err != nil {
		log.Printf("Resolving %s failed: %s\n", hostAndPort, err.Error())
		return err
	}

	listener, err := net.ListenTCP("tcp", serverAddr)
	if err != nil {
		log.Printf("listen %s failed: %s", hostAndPort, err.Error())
		return err
	}
	log.Printf("start tcp server %s\n", hostAndPort)
	
	var handler *MessageHandler
	var ok bool
	for {
		
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept error: %s , close this socket and exit\n", err.Error())
			listener.Close()
			return err
		}
			
		if this.handlers.Len() >= 1{
			e := this.handlers.Front()//从handler池中拿一个
			handler, ok = this.handlers.Remove(e).(*MessageHandler)
			if ok {
				handler.conn = conn
			} else {//类型断言失败？不可能吧
				log.Println("the element in handlers list is not MessageHandler type ???")
				handler = NewMessageHandler(conn, this.handlerCnt,this.reclaimer, this)
				this.handlerCnt++ 
			}
		} else {//handler池已经没有对象
			log.Println("there is not enought element in obj list!!!")
			handler = NewMessageHandler(conn, this.handlerCnt,this.reclaimer, this)
			this.handlerCnt++
		}
		//每次从回收池中获取一个对象，丢到列表中
		select {
		case h := <- this.reclaimer:
			this.handlers.PushBack(h)
		default:
			
		}
		
		//this.clientList = append(this.clientMap, handler)

		go handler.WaitingForRead()
		go handler.WaitingForWrite()
	
	}
}
/*
func (this *TcpServer) RegisterBusinessFunc(business int, f func(args... interface{})){
	this.businessFuncMap[business] = f
}*/

func NewMessageHandler(conn net.Conn, id int, reclaimer chan *MessageHandler, server *TcpServer) *MessageHandler{
	return &MessageHandler{
		conn : conn,
		ExitCmd : make(chan bool),
		writeChan : make(chan []byte, 1024),
		id : id,
		freeMessages : make(chan *proto.Message, 100*2),//容量是respMessages两倍，可以根据实际调整
		respMessages : make(chan *proto.Message, 100),
		reclaimer : reclaimer,
		owner : server,
	}
}

func (this *MessageHandler)WaitingForRead(){
	connFrom := this.conn.RemoteAddr().String()
	log.Printf("Connection from: %s\n", connFrom)
	var ibuf []byte = make([]byte, 10240)

	var needRead int = 1024

	var bodyLen uint16 = 0
	var endPos int = 0
	var startPos int = 0
	var magic uint16 = 0
	var seq uint32 = 0
	for {
		length, err := this.conn.Read(ibuf[endPos:])
		//log.Printf("read data: %d\n", length)
		switch err {
		case nil:
			endPos += length
			for {
				if endPos-startPos < 8 {
					break
				}
				if bodyLen == 0 {
					bodyLen = binary.BigEndian.Uint16(ibuf[startPos : startPos+2])
					magic = binary.BigEndian.Uint16(ibuf[startPos+2 : startPos+4])
					seq = binary.BigEndian.Uint32(ibuf[startPos+4 : startPos+8])
				}
				needRead = int(bodyLen) - (endPos - startPos - 8)
				//log.Printf("startPos:%d, endPos:%d, bodyLen:%d, magic:%d, seq:%d, needRead:%d", startPos, endPos, bodyLen, magic, seq, needRead)
				if needRead > 0 {
					break
				} else {
					res := this.handleMsg(8, bodyLen, ibuf[startPos:], magic, seq)
					if res == -1 {
						log.Printf("handle msg error, close the connection\n")
						goto DISCONNECT
					}
					startPos += int(bodyLen) + 8
					bodyLen = 0
					needRead = 0
					if startPos == endPos {
						startPos = 0
						endPos = 0
					}
				}
			}
			if startPos < endPos && startPos > 0 {
				reader := bytes.NewReader(ibuf)
				reader.ReadAt(ibuf, int64(endPos-startPos))
				startPos = 0
				endPos -= startPos
			}
		case syscall.Errno(0xb): // try again
			log.Printf("read need try again\n")
			continue
		default:
			log.Printf("read error %s\n", err.Error())
			goto DISCONNECT
		}

	}
DISCONNECT:
	
	this.ExitCmd <-true
	//log.Printf("MessageHandler: %d WaitingForRead exit  \n", this.id)

}

func (this *MessageHandler) handleMsg(headLen uint16, bodyLen uint16, buf []byte, magic uint16, seq uint32) int {
	//生成Message
	var msg *proto.Message
	var ok bool
	
	select {
	case msg, ok = <- this.freeMessages:
		if ok && msg != nil {
			msg.Head.BodyLen = bodyLen
			msg.Head.Magic = magic
			msg.Head.Seq = seq
			msg.BodyBuf.Write(buf[headLen:headLen+bodyLen])
		}
	default:
		head := proto.NewProtoHead(bodyLen, magic, seq)
		msg = proto.NewMessage(head)
		msg.BodyBuf.Write(buf[headLen:headLen+bodyLen])
	}
	//处理message
	messageHandler := this.owner.GetMessageProcessor(int(magic))
	if messageHandler == nil {
		log.Printf("not found message handler for magic: %d \n", magic)
		msg.Reset()
		this.freeMessages <- msg
		return 0
	}
	rspMsg := messageHandler.ProcessMsg(msg)
	if rspMsg == nil {//回收
		msg.Reset()
		this.freeMessages <- msg
	} else {
		this.respMessages <- rspMsg
	}
	
	return 0
}

func (this *MessageHandler) ProcessEchoMsg(headLen uint16, bodyLen uint16, buf []byte, magic uint16, seq uint32){
	var msg *proto.Message
	var ok bool
	
	select {
	case msg, ok = <- this.freeMessages:
		if ok && msg != nil {
			msg.Head.BodyLen = bodyLen
			msg.Head.Magic = magic
			msg.Head.Seq = seq
			msg.BodyBuf.Write(buf[headLen:headLen+bodyLen])
		}
	default:
		head := proto.NewProtoHead(bodyLen, magic, seq)
		msg = proto.NewMessage(head)
		msg.BodyBuf.Write(buf[headLen:headLen+bodyLen])
	}
	//处理message

	this.respMessages <- msg
}


func (this *MessageHandler) WaitingForWrite() {
	//var outbuf []byte = make([]byte, 10240)	
	for {
		select {
		case msg, ok := <-this.respMessages:
			if !ok {
				goto EXITHANDLER
			}
			_, err := this.conn.Write(msg.Encoding())
			if err == nil {
				//log.Printf("write to %s\n", this.conn.RemoteAddr().String())
				//回收
				msg.Reset()
				this.freeMessages <- msg
			} else {
				//write error, donot do something. or close the socket ?
				log.Printf("MessageHandler: %d, write error: %s", this.id, err.Error())
				//写失败也回收
				msg.Reset()
				this.freeMessages <- msg
				goto EXITHANDLER
			}
		case <- this.ExitCmd:
			//退出是消费完所有未回复的message
			//log.Printf("MessageHandler: %d, WaitingForWrite recv exit cmd, len:%d", this.id, len(this.respMessages))
			l := len(this.respMessages)
			for i:=0;i<l;i++ {
				msg := <-this.respMessages
				msg.Reset()
				this.freeMessages <- msg
			}
			
			goto EXITHANDLER
		}
	
	}
EXITHANDLER:
	this.conn.Close()
	this.reclaimer <- this
	//log.Printf("MessageHandler: %d WaitingForWrite exit\n", this.id)
}
	//1 MSG_ECHO
