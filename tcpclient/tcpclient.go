package tcpclient

import(
	"log"
	"net"
	//"encoding/binary"
	"bytes"
	"syscall"
	"errors"
)

type TcpClient struct {
	serverAddr string
	conn net.Conn
	readBufLen int
	headLen int
	outMessageChan chan *Message
	inMessageChan chan *Message
	writeExit chan bool
	//readExit chan bool
}


func NewTcpClient(serverAddr string, readBufLen int) (*TcpClient){
	return &TcpClient{
		serverAddr : serverAddr,
		readBufLen : readBufLen,
		headLen : HEADLEN,
		outMessageChan : make(chan *Message, 1024),
		inMessageChan : make(chan *Message, 1024),
		writeExit : make(chan bool),
		//readExit : make(chan bool),
	}
}


func (this *TcpClient)Start() error{
	conn, err := net.Dial("tcp", this.serverAddr)
	if err != nil {
		log.Printf("connect to %s fail\n", err.Error())
		return err
	}
	log.Println("connect to server succ...")
	this.conn = conn
	go this.WaitingForWrite()
	go this.WaitingForRead()
	return nil
}

func (this *TcpClient)WaitingForRead(){
	connFrom := this.conn.RemoteAddr().String()
	log.Printf("remote: %s\n", connFrom)
	var buf []byte = make([]byte, this.readBufLen)

	var needRead int = 0
	var head *ProtoHead
	var endPos int = 0
	var startPos int = 0
	for {
		length, err := this.conn.Read(buf[endPos:])
		log.Printf("read data: %d\n", length)
		switch err {
		case nil:
			endPos += length
			//读到数据，有可能含有多个包的数据，所以需要循环处理
			for {
				//读到的数据小于协议头长度，退出继续等待读
				if endPos-startPos < this.headLen {
					break
				}
				if needRead == 0 {
					head = ParseHead(buf[startPos:])
				}
				needRead = int(head.BodyLen) - (endPos - startPos - this.headLen)
				log.Printf("startPos:%d, endPos:%d, bodyLen:%d, magic:%d, seq:%d, needRead:%d", startPos, endPos, head.BodyLen, head.Magic, head.Seq, needRead)
				if needRead > 0 {
					//没有读够bodyLen个长度的数据，退出继续读
					break
				} else {
					//这里多余一次make和copy的调用
					bodyDate := make([]byte, head.BodyLen)
					copy(bodyDate, buf[startPos+this.headLen : ])
					err = this.handleMsg(bodyDate, head)
					if err != nil {
						log.Printf("handle msg error, close the connection\n")
						goto DISCONNECT
					}
					startPos += int(head.BodyLen) + this.headLen
					needRead = 0
					if startPos == endPos {
						startPos = 0
						endPos = 0
					}
				}
			}
			//循环处理完后，如果还有数据没有处理完，将其移动到buf的头部
			if startPos < endPos && startPos > 0 {
				reader := bytes.NewReader(buf)
				reader.ReadAt(buf, int64(endPos-startPos))
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
	//write goroutine退出
	this.Disconnect()
	this.writeExit <- true
	log.Printf("WaitingForRead exit \n")

}

func (this *TcpClient) handleMsg(buf []byte, head *ProtoHead)(error){
	msg := &Message{
		Head : head,
		BodyBuf : buf,
	}
	select {
	case this.inMessageChan <- msg:
	default:
		log.Printf("inMessageChan is full!!!\n")
		return errors.New("inMessageChan is full")
	}

	return nil
}


func (this *TcpClient) WaitingForWrite(){
	for {
		select {
		case msg, ok := <- this.outMessageChan:
			if !ok {
				log.Printf("message chan maybe closed by others")
				return
			}
			_, err := this.conn.Write(msg.Encoding())
			if err == nil {
				log.Printf("write to %s\n", this.conn.RemoteAddr().String())
			} else {
				//write错误，关闭socket，并通知read goroutine退出
				log.Printf("write error: %s\n", err.Error())
				goto DISCONNECT
			}
		case <- this.writeExit:
			log.Printf("WaitingForWrite recv exit\n")
			return
		}
	}
DISCONNECT:
	this.Disconnect()
	//this.readExit <- true
	log.Printf("WaitingForWrite exit \n")
}


func (this *TcpClient) Disconnect(){
	this.conn.Close()
	close(this.inMessageChan)
	close(this.outMessageChan)

}

func (this *TcpClient) Write(msg *Message)(error){
	select {
	case this.outMessageChan <- msg:
	default:
		log.Printf("messageChan is full !!!\n")
		return errors.New("messageChan is full")
	}
	return nil	
}

func (this *TcpClient) GetMessage()(*Message){
	for {
		select {
		case msg, ok := <- this.inMessageChan:
			if !ok {
				log.Printf("message chan maybe closed by others")
				return nil
			}
			return msg
		}
	}
}
