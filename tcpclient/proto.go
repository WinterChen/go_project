package main

import (
	"encoding/binary"
	"bytes"
)

const HEADLEN int = 8
type ProtoHead struct {
	bodyLen uint16
	magic uint16
	seq uint32
} 

type Message struct {
	head *ProtoHead
	bodyBuf []byte
}


func ParseHead(buf []byte)(*ProtoHead){
	bodyLen := binary.BigEndian.Uint16(buf[:2])
	magic := binary.BigEndian.Uint16(buf[2 : 4])
	seq := binary.BigEndian.Uint32(buf[4 : 8])
	return &ProtoHead{
		bodyLen : bodyLen,
		magic : magic,
		seq : seq,
	}
}
func (this *ProtoHead)Reset(){
	this.bodyLen = 0
	this.magic = 0
	this.seq = 0
}

func (this *Message) Encoding()([]byte){
	msgBuf := new(bytes.Buffer)
	binary.Write(msgBuf, binary.BigEndian, this.head)
	msgBuf.Write(this.bodyBuf)
	return msgBuf.Bytes()
	
}




