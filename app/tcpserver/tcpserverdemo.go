package main

import (
	//"log"
	"flag"
	"go_project/tcpserver"
)


func main(){
	serverAddr := flag.String("addr", "127.0.0.1:33333", "tcp listen addr")
	flag.Parse()
	tcpServer := tcpserver.NewTcpServer(*serverAddr)
	tcpServer.Start()
	<- tcpServer.ExitCmd
}