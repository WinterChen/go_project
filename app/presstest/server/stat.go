package main

import (
	"flag"
	"go_project/tcpserver"
	"go_project/proto"
	"fmt"
	"time"
	"encoding/json"
	"sync"
)

type StatInfo struct{
	Id int `json:"id"`
	Conns int `json:"conns"`
	Msgs int `json:"msgs"`
	TimeUse float64 `json:"timeUse"`
	lock sync.Mutex 
}

func NewStatInfo()(*StatInfo){
	stat := &StatInfo{
		Id : 0,
		Conns : 0,
		Msgs : 0,
		TimeUse : 0,
	}
	return stat

}

type StatHandler struct{
	Stats map[int]*StatInfo
	lock sync.Mutex 


}
func NewStatHandler()(*StatHandler){
	stat := &StatHandler{
		Stats : make(map[int] *StatInfo),

	}
	return stat

}
//每隔1分钟打印统计情况
func (this *StatHandler) MainLoop(){
	
	for {
		select {
		case <-time.Tick(time.Minute * 1):
			total := NewStatInfo()
			this.lock.Lock()//加锁
			for _, stat :=  range this.Stats {
				stat.lock.Lock()
				total.Conns += stat.Conns
				stat.Conns = 0
				total.Msgs += stat.Msgs
				stat.Msgs = 0
				total.TimeUse += stat.TimeUse
				stat.TimeUse = 0
				stat.lock.Unlock()
			}
			fmt.Println("---------------------------------------")
			fmt.Printf("total conns: %d\n", total.Conns)
			fmt.Printf("total msgs: %d\n", total.Msgs)
			fmt.Printf("total timeuse: %f\n", total.TimeUse)
			
			avgTime := total.TimeUse/float64(total.Msgs)
			
			fmt.Printf("avg timeuse per msg(ms): %f\n", avgTime)
			this.lock.Unlock()//解锁
		}
	}
	
	fmt.Println("mainloop end")
}
func (this *StatHandler) ProcessMsg(msg *proto.Message) (*proto.Message){
	statinfo := NewStatInfo()
	err := json.Unmarshal(msg.BodyBuf.Bytes(), statinfo)
	if err != nil {
		fmt.Printf("StatHandler ProcessMsg parse error:%s. body:%s\n", err.Error(), string(msg.BodyBuf.Bytes()))
		return nil
	}
	//fmt.Printf("msg:%v", statinfo)
	st, ok := this.Stats[statinfo.Id] 
	if !ok {
		this.lock.Lock()
		this.Stats[statinfo.Id] = statinfo
		st = statinfo
		this.lock.Unlock()
	} else {
		st.lock.Lock()
		st.Conns += statinfo.Conns
		st.Msgs += statinfo.Msgs
		st.TimeUse += statinfo.TimeUse
		st.lock.Unlock()
	}
	
	return nil
}
//统计模块
func main(){
	serverAddr := flag.String("addr", "127.0.0.1:44444", "tcp listen addr")
	
	flag.Parse()
	statHandler := NewStatHandler()
	go statHandler.MainLoop()
	tcpServer := tcpserver.NewTcpServer(*serverAddr)
	tcpServer.RegisterMessageProcessor(proto.MSG_STAT, statHandler)
	tcpServer.Start()
	<- tcpServer.ExitCmd
	
}