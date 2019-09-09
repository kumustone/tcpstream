package tcpstream

import (
	"fmt"
	"testing"
	"time"
)

type handler struct {
}

func (h *handler) OnData(conn *TcpStream, msg *Message) error {
	_ = conn.Write(&Message{
		//因为是多路复用的，返回的seq必须要与client请求的seq对应
		Header: ProtoHeader{
			Seq: msg.Header.Seq,
		},
		Body: []byte("I am server"),
	})
	return nil
}
func (h *handler) OnConn(conn *TcpStream)    {}
func (h *handler) OnDisConn(conn *TcpStream) {}

var s = NewTCPServer("127.0.0.1:7001", &handler{})

func TestSyncClient_Call(t *testing.T) {

	s.Serve()

	//waiting for s accept
	time.Sleep(time.Millisecond * 10)
	var c = NewSyncClient("127.0.0.1:7001")

	m := Message{
		Body: []byte("Hello, I AM CLIENT"),
	}

	if resp, err := c.Call(&m, time.Duration(2*time.Second)); err != nil {
		fmt.Println("call fail ", err.Error())
	} else {
		fmt.Println("resp : ", string(resp.Body))
	}
}
