package tcpserver

import (
	"net"
	"log"
	"encoding/binary"
	"bytes"
	"syscall"
)
type TcpServer struct {
	ExitCmd chan bool
	tcpAddr string
	clientMap map[uint64] *MessageHandler//[ip+port]->MessageHandler
}

type MessageHandler struct {
	
	conn net.Conn 
	ExitCmd chan bool
	writeChan chan []byte
}

const (
	MSG_ECHO = 1
	OTHER = 2
)

func NewTcpServer(tcpAddr string) (*TcpServer){
	return &TcpServer{
		ExitCmd : make(chan bool),
		tcpAddr : tcpAddr,
		clientMap : make(map[uint64] *MessageHandler),
	}
}

func (this *TcpServer)Start(){
	this.StartTcpServer(this.tcpAddr)
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
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept error: %s , close this socket and exit\n", err.Error())
			listener.Close()
			return err
		}
		handler := NewMessageHandler(conn)
		//this.clientList = append(this.clientMap, handler)

		go handler.WaitingForRead()
		go handler.WaitingForWrite()
	}
}

func NewMessageHandler(conn net.Conn) *MessageHandler{
	return &MessageHandler{
		conn : conn,
		ExitCmd : make(chan bool),
		writeChan : make(chan []byte, 1024*1024),
	}
}
//消息格式为
/*
 	     8        16       24       32 
|--------|--------|--------|--------|
|      bodyLen    |        magic    |
|--------|--------|--------|--------|
|                seq                |   
|--------|--------|--------|--------|
|                 body              |
|                                   |

*/
func (this *MessageHandler)WaitingForRead(){
	connFrom := this.conn.RemoteAddr().String()
	log.Printf("Connection from: %s\n", connFrom)
	var ibuf []byte = make([]byte, 1024)

	var needRead int = 1024

	var bodyLen uint16 = 0
	var endPos int = 0
	var startPos int = 0
	var magic uint16 = 0
	var seq uint32 = 0
	for {
		length, err := this.conn.Read(ibuf[endPos:])
		log.Printf("read data: %d\n", length)
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
				log.Printf("startPos:%d, endPos:%d, bodyLen:%d, magic:%d, seq:%d, needRead:%d", startPos, endPos, bodyLen, magic, seq, needRead)
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
	this.conn.Close()
	this.ExitCmd <- true
	log.Printf("Closed connection \n")

}

func (this *MessageHandler) handleMsg(headLen uint16, bodyLen uint16, buf []byte, magic uint16, seq uint32) int {
	//1 MSG_ECHO
	switch magic  {
	case MSG_ECHO:
		rspBuf := make([]byte, headLen+bodyLen)
		copy(rspBuf, buf[:headLen+bodyLen])
		this.writeChan <- rspBuf
	}
	return 0
}

func (this *MessageHandler) WaitingForWrite() {
	for {
		select {
		case buf := <-this.writeChan:
			
			_, err := this.conn.Write(buf)
			if err == nil {
				log.Printf("write to %s\n", this.conn.RemoteAddr().String())
			} else {
				//write error, donot do something. or close the socket ?
				log.Printf("write error: %s", err.Error())
			}
		case <-this.ExitCmd:
			log.Printf("recv exit cmd\n")
			goto EXITHANDLER
		}
	}
EXITHANDLER:
log.Printf("MessageHandler:WaitingForWrite exit\n")
}
	//1 MSG_ECHO
