package tcpserver

import(
	"testing"
)

func TestNewTcpServer(t *testing.T){
	tcpServer := NewTcpServer("127.0.0.1:12345")
	err := tcpServer.StartTcpServer(tcpServer.tcpAddr) 
	if err != nil {
		t.Errorf("start tcp server %s error: %s", tcpServer.tcpAddr, err.Error())
	}
	
}