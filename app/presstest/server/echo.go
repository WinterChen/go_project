package main

import (
	"log"
	"flag"
	"go_project/tcpserver"
	"os"
	"os/signal"
	"syscall"
	"runtime/pprof"
	_ "net/http/pprof"
	"net/http"
	"runtime"
	"fmt"
	"time"
	"go_project/proto"
)

type EchoServer struct {

}
func (this *EchoServer) ProcessMsg(msg *proto.Message) (*proto.Message){
	//fmt.Println("EchoServer, processmsg")
	return msg
}
func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	p := pprof.Lookup("goroutine")
	p.WriteTo(w, 1)
}

func WaitingForSignal(){
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)
	signal.Notify(signalChan, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSTOP)
	<-signalChan
	signal.Stop(signalChan)
	saveHeapProfile()
	os.Exit(0)
}
//生成内存prof，方便定位内存泄漏问题
func saveHeapProfile() {
		runtime.GC()//先GC
		f, err := os.Create(fmt.Sprintf("./heap_%s.prof", time.Now().Format("2006_01_02_03_04_05")))
		if err != nil {
			return
		}
		pprof.Lookup("heap").WriteTo(f, 1)
		f.Close()
	} 

func main(){
	serverAddr := flag.String("addr", "127.0.0.1:33333", "tcp listen addr")
	pprofAddr := flag.String("pprofAddr", "localhost:6060", "pprof http listen addr")
	flag.Parse()
	go func(){
		log.Println(http.ListenAndServe(*pprofAddr, nil)) 
	}()
	go WaitingForSignal()
	echoServer := &EchoServer{

	}
	/*
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
    signal.Notify(killSignalChan, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSTOP)*/
	tcpServer := tcpserver.NewTcpServer(*serverAddr)
	tcpServer.RegisterMessageProcessor(proto.MSG_ECHO, echoServer)
	tcpServer.Start()
	<- tcpServer.ExitCmd
}