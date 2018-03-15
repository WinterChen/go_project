package proto

import (
	"log"
	//"log"
	"encoding/binary"
	"bytes"
)

const (
	MSG_ECHO = 1
	MSG_STAT = 2
	OTHER = 3
)
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
const HEADLEN int = 8
type ProtoHead struct {
	BodyLen uint16
	Magic uint16
	Seq uint32
} 

type Message struct {
	Head *ProtoHead
	BodyBuf *bytes.Buffer//body的buffer
	EncodingBuf *bytes.Buffer//用于编码的buffer
}


func ParseHead(buf []byte)(*ProtoHead){
	bodyLen := binary.BigEndian.Uint16(buf[:2])
	magic := binary.BigEndian.Uint16(buf[2 : 4])
	seq := binary.BigEndian.Uint32(buf[4 : 8])
	//log.Printf("bodyLen:%d, magic:%d, seq:%d", bodyLen, magic, seq)
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

func NewProtoHead(bodyLen uint16, magic uint16, seq uint32) *ProtoHead{
	return &ProtoHead{
		BodyLen : bodyLen,
		Magic : magic,
		Seq : seq,
	}
}

func NewMessage(head *ProtoHead) *Message{
	return &Message{
		Head : head,
		BodyBuf : new(bytes.Buffer),
		EncodingBuf : new(bytes.Buffer),
	}
}

func (this *Message) WriteBody(buf []byte){
	_, err := this.BodyBuf.Write(buf)
	if err != nil {
		log.Printf("write err:%s\n", err.Error())
	}
}

func (this *Message) Encoding()([]byte){
	this.EncodingBuf.Reset()//encoding的时候先reset
	binary.Write(this.EncodingBuf, binary.BigEndian, this.Head)
	this.EncodingBuf.Write(this.BodyBuf.Bytes())
	return this.EncodingBuf.Bytes()
	
}

func (this *Message) Reset(){
	this.Head.Reset()
	this.BodyBuf.Reset()
	this.EncodingBuf.Reset()
}




