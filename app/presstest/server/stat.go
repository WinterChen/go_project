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
	AvgTime int `json:"avgTime"`
	lock sync.Mutex 
}

func NewStatInfo()(*StatInfo){
	stat := &StatInfo{
		Id : 0,
		Conns : 0,
		Msgs : 0,
		AvgTime : 0,
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
	cnt := 0
	select {
	case <-time.After(time.Minute * 1):
		total := NewStatInfo()
		this.lock.Lock()//加锁
		for _, stat :=  range this.Stats {
			stat.lock.Lock()
			total.Conns += stat.Conns
			stat.Conns = 0
			total.Msgs += stat.Msgs
			stat.Msgs = 0
			total.AvgTime += stat.AvgTime
			if total.AvgTime != 0 {
				cnt ++
			}
			stat.AvgTime = 0
			stat.lock.Unlock()
		}
		fmt.Println("---------------------------------------")
		fmt.Printf("total conns: %d\n", total.Conns)
		fmt.Printf("total msgs: %d\n", total.Msgs)
		avgTime := 0
		if cnt != 0 {
			avgTime = total.AvgTime/cnt
		}
		fmt.Printf("avg timeuse per msg(ms): %d\n", avgTime)
		this.lock.Unlock()//解锁
	}
}
func (this *StatHandler) ProcessMsg(msg *proto.Message) (*proto.Message){
	statinfo := NewStatInfo()
	err := json.Unmarshal(msg.BodyBuf.Bytes(), statinfo)
	if err != nil {
		fmt.Printf("StatHandler ProcessMsg parse error:%s\n", err.Error())
		return nil
	}
	st, ok := this.Stats[statinfo.Id] 
	if !ok {
		this.lock.Lock()
		this.Stats[statinfo.Id] = statinfo
		this.lock.Unlock()
	}
	st.lock.Lock()
	st.Conns += statinfo.Conns
	st.Msgs += statinfo.Msgs
	st.AvgTime = (st.AvgTime + statinfo.AvgTime)/2
	st.lock.Unlock()
	return nil
}
//统计模块
func main(){
	serverAddr := flag.String("addr", "127.0.0.1:44444", "tcp listen addr")
	
	flag.Parse()
	statHandler := NewStatHandler()
	statHandler.MainLoop()
	tcpServer := tcpserver.NewTcpServer(*serverAddr)
	tcpServer.RegisterMessageProcessor(proto.MSG_STAT, statHandler)
	tcpServer.Start()
	<- tcpServer.ExitCmd
	
}