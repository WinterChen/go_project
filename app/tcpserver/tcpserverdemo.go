package main

import (
	"log"
	"flag"
	"go_project/tcpserver"
	"os"
	"os/signal"
	"syscall"
	//"runtime/pprof"
)


func main(){
	serverAddr := flag.String("addr", "127.0.0.1:33333", "tcp listen addr")
	flag.Parse()
	//生成内存prof，方便定位内存泄漏问题
	f, err := os.OpenFile("./mem.prof", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("open file ./mem.prof file:%s\n", err.Error())
		return
	}
	killSignalChan := make(chan os.Signal, 1)
    go func(f *os.File) {
        //阻塞程序运行，直到收到终止的信号
        sig := <-killSignalChan
		log.Printf("recv signal: %v\n", sig)
		//pprof.WriteHeapProfile(f)
		f.Close()
		os.Exit(0)
        
    }(f)
    signal.Notify(killSignalChan, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSTOP)
	tcpServer := tcpserver.NewTcpServer(*serverAddr)
	tcpServer.Start()
	<- tcpServer.ExitCmd
}