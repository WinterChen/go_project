package tcpclient

import (
	"log"
	"encoding/binary"
	"bytes"
)

const HEADLEN int = 8
type ProtoHead struct {
	BodyLen uint16
	Magic uint16
	Seq uint32
} 

type Message struct {
	Head *ProtoHead
	BodyBuf []byte
}


func ParseHead(buf []byte)(*ProtoHead){
	bodyLen := binary.BigEndian.Uint16(buf[:2])
	magic := binary.BigEndian.Uint16(buf[2 : 4])
	seq := binary.BigEndian.Uint32(buf[4 : 8])
	log.Printf("bodyLen:%d, magic:%d, seq:%d", bodyLen, magic, seq)
	return &ProtoHead{
		BodyLen : bodyLen,
		Magic : magic,
		Seq : seq,
	}
}
func (this *ProtoHead)Reset(){
	this.BodyLen = 0
	this.Magic = 0
	this.Seq = 0
}

func (this *Message) Encoding()([]byte){
	msgBuf := new(bytes.Buffer)
	binary.Write(msgBuf, binary.BigEndian, this.Head)
	msgBuf.Write(this.BodyBuf)
	return msgBuf.Bytes()
	
}




