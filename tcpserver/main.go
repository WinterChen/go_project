package main

import (
	//"log"
	"flag"
)


func main(){
	serverAddr := flag.String("addr", "127.0.0.1:33333", "tcp listen addr")
	flag.Parse()
	tcpServer := NewTcpServer(*serverAddr)
	tcpServer.Start()
	<- tcpServer.exitCmd
}